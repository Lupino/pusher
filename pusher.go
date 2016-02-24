package pusher

import (
	"encoding/json"
	"gopkg.in/redis.v3"
	"log"
)

// Pusher of pusher
type Pusher struct {
	ID          string   `json:"id"`
	Email       string   `json:"email"`
	RealName    string   `json:"realname"`
	NickName    string   `json:"nickname"`
	PhoneNumber string   `json:"phoneNumber"`
	Senders     []string `json:"senders"`
	CreatedAt   int64    `json:"createdAt"`
}

// NewPusher create a pusher from json bytes
func NewPusher(payload string) (pusher Pusher, err error) {
	err = json.Unmarshal([]byte(payload), &pusher)
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

// SetPusher to redis and index
func SetPusher(pusher Pusher) error {
	var key = PREFIX + "pusher:" + pusher.ID
	if err := redisClient.Set(key, pusher.Bytes(), 0).Err(); err != nil {
		log.Printf("redis.Client.Set() failed(%s)", err)
		return err
	}
	if err := redisClient.ZAdd(secondIndex, redis.Z{Score: float64(pusher.CreatedAt), Member: pusher.ID}).Err(); err != nil {
		log.Printf("redis.Client.ZAdd() failed(%s)", err)
	}
	if err := index.Index(pusher.ID, pusher); err != nil {
		log.Printf("bleve.Index.Index() failed(%s)", err)
	}
	return nil
}

// GetPusher from redis by id
func GetPusher(id string) (pusher Pusher, err error) {
	var payload string
	var key = PREFIX + "pusher:" + id
	if payload, err = redisClient.Get(key).Result(); err != nil {
		log.Printf("redis.Client.Get() failed(%s)", err)
		return
	}
	return NewPusher(payload)
}

// DelPusher from redis by id
func DelPusher(id string) error {
	var key = PREFIX + "pusher:" + id
	if err := redisClient.Del(key).Err(); err != nil {
		log.Printf("redis.Client.Del() failed(%s)", err)
		return err
	}
	if err := redisClient.ZRem(secondIndex, id).Err(); err != nil {
		log.Printf("redis.Client.ZRem() failed(%s)", err)
	}
	if err := index.Delete(id); err != nil {
		log.Printf("bleve.Index.Index() failed(%s)", err)
	}
	return nil
}
