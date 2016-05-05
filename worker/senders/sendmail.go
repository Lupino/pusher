package senders

import (
	"bytes"
	"encoding/json"
	pusherLib "github.com/Lupino/pusher"
	"github.com/Lupino/pusher/worker"
	"github.com/sendgrid/sendgrid-go"
	"log"
	"text/template"
)

// MailSender a sendgrid send mail sender
type MailSender struct {
	sg       *sendgrid.SGClient
	from     string
	fromName string
	w        worker.Worker
}

// NewMailSender new sendgrid send mail sender
func NewMailSender(w worker.Worker, sg *sendgrid.SGClient, from, fromName string) MailSender {
	return MailSender{sg: sg, from: from, fromName: fromName, w: w}
}

type mail struct {
	Subject   string `json:"subject"`
	Text      string `json:"text"`
	CreatedAt int64  `json:"createdAt"`
}

// GetName for the periodic funcName
func (MailSender) GetName() string {
	return "sendmail"
}

// Send message to pusher then return sendlater
func (s MailSender) Send(pusher, data string, counter int) (int, error) {
	var (
		m      mail
		err    error
		name   string
		p      pusherLib.Pusher
		text   string
		tpl    *template.Template
		buffer = bytes.NewBuffer(nil)
	)
	if err = json.Unmarshal([]byte(data), &m); err != nil {
		log.Printf("json.Unmarshal() failed (%s)", err)
		return 0, nil
	}

	if p, err = s.w.GetAPI().GetPusher(pusher); err != nil {
		return 0, nil
	}

	if p.Email == "" {
		return 0, nil
	}

	name = p.NickName

	text = m.Text
	if tpl, err = template.New("text").Parse(m.Text); err != nil {
		log.Printf("template.New().Parse() failed (%s)", err)
	} else {
		if err = tpl.Execute(buffer, p); err != nil {
			log.Printf("template.Template.Execute() failed (%s)", err)
		} else {
			text = string(buffer.Bytes())
		}
	}

	message := sendgrid.NewMail()
	message.AddTo(p.Email)
	message.AddToName(name)
	message.SetSubject(m.Subject)
	message.SetHTML(text)
	message.SetFrom(s.from)
	message.SetFromName(s.fromName)
	err = s.sg.Send(message)
	if err != nil {
		log.Printf("sendgrid.SGClient.Send() failed (%s)", err)
	}
	return 0, nil
}
