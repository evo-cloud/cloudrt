package jobs

// JobState defines the state of the job
type JobState int

// Job states
const (
	JobCreated  JobState = iota // Job is created and ready for run
	JobRunning                  // Job is running
	JobStucked                  // Some of the job is stucked
	JobFinished                 // Job completed
)

// Job defines the details of a job
type Job struct {
	ID    string   `json:"id"`    // globally unique job id
	Name  string   `json:"name"`  // job name, optionally
	Task  *Task    `json:"task"`  // the entry task
	State JobState `json:"state"` // current job state
}

// JobSubmitter defines the contract which submits a job
type JobSubmitter interface {
	SubmitJob(*Job) error
}

// JobBuilder is the helper creates a job
type JobBuilder struct {
	Submitter JobSubmitter
	ID        string
	Name      string
	Task      *Task
}

// SetID specifies the globally unique job id
func (b *JobBuilder) SetID(id string) *JobBuilder {
	b.ID = id
	return b
}

// SetName specifies the job name
func (b *JobBuilder) SetName(name string) *JobBuilder {
	b.Name = name
	return b
}

// SetTask specifies the entry task
func (b *JobBuilder) SetTask(task *Task) *JobBuilder {
	b.Task = task
	return b
}

// Submit submits the job for execution
func (b *JobBuilder) Submit() (*Job, error) {
	job := &Job{
		ID:   b.ID,
		Name: b.Name,
		Task: b.Task,
	}
	if job.ID == "" {
		// TODO generate ID
	}
	job.Task.JobID = job.ID
	return job, b.Submitter.SubmitJob(job)
}
