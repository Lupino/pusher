package main

import (
	"flag"
	"github.com/Lupino/go-periodic"
	"github.com/Lupino/pusher"
	"github.com/codegangsta/negroni"
	"gopkg.in/redis.v3"
	"log"
)

var periodicPort string
var redisHost string

func init() {
	flag.StringVar(&periodicPort, "periodic_port", "unix:///tmp/periodic.sock", "the periodic server port.")
	flag.StringVar(&redisHost, "redis_host", "localhost:6379", "the redis server host.")
	flag.Parse()
}

func main() {
	rc := redis.NewClient(&redis.Options{
		Addr: redisHost,
	})
	if err := rc.Ping().Err(); err != nil {
		log.Fatal(err)
	}

	pc := periodic.NewClient()
	if err := pc.Connect(periodicPort); err != nil {
		log.Fatal(err)
	}

	pusher.SetBackend(rc, pc)
	n := negroni.New(negroni.NewRecovery(), negroni.NewLogger())
	n.UseHandler(pusher.NewRouter())
	n.Run(":3000")
}
