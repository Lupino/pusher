A sample push server

Install
-------

    go get -v github.com/Lupino/pusher

Usage
-----

pusher server also see [cmd/pusher](https://github.com/Lupino/pusher/tree/master/cmd/pusher)

```go
import (
	"github.com/Lupino/go-periodic"
	"github.com/Lupino/pusher"
	"github.com/blevesearch/bleve"
	"gopkg.in/redis.v3"
	"github.com/codegangsta/negroni"
)

var rc *redis.Client
var pc *periodic.Client
var index bleve.Index
pusher.SetBackend(rc, pc, index)
n := negroni.New(negroni.NewRecovery(), negroni.NewLogger())
n.UseHandler(pusher.NewRouter())
n.Run(":3000")
```

pusher worker also see [cmd/sender](https://github.com/Lupino/pusher/tree/master/cmd/sender)

```go
import (
	"github.com/Lupino/go-periodic"
	"github.com/Lupino/pusher"
	"github.com/Lupino/pusher/senders"
	"github.com/sendgrid/sendgrid-go"
)

var (
	periodicPort string
	sgUser       string
	sgKey        string
	dayuKey      string
	dayuSecret   string
	from         string
	fromName     string
)

pw := periodic.NewWorker()
var sg = sendgrid.NewSendGridClient(sgUser, sgKey)
var mailSender = senders.NewMailSender(sg, from, fromName)
var smsSender = senders.NewSMSSender(dayuKey, dayuSecret)
pusher.RunWorker(pw, mailSender, smsSender)
}
```

Requirements
------------

* [golang](http://golang.org)
* [periodic](https://github.com/Lupino/periodic)
* [Redis](http://redis.io)
