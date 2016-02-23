A sample push server

Install
-------

    go get -v github.com/Lupino/pusher

Usage
-----

pusher server also see [cmd/pusher](https://github.com/Lupino/pusher/tree/master/cmd/pusher)

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

pusher worker also see [cmd/sender](https://github.com/Lupino/pusher/tree/master/cmd/sender)

```go
package main

import (
	"flag"
	"github.com/Lupino/go-periodic"
	"github.com/Lupino/pusher"
	"github.com/Lupino/pusher/senders"
	"github.com/sendgrid/sendgrid-go"
	"log"
)

var (
	periodicPort string
	sgUser       string
	sgKey        string
	dayuKey      string
	dayuSecret   string
	from         string
	fromName     string
	signName     string
	template     string
)

func main() {
	pw := periodic.NewWorker()
	if err := pw.Connect(periodicPort); err != nil {
		log.Fatal(err)
	}
	var sg = sendgrid.NewSendGridClient(sgUser, sgKey)
	var mailSender = senders.NewMailSender(sg, from, fromName)
	var smsSender = senders.NewSMSSender(dayuKey, dayuSecret, signName, template)
	pusher.RunWorker(pw, mailSender, smsSender)
}
```

Requirements
------------

* [golang](http://golang.org)
* [periodic](https://github.com/Lupino/periodic)
* [Redis](http://redis.io)
