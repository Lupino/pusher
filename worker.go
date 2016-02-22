package pusher

import (
	"github.com/Lupino/go-periodic"
)

var periodicWorker *periodic.Worker

func warperPlugin(plugin Plugin) func(periodic.Job) {
	return func(job periodic.Job) {
		later, err := plugin.Do(job.Name, job.Args)

		if err != nil {
			job.Fail()
		} else if later > 0 {
			job.SchedLater(later)
		} else {
			job.Done()
		}
	}
}

// RunWorker defined new pusher worker
func RunWorker(w *periodic.Worker, plugins ...Plugin) {
	for _, plugin := range plugins {
		w.AddFunc(plugin.GetGroupName(), warperPlugin(plugin))
	}
	w.Work()
}
