package jobs

type localWatcher struct {
	id         string
	dispatcher *Dispatcher
}

func (w *localWatcher) Run(stopCh StopChan) {
	// TODO
}
