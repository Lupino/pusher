package senders

import (
	"encoding/json"
	pusherLib "github.com/Lupino/pusher"
	"github.com/sendgrid/sendgrid-go"
	"log"
)

// MailSender a sendgrid send mail sender
type MailSender struct {
	sg       *sendgrid.SGClient
	from     string
	fromName string
}

// NewMailSender new sendgrid send mail sender
func NewMailSender(sg *sendgrid.SGClient, from, fromName string) MailSender {
	return MailSender{sg: sg, from: from, fromName: fromName}
}

type mail struct {
	Subject   string `json:"subject"`
	Text      string `json:"text"`
	CreatedAt int64  `json:"created_at"`
}

// GetName for the periodic funcName
func (MailSender) GetName() string {
	return "sendmail"
}

// Send message to pusher then return sendlater
func (s MailSender) Send(pusher, data string) (int, error) {
	var (
		m    mail
		err  error
		name string
		info pusherLib.Info
	)
	if err = json.Unmarshal([]byte(data), &m); err != nil {
		log.Printf("json.Unmarshal() failed (%s)", err)
		return 0, nil
	}

	if info, err = pusherLib.GetInfo(pusher); err != nil {
		log.Printf("pusher.GetInfo() failed (%s)", err)
		return 0, nil
	}

	if info.Email == "" {
		return 0, nil
	}

	name = info.RealName
	if name == "" {
		name = info.NickName
	}

	message := sendgrid.NewMail()
	message.AddTo(info.Email)
	message.AddToName(name)
	message.SetSubject(m.Subject)
	message.SetText(m.Text)
	message.SetFrom(s.from)
	message.SetFromName(s.fromName)
	err = s.sg.Send(message)
	if err == nil {
		return 0, nil
	}
	return 10, nil
}
