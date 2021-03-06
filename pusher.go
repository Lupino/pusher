package pusher

import (
	"encoding/json"
	"github.com/Lupino/go-periodic"
	"github.com/Lupino/pusher/utils"
	"github.com/blevesearch/bleve"
	"log"
)

// PREFIX the default perfix key of pusher.
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
	key    string
	secret string
	path   string
	prefix string
	index  bleve.Index
}

// NewSPusher create a server pusher instance
func NewSPusher(storer Storer, p *periodic.Client, path string) (sp SPusher, err error) {
	var index bleve.Index
	if index, err = openIndex(path); err != nil {
		return
	}
	sp = SPusher{storer: storer, p: p, path: path, index: index, prefix: PREFIX}
	return
}

// SetKey server pusher app key
func (s *SPusher) SetKey(key string) {
	s.key = key
}

// SetSecret server pusher app secret
func (s *SPusher) SetSecret(secret string) {
	s.secret = secret
}

// SetPrefix set prefix key for periodic
func (s *SPusher) SetPrefix(prefix string) {
	s.prefix = prefix
}

func (s SPusher) addSender(p Pusher, senders ...string) (err error) {
	changed := false
	for _, sender := range senders {
		if p.AddSender(sender) {
			changed = true
		}
	}

	if changed {
		if err = s.savePusher(p); err != nil {
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
		if err = s.savePusher(p); err != nil {
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
		if err = s.savePusher(p); err != nil {
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
		if err = s.savePusher(p); err != nil {
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
	if err := s.p.SubmitJob(s.prefix+sender, name, opts); err != nil {
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
	if err := s.p.SubmitJob(s.prefix+"pushall", name, opts); err != nil {
		return "", err
	}
	return name, nil
}

func (s SPusher) savePusher(p Pusher) (err error) {
	if err = s.storer.Set(p); err != nil {
		return
	}
	if err = s.index.Index(p.ID, p); err != nil {
		log.Printf("bleve.Index.Index() failed(%s)", err)
	}
	return nil
}

func (s SPusher) removePusher(p string) (err error) {
	if err = s.storer.Del(p); err != nil {
		return
	}
	if err = s.index.Delete(p); err != nil {
		log.Printf("bleve.Index.Delete() failed(%s)", err)
	}
	return nil
}
