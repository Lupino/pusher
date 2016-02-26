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

pusher worker also see [cmd/sender](https://github.com/Lupino/pusher/tree/master/cmd/sender)

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
