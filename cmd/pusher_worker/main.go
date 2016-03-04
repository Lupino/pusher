package main

import (
	"encoding/json"
	"flag"
	"github.com/Lupino/go-periodic"
	"github.com/Lupino/pusher/worker"
	"github.com/Lupino/pusher/worker/senders"
	"github.com/sendgrid/sendgrid-go"
	"log"
	"os"
	"runtime"
)

type hookConfig struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	Secret string `json:"secret"`
}

var (
	periodicPort string
	pusherHost   string
	sgUser       string
	sgKey        string
	dayuKey      string
	dayuSecret   string
	from         string
	fromName     string
	key          string
	secret       string
	hooksFile    string
	size         int
)

func init() {
	flag.StringVar(&periodicPort, "periodic_port", "unix:///tmp/periodic.sock", "the periodic server port.")
	flag.StringVar(&pusherHost, "pusher_host", "localhost:6000", "the pusher server host.")
	flag.StringVar(&sgUser, "sendgrid_user", "", "The SendGrid username.")
	flag.StringVar(&sgKey, "sendgrid_key", "", "The SendGrid password.")
	flag.StringVar(&dayuKey, "alidayu_key", "", "The alidayu app key.")
	flag.StringVar(&dayuSecret, "alidayu_secret", "", "The alidayu app secret.")
	flag.StringVar(&from, "from", "", "The sendmail from address.")
	flag.StringVar(&fromName, "from_name", "", "The sendmail from name.")
	flag.StringVar(&key, "key", "", "the pusher server app key. (optional)")
	flag.StringVar(&secret, "secret", "", "the pusher server app secret. (optional)")
	flag.StringVar(&hooksFile, "hooks", "", "the hook sender config file. (optional)")
	flag.IntVar(&size, "size", runtime.NumCPU()*2, "the size of goroutines. (optional)")
	flag.Parse()
}

func main() {
	pw := periodic.NewWorker(size)
	if err := pw.Connect(periodicPort); err != nil {
		log.Fatal(err)
	}

	w := worker.New(pw, pusherHost, key, secret)
	var sg = sendgrid.NewSendGridClient(sgUser, sgKey)
	var mailSender = senders.NewMailSender(w, sg, from, fromName)
	var smsSender = senders.NewSMSSender(w, dayuKey, dayuSecret)
	var pushAllSender = senders.NewPushAllSender(w)

	var hooks []worker.Sender
	if len(hooksFile) > 0 {
		var hooksConfig []hookConfig
		file, err := os.Open(hooksFile)
		if err != nil {
			log.Fatal(err)
		}
		decoder := json.NewDecoder(file)
		if err = decoder.Decode(&hooksConfig); err != nil {
			log.Printf("json.NewDecoder().Decode() failed (%s)", err)
			return
		}
		for _, config := range hooksConfig {
			hook := senders.NewHookSender(w, config.Name, config.URL, config.Secret)
			hooks = append(hooks, hook)
		}
	}
	hooks = append(hooks, mailSender, smsSender, pushAllSender)
	w.RunSender(hooks...)
}
