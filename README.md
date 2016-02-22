A sample push server

Install
-------

    go get -v github.com/Lupino/pusher

Usage
-----

pusher server also see cmd/pusher/main.go

```go
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

	n := negroni.New(negroni.NewRecovery(), negroni.NewLogger())
	n.UseHandler(pusher.NewPusher(rc, pc))
	n.Run(":3000")
}
```

pusher worker also see cmd/pusher_sample_worker

```go
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
```

Requirements
------------

* [golang](http://golang.org)
* [periodic](https://github.com/Lupino/periodic)
* [Redis](http://redis.io)
