package pusher

import (
	"encoding/json"
	"gopkg.in/redis.v3"
	"log"
)

// Info of pusher
type Info struct {
	ID          string   `json:"id"`
	Email       string   `json:"email"`
	RealName    string   `json:"realname"`
	NickName    string   `josn:"nickname"`
	PhoneNumber string   `json:'phoneNumber'`
	Senders     []string `json:"senders"`
	CreatedAt   int64    `json:"createdAt"`
}

// NewInfo create a info from json bytes
func NewInfo(payload string) (info Info, err error) {
	err = json.Unmarshal([]byte(payload), &info)
	return
}

// Bytes encode info to json bytes
func (info Info) Bytes() (data []byte) {
	data, _ = json.Marshal(info)
	return
}

// HasSender on a pusher info
func (info Info) HasSender(sender string) bool {
	for _, s := range info.Senders {
		if s == sender {
			return true
		}
	}
	return false
}

// AddSender to a pusher info
func (info *Info) AddSender(sender string) bool {
	if !info.HasSender(sender) {
		info.Senders = append(info.Senders, sender)
		return true
	}
	return false
}

// DelSender to a pusher info
func (info *Info) DelSender(sender string) bool {
	var newSenders []string
	if !info.HasSender(sender) {
		return false
	}
	for _, s := range info.Senders {
		if s != sender {
			newSenders = append(newSenders, sender)
		}
	}
	info.Senders = newSenders
	return true
}

// SetInfo to redis and index
func SetInfo(info Info) error {
	var key = PREFIX + "info:" + info.ID
	if err := redisClient.Set(key, info.Bytes(), 0).Err(); err != nil {
		log.Printf("redis.Client.Set() failed(%s)", err)
		return err
	}
	if err := redisClient.ZAdd(secondIndex, redis.Z{Score: float64(info.CreatedAt), Member: info.ID}).Err(); err != nil {
		log.Printf("redis.Client.ZAdd() failed(%s)", err)
	}
	if err := index.Index(info.ID, info); err != nil {
		log.Printf("bleve.Index.Index() failed(%s)", err)
	}
	return nil
}

// GetInfo from redis by id
func GetInfo(id string) (info Info, err error) {
	var payload string
	var key = PREFIX + "info:" + id
	if payload, err = redisClient.Get(key).Result(); err != nil {
		log.Printf("redis.Client.Get() failed(%s)", err)
		return
	}
	return NewInfo(payload)
}

// DelInfo from redis by id
func DelInfo(id string) error {
	var key = PREFIX + "info:" + id
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
