package jobs

import (
	"fmt"
	"sync"
)

// Strategy is the contract for scheduling strategy
type Strategy interface {
	SubmitJob(*Job) error
	CancelJob(id string) error
	IsJobCanceling(id string) (bool, error)
	QueryJob(id string) (*Job, error)
	QueryTask(id string) (*Task, error)
	NewWorker(id string) WorkerStrategy
}

// WorkerStrategy is strategy instance per worker
type WorkerStrategy interface {
	FetchTask() (TaskHandle, error)
}

// TaskHandle is the handle of a running task owned by a worker
type TaskHandle interface {
	Task() *Task
	SubmitTask(*Task) error
	Update(*Task) error
	Done() error
}

// Dispatcher submits jobs and executes tasks
type Dispatcher struct {
	Strategy Strategy
	Tasks    []*TaskExec
}

// Worker executes tasks
type Worker interface {
	Run()
}

// NewJob starts creating a job
func (d *Dispatcher) NewJob() *JobBuilder {
	return &JobBuilder{Submitter: d}
}

// SubmitJob implements JobSubmitter
func (d *Dispatcher) SubmitJob(job *Job) error {
	// TODO validate job
	return d.Strategy.SubmitJob(job)
}

// AddTaskExecs adds task executors
func (d *Dispatcher) AddTaskExecs(execs ...*TaskExec) {
	d.Tasks = append(d.Tasks, execs...)
}

// Worker spawns a worker
func (d *Dispatcher) Worker(id string) Worker {
	return &localWorker{dispatcher: d, strategy: d.Strategy.NewWorker(id)}
}

// Task queries task by id
func (d *Dispatcher) Task(id string) (Task, error) {
	task, err := d.Strategy.QueryTask(id)
	if err != nil {
		return Task{}, err
	}
	if task == nil {
		return Task{}, NotExist(id)
	}
	return *task, nil
}

// Job queries job by id
func (d *Dispatcher) Job(id string) (Job, error) {
	job, err := d.Strategy.QueryJob(id)
	if err != nil {
		return Job{}, err
	}
	if job == nil {
		return Job{}, NotExist(id)
	}
	return *job, nil
}

func (d *Dispatcher) findStage(name, stage string) *Stage {
	for _, t := range d.Tasks {
		if t.Name != name || len(t.Stages) == 0 {
			continue
		}
		if stage == "" {
			return &t.Stages[0]
		}
		for _, s := range t.Stages {
			if s.Name == stage {
				return &s
			}
		}
	}
	return nil
}

type localWorker struct {
	dispatcher *Dispatcher
	strategy   WorkerStrategy
}

type localContext struct {
	worker   *localWorker
	handle   TaskHandle
	subTasks []Task
	lock     sync.Mutex
}

func (l *localContext) dispatcher() *Dispatcher {
	return l.worker.dispatcher
}

func (l *localContext) taskHandle() TaskHandle {
	return l.handle
}

func (l *localContext) strategy() Strategy {
	return l.dispatcher().Strategy
}

func (l *localContext) pushSubTask(task Task) {
	l.lock.Lock()
	defer l.lock.Unlock()
	current := l.handle.Task()
	task.JobID = current.JobID
	task.ParentID = current.ID
	l.subTasks = append(l.subTasks, task)
}

func (w *localWorker) Run() {
	for {
		handle, err := w.strategy.FetchTask()
		if err == nil && handle != nil {
			w.runTaskByHandle(handle)
			handle.Done()
		}
	}
}

func (w *localWorker) runTaskByHandle(handle TaskHandle) {
	ctx := Context{
		local: &localContext{
			worker: w,
			handle: handle,
		},
	}

	err := w.runTask(ctx)
	if err != nil {
		taskErr, ok := err.(*TaskError)
		if !ok {
			taskErr = ctx.Fail(err)
		}
		err = w.taskComplete(ctx, taskErr)
	} else {
		err = w.taskComplete(ctx, nil)
	}
	if err != nil {
		// TODO
	}
}

func (w *localWorker) runTask(ctx Context) error {
	task := ctx.Task()
	stage := w.dispatcher.findStage(task.Name, task.Stage)
	if stage == nil {
		return fmt.Errorf("invalid task/stage: %s/%s", task.Name, task.Stage)
	}

	if stage.Fn == nil {
		return nil
	}

	return stage.Fn(ctx)
}

func (w *localWorker) taskComplete(ctx Context, taskErr *TaskError) error {
	task := ctx.Task()
	if taskErr != nil {
		task.Errors = append(task.Errors, *taskErr)
		switch taskErr.Type {
		case TaskErrIgnored:
			task.State = TaskCompleted
			if task.Revert {
				task.Result = TaskAborted
			} else {
				task.Result = TaskSuccess
			}
		case TaskErrFail:
			// TODO
		case TaskErrRetry:
			// TODO
		case TaskErrRevert:
			// TODO
		case TaskErrStuck:
			task.State = TaskStucked
		}
		err := ctx.local.handle.Update(&task)
		if err != nil {
			return err
		}
	} else {
		for _, subTask := range ctx.local.subTasks {
			task.SubTaskIDs = append(task.SubTaskIDs, subTask.ID)
		}
		// TODO
	}
	return nil
}
