package main

import (
	"flag"
	"github.com/Lupino/go-periodic"
	"github.com/Lupino/pusher/worker"
	"github.com/Lupino/pusher/worker/senders"
	"github.com/sendgrid/sendgrid-go"
	"log"
)

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
	flag.Parse()
}

func main() {
	pw := periodic.NewWorker()
	if err := pw.Connect(periodicPort); err != nil {
		log.Fatal(err)
	}

	w := worker.New(pw, pusherHost, key, secret)
	var sg = sendgrid.NewSendGridClient(sgUser, sgKey)
	var mailSender = senders.NewMailSender(w, sg, from, fromName)
	var smsSender = senders.NewSMSSender(w, dayuKey, dayuSecret)
	var pushAllSender = senders.NewPushAllSender(w)
	w.RunSender(mailSender, smsSender, pushAllSender)
}
