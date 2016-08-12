package simple

import (
	"encoding/json"
	"time"

	"github.com/evo-cloud/cloudrt/jobs"
)

// Strategy implements jobs.Strategy
type Strategy struct {
	Store jobs.Store
}

// JobDoc is the persisted document of job
type JobDoc struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	TaskID    string    `json:"task-id"`
	CreatedAt time.Time `json:"created-at"`
	UpdatedAt time.Time `json:"updated-at"`
}

// NewJobDoc creates a JobDoc from a job
func NewJobDoc(job *jobs.Job) *JobDoc {
	return &JobDoc{
		ID:        job.ID,
		Name:      job.Name,
		TaskID:    job.Task.ID,
		CreatedAt: job.CreatedAt,
		UpdatedAt: job.UpdatedAt,
	}
}

// ToJob converts JobDoc to Job
func (d *JobDoc) ToJob() *jobs.Job {
	return &jobs.Job{
		ID:        d.ID,
		Name:      d.Name,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

// TaskDoc is the persisted document of task
type TaskDoc struct {
	ID         string           `json:"id"`          // globally unique task id
	ParentID   string           `json:"parent-id"`   // parent task id
	JobID      string           `json:"job-id"`      // job id
	Name       string           `json:"name"`        // task name
	Params     json.RawMessage  `json:"params"`      // encoded parameters
	State      jobs.TaskState   `json:"state"`       // current state
	Result     jobs.TaskResult  `json:"result"`      // result when task completes
	Revert     bool             `json:"revert"`      // in rollback direction
	Retries    uint             `json:"retries"`     // current retry number
	MaxRetries uint             `json:"max-retries"` // max count of retries
	Stage      string           `json:"stage"`       // current stage
	ResumeTo   string           `json:"resume-to"`   // next stage resume to
	Data       json.RawMessage  `json:"data"`        // task specific data
	Output     json.RawMessage  `json:"output"`      // output when completed
	Errors     []jobs.TaskError `json:"errors"`      // errors happened
	CreatedAt  time.Time        `json:"created-at"`  // task creation time
	UpdatedAt  time.Time        `json:"updated-at"`  // last modification time
	SubTaskIDs []string         `json:"subtask-ids"` // subtask ID list
}

// NewTaskDoc creates a TaskDoc from a Task
func NewTaskDoc(task *jobs.Task) *TaskDoc {
	return &TaskDoc{
		ID:         task.ID,
		ParentID:   task.ParentID,
		JobID:      task.JobID,
		Name:       task.Name,
		Params:     json.RawMessage(task.Params),
		State:      task.State,
		Result:     task.Result,
		Revert:     task.Revert,
		Retries:    task.Retries,
		MaxRetries: task.MaxRetries,
		Stage:      task.Stage,
		ResumeTo:   task.ResumeTo,
		Data:       json.RawMessage(task.Data),
		Output:     json.RawMessage(task.Output),
		Errors:     task.Errors,
		CreatedAt:  task.CreatedAt,
		UpdatedAt:  task.UpdatedAt,
		SubTaskIDs: task.SubTaskIDs,
	}
}

// ToTask converts TaskDoc to Task
func (d *TaskDoc) ToTask() *jobs.Task {
	return &jobs.Task{
		ID:         d.ID,
		ParentID:   d.ParentID,
		JobID:      d.JobID,
		Name:       d.Name,
		Params:     []byte(d.Params),
		State:      d.State,
		Result:     d.Result,
		Revert:     d.Revert,
		Retries:    d.Retries,
		MaxRetries: d.MaxRetries,
		Stage:      d.Stage,
		ResumeTo:   d.ResumeTo,
		Data:       []byte(d.Data),
		Output:     []byte(d.Output),
		Errors:     d.Errors,
		CreatedAt:  d.CreatedAt,
		UpdatedAt:  d.UpdatedAt,
		SubTaskIDs: d.SubTaskIDs,
	}
}

// Names
const (
	JobsBucket      = "jobs"
	TasksBucket     = "tasks"
	TaskStatsBucket = "task-stats"
	CancelList      = "job-cancellation"
	PendingList     = "task-pending"
	WaitingList     = "task-waiting"
)

// SubmitJob implements Strategy
func (s *Strategy) SubmitJob(job *jobs.Job) (err error) {
	doc := NewJobDoc(job)
	err = s.Store.Bucket(JobsBucket).Put(doc.ID, doc, jobs.Infinite)
	if err != nil {
		return
	}
	taskDoc := NewTaskDoc(job.Task)
	taskDoc.State = jobs.TaskPending
	return s.saveTask(taskDoc, nil)
}

// CancelJob implements Strategy
func (s *Strategy) CancelJob(id string) error {
	return s.Store.OrderedList(CancelList).Set(id, true)
}

// IsJobCanceling implements Strategy
func (s *Strategy) IsJobCanceling(id string) (bool, error) {
	return s.Store.OrderedList(CancelList).Has(id)
}

// QueryJob implements Strategy
func (s *Strategy) QueryJob(id string) (*jobs.Job, error) {
	doc, err := s.queryJobDoc(id)
	if err != nil {
		return nil, err
	}
	job := doc.ToJob()
	job.Task, err = s.QueryTask(doc.TaskID)
	return job, err
}

// QueryTask implements Strategy
func (s *Strategy) QueryTask(id string) (*jobs.Task, error) {
	doc, err := s.queryTaskDoc(id)
	if err != nil {
		return nil, err
	}
	stats, err := s.queryTaskStats(id)
	if err != nil {
		return nil, err
	}
	task := doc.ToTask()
	if stats != nil {
		task.Stats = stats
	}
	return task, nil
}

// NewWorker creates a worker strategy
func (s *Strategy) NewWorker(id string) jobs.WorkerStrategy {
	return &WorkerStrategy{WorkerID: id, Strategy: s}
}

// HouseKeep runs house keeping logic
func (s *Strategy) HouseKeep(id string, logic jobs.HouseKeepLogic) error {
	// TODO
	return nil
}

func (s *Strategy) queryJobDoc(id string) (*JobDoc, error) {
	val, err := s.Store.Bucket(JobsBucket).Get(id)
	if err != nil || val == nil {
		return nil, err
	}
	doc := &JobDoc{}
	return doc, val.Unmarshal(doc)
}

func (s *Strategy) queryTaskDoc(id string) (*TaskDoc, error) {
	val, err := s.Store.Bucket(TasksBucket).Get(id)
	if err != nil || val == nil {
		return nil, err
	}
	doc := &TaskDoc{}
	return doc, val.Unmarshal(doc)
}

func (s *Strategy) queryTaskStats(id string) (*jobs.TaskStats, error) {
	val, err := s.Store.Bucket(TaskStatsBucket).Get(id)
	if err != nil || val == nil {
		return nil, err
	}
	stats := &jobs.TaskStats{}
	return stats, val.Unmarshal(stats)
}

func (s *Strategy) cancelRequested(id string) (bool, error) {
	return s.Store.OrderedList(CancelList).Has(id)
}

func (s *Strategy) saveTask(doc *TaskDoc, stats *jobs.TaskStats) (err error) {
	doc.UpdatedAt = time.Now()
	if err = s.Store.Bucket(TasksBucket).Put(doc.ID, doc, jobs.Infinite); err != nil {
		return
	}
	s.Store.OrderedList(PendingList).Set(doc.ID, doc.State == jobs.TaskPending)
	s.Store.OrderedList(WaitingList).Set(doc.ID, doc.State == jobs.TaskWaiting)
	if stats != nil {
		err = s.Store.Bucket(TaskStatsBucket).Put(doc.ID, stats, jobs.Infinite)
	}
	return
}

// WorkerStrategy implements jobs.WorkerStrategy
type WorkerStrategy struct {
	WorkerID string
	Strategy *Strategy
}

// FetchTask implements WorkerStrategy
func (w *WorkerStrategy) FetchTask() (jobs.TaskHandle, error) {
	e := w.Strategy.Store.OrderedList(PendingList).Enumerate(jobs.EnumOptions{PageSize: 10})
	for {
		tasks, err := e.Next()
		if err != nil {
			return nil, err
		}
		if tasks == nil {
			break
		}
		for _, val := range tasks {
			var id string
			if err := val.Unmarshal(&id); err != nil || id == "" {
				continue
			}
			acq, err := w.Strategy.Store.Acquire("task:"+id, w.WorkerID)
			if err != nil || !acq.Acquired() {
				continue
			}
			task, err := w.Strategy.QueryTask(id)
			if err != nil || task == nil {
				acq.Release()
				continue
			}
			return &TaskHandle{
				WorkerStrategy: w,
				TaskID:         id,
				CachedTask:     task,
				Acquisition:    acq,
			}, nil
		}
	}
	return nil, nil
}

// TaskHandle implements jobs.TaskHandle
type TaskHandle struct {
	WorkerStrategy *WorkerStrategy
	TaskID         string
	CachedTask     *jobs.Task
	Acquisition    jobs.Acquisition
}

// Task implements TaskHandle
func (h *TaskHandle) Task() *jobs.Task {
	return h.CachedTask
}

// SubmitTask implements TaskHandle
func (h *TaskHandle) SubmitTask(task *jobs.Task) (err error) {
	if err = h.refreshTask(); err != nil {
		return
	}
	doc := NewTaskDoc(h.CachedTask)
	doc.SubTaskIDs = append(doc.SubTaskIDs, task.ID)
	if err = h.WorkerStrategy.Strategy.saveTask(doc, nil); err == nil {
		doc = NewTaskDoc(task)
		doc.State = jobs.TaskPending
		err = h.WorkerStrategy.Strategy.saveTask(doc, nil)
	}
	if err == nil {
		err = h.refreshTask()
	}
	return
}

// Update implements TaskHandle
func (h *TaskHandle) Update(task *jobs.Task) (err error) {
	if err = h.refreshTask(); err != nil {
		return
	}
	doc := NewTaskDoc(h.CachedTask)
	doc.Stage = task.Stage
	doc.ResumeTo = task.ResumeTo
	doc.State = task.State
	doc.Result = task.Result
	doc.Data = json.RawMessage(task.Data)
	doc.Output = json.RawMessage(task.Output)
	doc.Errors = task.Errors
	doc.SubTaskIDs = task.SubTaskIDs
	if err = h.WorkerStrategy.Strategy.saveTask(doc, task.Stats); err == nil {
		err = h.refreshTask()
	}
	return
}

// Done implements TaskHandle
func (h *TaskHandle) Done() error {
	return h.Acquisition.Release()
}

func (h *TaskHandle) refreshTask() error {
	err := h.Acquisition.Refresh(h.Acquisition.TTL())
	if err != nil {
		return err
	}
	task, err := h.WorkerStrategy.Strategy.QueryTask(h.TaskID)
	if err != nil {
		return err
	}
	h.CachedTask = task
	return nil
}
