package jobs

import "fmt"

// Strategy is the contract for scheduling strategy
type Strategy interface {
	SubmitJob(*Job) error
	NewWorker() WorkerStrategy
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
	Done(*TaskError) error
}

// Store is the persistent storage for jobs/tasks
type Store interface {
}

// Dispatcher submits jobs and executes tasks
type Dispatcher struct {
	Strategy Strategy
	Store    Store
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
	// TODO
	return nil
}

// AddTaskExecs adds task executors
func (d *Dispatcher) AddTaskExecs(execs ...*TaskExec) {
	d.Tasks = append(d.Tasks, execs...)
}

// Worker spawns a worker`
func (d *Dispatcher) Worker() Worker {
	return &localWorker{dispatcher: d, strategy: d.Strategy.NewWorker()}
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

func (w *localWorker) Run() {
	for {
		handle, err := w.strategy.FetchTask()
		if err == nil && handle != nil {
			w.runTaskByHandle(handle)
		}
	}
}

func (w *localWorker) runTaskByHandle(handle TaskHandle) {
	ctx := Context{
		strategy:   w.strategy,
		taskHandle: handle,
	}

	err := w.runTask(ctx)
	if err != nil {
		taskErr, ok := err.(*TaskError)
		if !ok {
			taskErr = ctx.Fail(err)
		}
		err = handle.Done(taskErr)
	} else {
		err = handle.Done(nil)
	}
	if err != nil {
		// TODO
	}
}

func (w *localWorker) runTask(ctx Context) error {
	task := ctx.Current()
	stage := w.dispatcher.findStage(task.Name, task.Stage)
	if stage == nil {
		return fmt.Errorf("invalid task/stage: %s/%s", task.Name, task.Stage)
	}

	if stage.Fn == nil {
		return nil
	}

	return stage.Fn(ctx)
}
