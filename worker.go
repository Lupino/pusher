package pusher

import (
	"github.com/Lupino/go-periodic"
	"log"
	"strings"
)

var periodicWorker *periodic.Worker

func extractName(name string) string {
	parts := strings.SplitN(name, "_", 2)
	return parts[0]
}

func verifyData(expect, pusher, data string) bool {
	got := generateName(pusher, data)
	return expect == got
}

func warperPlugin(plugin Plugin) func(periodic.Job) {
	return func(job periodic.Job) {
		pusher := extractPusher(job.Name)
		if !verifyData(job.Name, pusher, job.Args) {
			log.Printf("verifyData() failed (%s) ignore\n", job.Name)
			job.Done() // ignore invalid job
			return
		}
		later, err := plugin.Do(pusher, job.Args)

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
		w.AddFunc(PREFIX+plugin.GetGroupName(), warperPlugin(plugin))
	}
	w.Work()
}
