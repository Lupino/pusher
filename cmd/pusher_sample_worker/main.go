package main

import (
	"flag"
	"github.com/Lupino/go-periodic"
	"github.com/Lupino/pusher"
	"log"
)

var periodicPort string

type samplePlugin struct{}

func (p samplePlugin) GetGroupName() string {
	return "sample_plugin"
}

func (p samplePlugin) Do(pusher, data string) (int, error) {

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
	pusher.RunWorker(pw, samplePlugin{})
}
