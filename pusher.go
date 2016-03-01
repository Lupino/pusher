package pusher

import (
	"encoding/json"
	"github.com/Lupino/go-periodic"
	"github.com/Lupino/pusher/utils"
)

// PREFIX the perfix key of pusher.
const PREFIX = "pusher:"

// Pusher of pusher
type Pusher struct {
	ID          string   `json:"id"`
	Email       string   `json:"email"`
	NickName    string   `json:"nickname"`
	PhoneNumber string   `json:"phoneNumber"`
	Senders     []string `json:"senders"`
	Tags        []string `json:"tags"`
	CreatedAt   int64    `json:"createdAt"`
}

// NewPusher create a pusher from json bytes
func NewPusher(payload []byte) (pusher Pusher, err error) {
	err = json.Unmarshal(payload, &pusher)
	return
}

// Bytes encode pusher to json bytes
func (pusher Pusher) Bytes() (data []byte) {
	data, _ = json.Marshal(pusher)
	return
}

// HasSender on a pusher
func (pusher Pusher) HasSender(sender string) bool {
	for _, s := range pusher.Senders {
		if sender == s {
			return true
		}
	}
	return false
}

// AddSender to a pusher
func (pusher *Pusher) AddSender(sender string) bool {
	if !pusher.HasSender(sender) {
		pusher.Senders = append(pusher.Senders, sender)
		return true
	}
	return false
}

// DelSender to a pusher
func (pusher *Pusher) DelSender(sender string) bool {
	var newSenders []string
	if !pusher.HasSender(sender) {
		return false
	}
	for _, s := range pusher.Senders {
		if sender == s {
			continue
		}
		newSenders = append(newSenders, s)
	}
	pusher.Senders = newSenders
	return true
}

// HasTag on a pusher
func (pusher Pusher) HasTag(sender string) bool {
	for _, s := range pusher.Tags {
		if sender == s {
			return true
		}
	}
	return false
}

// AddTag to a pusher
func (pusher *Pusher) AddTag(sender string) bool {
	if !pusher.HasTag(sender) {
		pusher.Tags = append(pusher.Tags, sender)
		return true
	}
	return false
}

// DelTag to a pusher
func (pusher *Pusher) DelTag(sender string) bool {
	var newTags []string
	if !pusher.HasTag(sender) {
		return false
	}
	for _, s := range pusher.Tags {
		if sender == s {
			continue
		}
		newTags = append(newTags, s)
	}
	pusher.Tags = newTags
	return true
}

// SPusher server pusher
type SPusher struct {
	storer Storer
	p      *periodic.Client
}

// NewSPusher create a server pusher instance
func NewSPusher(storer Storer, p *periodic.Client) SPusher {
	return SPusher{storer: storer, p: p}
}

func (s SPusher) addSender(p Pusher, senders ...string) (err error) {
	changed := false
	for _, sender := range senders {
		if p.AddSender(sender) {
			changed = true
		}
	}

	if changed {
		if err = s.storer.Set(p); err != nil {
			return
		}
	}
	return
}

func (s SPusher) removeSender(p Pusher, senders ...string) (err error) {
	changed := false
	for _, sender := range senders {
		if p.DelSender(sender) {
			changed = true
		}
	}

	if changed {
		if err = s.storer.Set(p); err != nil {
			return
		}
	}
	return
}

func (s SPusher) addTag(p Pusher, tags ...string) (err error) {
	changed := false
	for _, tag := range tags {
		if p.AddTag(tag) {
			changed = true
		}
	}

	if changed {
		if err = s.storer.Set(p); err != nil {
			return
		}
	}
	return
}

func (s SPusher) removeTag(p Pusher, tags ...string) (err error) {
	changed := false
	for _, tag := range tags {
		if p.DelTag(tag) {
			changed = true
		}
	}

	if changed {
		if err = s.storer.Set(p); err != nil {
			return
		}
	}
	return
}

func (s SPusher) push(sender, pusher, data, schedat string) (string, error) {
	var opts = map[string]string{
		"args":    data,
		"schedat": schedat,
	}
	var name = utils.GenerateName(pusher, data)
	if err := s.p.SubmitJob(PREFIX+sender, name, opts); err != nil {
		return "", err
	}
	return name, nil
}

func (s SPusher) pushAll(sender, data, schedat string) (string, error) {
	var opts = map[string]string{
		"args":    data,
		"schedat": schedat,
	}
	var name = utils.GenerateName(sender, data)
	if err := s.p.SubmitJob(PREFIX+"pushall", name, opts); err != nil {
		return "", err
	}
	return name, nil
}
