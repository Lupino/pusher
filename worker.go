package pusher

import (
	"github.com/Lupino/go-periodic"
	"log"
)

var periodicWorker *periodic.Worker

func warperSender(sender Sender) func(periodic.Job) {
	return func(job periodic.Job) {
		pusher := extractPusher(job.Name)
		if !verifyData(job.Name, pusher, job.Args) {
			log.Printf("verifyData() failed (%s) ignore\n", job.Name)
			job.Done() // ignore invalid job
			return
		}
		later, err := sender.Send(pusher, job.Args)

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
func RunWorker(w *periodic.Worker, senders ...Sender) {
	for _, sender := range senders {
		w.AddFunc(PREFIX+sender.GetName(), warperSender(sender))
		log.Printf("Loaded sender (%s)", sender.GetName())
	}
	w.Work()
}
