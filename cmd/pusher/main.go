package main

import (
	"flag"
	"github.com/Lupino/go-periodic"
	"github.com/Lupino/pusher"
	"github.com/blevesearch/bleve"
	"github.com/codegangsta/negroni"
	"gopkg.in/redis.v3"
	"log"
)

var periodicPort string
var redisHost string
var root string
var host string

func init() {
	flag.StringVar(&periodicPort, "periodic_port", "unix:///tmp/periodic.sock", "the periodic server port.")
	flag.StringVar(&redisHost, "redis_host", "localhost:6379", "the redis server host.")
	flag.StringVar(&host, "host", "localhost:6000", "the pusher server host.")
	flag.StringVar(&root, "work_dir", ".", "The pusher work dir.")
	flag.Parse()
}

func main() {
	var err error
	rc := redis.NewClient(&redis.Options{
		Addr: redisHost,
	})
	if err = rc.Ping().Err(); err != nil {
		log.Fatal(err)
	}

	pc := periodic.NewClient()
	if err = pc.Connect(periodicPort); err != nil {
		log.Fatal(err)
	}

	var index bleve.Index
	if index, err = bleve.Open(root + "/pusher.bleve"); err != nil {
		mapping := bleve.NewIndexMapping()
		if index, err = bleve.New(root+"/pusher.bleve", mapping); err != nil {
			log.Fatal(err)
		}
	}

	pusher.SetBackend(rc, pc, index)
	n := negroni.New(negroni.NewRecovery(), negroni.NewLogger())
	n.UseHandler(pusher.NewRouter())
	n.Run(host)
}
