package senders

import (
	"encoding/json"
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
	Email     string `json:"email"`
	UserName  string `json:"username"`
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
		m   mail
		err error
	)
	if err = json.Unmarshal([]byte(data), &m); err != nil {
		log.Printf("json.Unmarshal() failed (%s)", err)
		return 0, nil
	}

	message := sendgrid.NewMail()
	message.AddTo(m.Email)
	message.AddToName(m.UserName)
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
