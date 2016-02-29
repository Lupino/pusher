pusher
==============
pusher is a push server writen by golang.

## Features
 * Light weight
 * High performance
 * Pure Golang
 * Embeddable
 * Supports single push and pushall
 * Supports push or pushall with a schedule time
 * Scalable architecture (Unlimited dynamic message and sender modules)
 * Asynchronous push notification based on [Periodic task system](https://github.com/Lupino/periodic)

Install
-------

    go get -v github.com/Lupino/pusher
    go install ./...

Getting start with pusher command
---------------------------------

### First start the needed server
* Start [Periodic task system](https://github.com/Lupino/periodic) with a sample command `periodic -d`
* Start [Redis](http://redis.io) with the command `redis-server`

### Second start pusher api server
* Start pusher api server with the command `pusher`
* The pusher api server default run on <http://localhost:6000>

### Third start pusher worker
* Go to <https://sendgrid.com/> register an account and get the `key` and `user`
* Go to <http://www.alidayu.com/> register an account and build an app, then get the `app key` and `app secret`
* Start pusher worker with the command

```bash
pusher_worker -sendgrid_key=sgKey \
              -sendgrid_user=sgUser \
              -from=example@example.com \
              -from_name=example \
              -alidayu_key=alidayuAppKey \
              -alidayu_secret=alidayuAppSecret

```

### Fourth push the message with curl
* Create a pusher
```bash
curl -i http://localhost:6000/pusher/pushers/ \
     -d pusher=lupino \
     -d email=example@example.com \
     -d phoneNumber=12345678901 \
     -d nickname=xxxxx \
     -d createdAt=1456403493
```
* Add sender to pusher
```bash
curl -i http://localhost:6000/pusher/sendmail/add -d pusher=lupinno
curl -i http://localhost:6000/pusher/sendsms/add -d pusher=lupinno
```
* Push a message
```bash
curl -i http://localhost:6000/pusher/sendmail/push \
     -d pusher=lupino \
     -d data='{"subject": "subject", "text": "text"}'
```

* Full api docs sees <http://lupino.github.io/pusher/>

Use pusher as a package
-----------------------

Embed pusher server to you current http server,
also see [cmd/pusher](https://github.com/Lupino/pusher/tree/master/cmd/pusher)

```go
import "github.com/Lupino/pusher"
import "net/http"

pusher.SetBackend(...)
r := pusher.NewRouter()
http.ListenAndServe(":6000", r)
```

Embed pusher worker to you current project,
also see [cmd/pusher_worker](https://github.com/Lupino/pusher/tree/master/cmd/pusher_worker)

```go
import (
	"github.com/Lupino/pusher/worker"
	"github.com/Lupino/pusher/worker/senders"
	"github.com/Lupino/go-periodic"
)

pw := periodic.NewWorker()
var mailSender = senders.NewMailSender(...)
var smsSender = senders.NewSMSSender(...)
var pushAllSender = senders.NewPushAllSender(...)
worker.RunSender(pw, mailSender, smsSender, pushAllSender)
```

Write you own sender
--------------------

Write you own sender with the `Sender` interface.
see example [cmd/pusher_sample_worker](https://github.com/Lupino/pusher/tree/master/cmd/pusher_sample_worker)

```go
// Sender interface for pusher
type Sender interface {
	// GetName for the periodic funcName
	GetName() string
	// Send message to pusher then return sendlater
	// if err != nil job fail
	// if sendlater > 0 send later
	// if sendlater == 0 send done
	Send(pusher, data string) (sendlater int, err error)
}
```

Requirements
------------

* [golang](http://golang.org)
* [periodic](https://github.com/Lupino/periodic)
* [Redis](http://redis.io)
