package main

import (
	jobs "github.com/evo-cloud/cloudrt/jobs"
)

type processObjParams struct {
	Components []string
}

func createObj(ctx jobs.Context) error {
	var params processObjParams
	if err := ctx.GetParams(&params); err != nil {
		return ctx.Fail(err)
	}
	for _, component := range params.Components {
		_, err := ctx.NewTask("make-component").
			With(component).
			Submit()
		if err != nil {
			return ctx.Fail(err)
		}
	}
	return ctx.ResumeTo("process")
}

func processObj(ctx jobs.Context) error {
	subTasks, err := ctx.SubTasks()
	if err != nil {
		return ctx.FailRetry(err)
	}
	var output []string
	for _, t := range subTasks {
		var out string
		if err = t.GetOutput(&out); err != nil {
			return ctx.Fail(err)
		}
		output = append(output, out)
	}
	return ctx.SetOutput(output)
}

func makeComponent(ctx jobs.Context) error {
	var component string
	if err := ctx.GetParams(&component); err != nil {
		return ctx.Fail(err)
	}
	output := component + ":done"
	if err := ctx.SetOutput(output); err != nil {
		return ctx.Fail(err)
	}
	return nil
}

type dummyStore struct{}

func main() {
	dispatcher := &jobs.Dispatcher{Store: &dummyStore{}}

	// register task executors
	dispatcher.AddTaskExecs(
		&jobs.TaskExec{
			Name: "process-obj",
			Stages: []jobs.Stage{
				{Name: "entry", Fn: createObj},
				{Name: "process", Fn: processObj},
			},
		},
		&jobs.TaskExec{
			Name:   "make-component",
			Stages: []jobs.Stage{{Fn: makeComponent}},
		},
	)

	// start 4 workers
	for i := 0; i < 4; i++ {
		worker := dispatcher.Worker()
		go worker.Run()
	}

	dispatcher.NewJob().
		SetName("simple-job").
		SetTask(jobs.NewTask("process-obj").
			With(&processObjParams{Components: []string{"red", "blue", "green"}}).
			Build()).
		Submit()
	dispatcher.Wait()
}
