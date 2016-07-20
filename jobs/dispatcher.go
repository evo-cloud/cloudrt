package jobs

// Strategy is the contract for scheduling strategy
type Strategy interface {
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
	return &localWorker{dispatcher: d}
}

// Wait waits until all workers exit
func (d *Dispatcher) Wait() error {
	// TODO
	return nil
}

type localWorker struct {
	dispatcher *Dispatcher
}

func (w *localWorker) Run() {
	// TODO
}
