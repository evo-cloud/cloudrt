package jobs

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
