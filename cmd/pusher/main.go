package main

import (
	"flag"
	"github.com/Lupino/go-periodic"
	"github.com/Lupino/pusher"
	"github.com/Lupino/pusher/store/boltdb"
	"github.com/codegangsta/negroni"
	"log"
	"os"
)

var (
	periodicPort string
	root         string
	host         string
	key          string
	secret       string
)

func init() {
	flag.StringVar(&periodicPort, "periodic_port", "unix:///tmp/periodic.sock", "the periodic server port.")
	flag.StringVar(&host, "host", "localhost:6000", "the pusher server host.")
	flag.StringVar(&key, "key", "", "the pusher server app key. (optional)")
	flag.StringVar(&secret, "secret", "", "the pusher server app secret. (optional)")
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

	var path = root + "/pusher.db"
	if err = os.MkdirAll(path, 0755); err != nil {
		log.Fatal(err)
	}

	storer, err = boltdb.New(map[string]interface{}{
		"path": path,
	})
	if err != nil {
		log.Fatal(err)
	}

	sp, err := pusher.NewSPusher(storer, pc, path, key, secret)
	if err != nil {
		log.Fatal(err)
	}
	n := negroni.New(negroni.NewRecovery(), negroni.NewLogger())
	if len(key) > 0 {
		n.Use(negroni.HandlerFunc(sp.Auth))
	}
	n.UseHandler(sp.NewRouter())
	n.Run(host)
}
