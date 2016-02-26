pusher
==============
pusher is a push server writen by golang.

## Features
 * Light weight
 * High performance
 * Pure Golang
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

Usage
-----

pusher server also see [cmd/pusher](https://github.com/Lupino/pusher/tree/master/cmd/pusher)

```go
import "github.com/Lupino/pusher"
import "github.com/codegangsta/negroni"

pusher.SetBackend(...)
n := negroni.New(negroni.NewRecovery(), negroni.NewLogger())
n.UseHandler(pusher.NewRouter())
n.Run(":3000")
```

pusher worker also see [cmd/pusher_worker](https://github.com/Lupino/pusher/tree/master/cmd/pusher_worker)

```go
import (
	"github.com/Lupino/pusher"
	"github.com/Lupino/pusher/senders"
)

var mailSender = senders.NewMailSender(...)
var smsSender = senders.NewSMSSender(...)
var pushAllSender = senders.NewPushAllSender(...)
pusher.RunWorker(pw, mailSender, smsSender, pushAllSender)
```

Requirements
------------

* [golang](http://golang.org)
* [periodic](https://github.com/Lupino/periodic)
* [Redis](http://redis.io)
