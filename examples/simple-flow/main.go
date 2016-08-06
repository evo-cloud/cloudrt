package main

import (
	"fmt"

	jobs "github.com/evo-cloud/cloudrt/jobs"
	store "github.com/evo-cloud/cloudrt/jobs/stores/redis"
	strategy "github.com/evo-cloud/cloudrt/jobs/strategies/simple"
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

func main() {
	s := store.NewStore("localhost")
	dispatcher := jobs.NewDispatcher(s, &strategy.Strategy{Store: s})

	// register task executors
	dispatcher.
		NewTaskExec("process-job").Entry(createObj).Stage("process", processObj).
		NewTaskExec("make-component").Entry(makeComponent).
		Commit()

	dispatcher.Watcher("watcher")

	for i := 0; i < 4; i++ {
		dispatcher.Worker(fmt.Sprintf("worker%d", i))
	}

	dispatcher.NewJob().
		SetName("simple-job").
		SetTask(jobs.NewTask("process-obj").
			With(&processObjParams{Components: []string{"red", "blue", "green"}}).
			Build()).
		Submit()

	dispatcher.Run()
}
