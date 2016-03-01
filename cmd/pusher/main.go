package main

import (
	"flag"
	"github.com/Lupino/go-periodic"
	"github.com/Lupino/pusher"
	"github.com/Lupino/pusher/store/boltdb"
	"github.com/codegangsta/negroni"
	"log"
)

var periodicPort string
var root string
var host string

func init() {
	flag.StringVar(&periodicPort, "periodic_port", "unix:///tmp/periodic.sock", "the periodic server port.")
	flag.StringVar(&host, "host", "localhost:6000", "the pusher server host.")
	flag.StringVar(&root, "work_dir", ".", "The pusher work dir.")
	flag.Parse()
}

func main() {
	var err error
	var storer pusher.Storer

	pc := periodic.NewClient()
	if err = pc.Connect(periodicPort); err != nil {
		log.Fatal(err)
	}

	storer, err = boltdb.New(map[string]interface{}{
		"path": root + "/pusher.db",
	})
	if err != nil {
		log.Fatal(err)
	}

	sp := pusher.NewSPusher(storer, pc)
	n := negroni.New(negroni.NewRecovery(), negroni.NewLogger())
	n.UseHandler(sp.NewRouter())
	n.Run(host)
}
