package jobs

import "encoding/json"

// Context provides the context for a running task
type Context struct {
}

// JobID retrieves the current job id
func (c Context) JobID() string {
	return c.Current().JobID
}

// TaskID retrieves the current task id
func (c Context) TaskID() string {
	return c.Current().ID
}

// Current returns a copy of current task
func (c Context) Current() Task {
	// TODO
	return Task{}
}

// SubTasks retrieves sub tasks
func (c Context) SubTasks() ([]Task, error) {
	// TODO
	return nil, nil
}

// GetParams extracts the parameters for current task
func (c Context) GetParams(p interface{}) error {
	params := c.Current().Params
	if params == nil {
		return nil
	}
	return json.Unmarshal(params, p)
}

// SetData saves the data of the task
func (c Context) SetData(p interface{}) error {
	// TODO
	return nil
}

// SetOutput saves the output of the task
func (c Context) SetOutput(p interface{}) error {
	// TODO
	return nil
}

// ResumeTo specifies the next stage when sub tasks finish
func (c Context) ResumeTo(stage string) error {
	// TODO
	return nil
}

// NewTask starts creating a new sub task
func (c Context) NewTask(name string) *TaskBuilder {
	return &TaskBuilder{Submitter: c, Name: name}
}

// SubmitTask implements TaskSubmitter
func (c Context) SubmitTask(task *Task) error {
	task.JobID = c.JobID()
	task.ParentID = c.TaskID()
	// TODO
	return nil
}
