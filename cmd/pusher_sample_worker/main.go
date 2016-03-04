package main

import (
	"flag"
	"github.com/Lupino/go-periodic"
	"github.com/Lupino/pusher/worker"
	"log"
)

var periodicPort string

type sampleSender struct{}

func (p sampleSender) GetName() string {
	return "sample_sender"
}

func (p sampleSender) Send(pusher, data string, counter int) (int, error) {

	// schedlater 10s
	if data == "1" {
		return 10, nil
	}

	// fail the job
	// return 0, fmt.Errorf("pusher[%s] do fail", pusher)

	// done the job
	return 0, nil
}

func init() {
	flag.StringVar(&periodicPort, "periodic_port", "unix:///tmp/periodic.sock", "the periodic server port.")
	flag.Parse()
}

func main() {
	pw := periodic.NewWorker()
	if err := pw.Connect(periodicPort); err != nil {
		log.Fatal(err)
	}

	w := worker.New(pw, "", "", "")
	w.RunSender(sampleSender{})
}
