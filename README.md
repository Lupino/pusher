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
    cd $GOPATH/src/github.com/Lupino/pusher
    go install ./...

Getting start with pusher command
---------------------------------

### First start the needed server
* Start [Periodic task system](https://github.com/Lupino/periodic) with a sample command `periodic -d`

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

p := pusher.NewSPusher(...)
http.ListenAndServe(":6000", p.NewRouter())
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
w := worker.New(pw, periodicHost, key, secret)
var mailSender = senders.NewMailSender(w, ...)
var smsSender = senders.NewSMSSender(w, ...)
var pushAllSender = senders.NewPushAllSender(w, ...)
w.RunSender(mailSender, smsSender, pushAllSender)
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

Write you own backend storage
-----------------------------
Write you own backend with the `Storer` interface.
see example [store/boltdb](https://github.com/Lupino/pusher/tree/master/store/boltdb)

```go
// Storer interface for store pusher data
type Storer interface {
	Set(Pusher) error
	Get(string) (Pusher, error)
	Del(string) error
	GetAll(from, size int) (uint64, []Pusher, error)
}
```

Use pusher auth middleware
--------------------------
If you need auth the pusher api, just add `Auth` middleware to you http server,
or pass `-key` and `-secret` on the pusher command.

On client add http header:
```
X-App-Key: app_key
X-Request-Time: current unix time stamp
X-Request-Signature: signature
```

the `signature` use `hmac_md5`:

* First initial `hmac_md5` use the `secret`
* Second get sign params with request path and query or data. eg:
```go
var signParams = make(map[string]string)
signParams["app_key"] = key
signParams["path"] = "/pusher/search/"
signParams["timestamp"] = timestamp
signParams["q"] = q
signParams["from"] = from
```
* Third sort the key with `ascii` order, join them.
eg: `foo=1`, `bar=2`, `foo_bar=3`, `foobar=4`, sort and join them is `bar2foo1foo_bar3foobar4`
* Fourth `hmac_md5` the join data.
* Five `hex` them and get upper case
* also see [HmacMD5](https://github.com/Lupino/pusher/blob/master/utils/utils.go#L35) and [worker/api.go](https://github.com/Lupino/pusher/blob/master/worker/api.go)

Requirements
------------

* [golang](http://golang.org)
* [periodic](https://github.com/Lupino/periodic)
