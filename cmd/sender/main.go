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

func init() {
	flag.StringVar(&periodicPort, "periodic_port", "unix:///tmp/periodic.sock", "the periodic server port.")
	flag.StringVar(&sgUser, "sendgrid_user", "", "The SendGrid username.")
	flag.StringVar(&sgKey, "sendgrid_key", "", "The SendGrid password.")
	flag.StringVar(&dayuKey, "alidayu_key", "", "The alidayu app key.")
	flag.StringVar(&dayuSecret, "alidayu_secret", "", "The alidayu app secret.")
	flag.StringVar(&from, "from", "", "The sendmail from address.")
	flag.StringVar(&fromName, "from_name", "", "The sendmail from name.")
	flag.StringVar(&signName, "sign_name", "", "The alidayu SMS sign name.")
	flag.StringVar(&template, "template", "", "The alidayu SMS template.")
	flag.Parse()
}

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
