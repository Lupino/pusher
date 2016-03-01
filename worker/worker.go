package worker

import (
	"github.com/Lupino/go-periodic"
	"github.com/Lupino/pusher/utils"
	"log"
)

// PREFIX the perfix key of pusher.
const PREFIX = "pusher:"

func warperSender(sender Sender) func(periodic.Job) {
	return func(job periodic.Job) {
		pusher := utils.ExtractPusher(job.Name)
		if !utils.VerifyData(job.Name, pusher, job.Args) {
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

// Worker for pusher
type Worker struct {
	w   *periodic.Worker
	api API
}

// New worker
func New(w *periodic.Worker, host string) Worker {
	return Worker{
		w:   w,
		api: API{host: host},
	}
}

// RunSender by periodic worker
func (w Worker) RunSender(senders ...Sender) {
	for _, sender := range senders {
		w.w.AddFunc(PREFIX+sender.GetName(), warperSender(sender))
		log.Printf("Loaded sender (%s)", sender.GetName())
	}
	w.w.Work()
}

// GetAPI return some implement pusher client api
func (w Worker) GetAPI() API {
	return w.api
}
