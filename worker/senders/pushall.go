package senders

import (
	"encoding/json"
	pusherLib "github.com/Lupino/pusher"
	"github.com/Lupino/pusher/worker"
	"log"
)

// PushAllSender a pushall sender to process pushall api
type PushAllSender struct {
	w worker.Worker
}

// NewPushAllSender new push all sender
func NewPushAllSender(w worker.Worker) PushAllSender {
	return PushAllSender{
		w: w,
	}
}

// GetName for the periodic funcName
func (PushAllSender) GetName() string {
	return "pushall"
}

// Send message to pusher then return sendlater
func (s PushAllSender) Send(sender, data string, counter int) (int, error) {
	var (
		err      error
		pushers  []pusherLib.Pusher
		total    int
		from     = 0
		size     = 10
		query    = "senders:" + sender
		workdata map[string]string
	)
	if err = json.Unmarshal([]byte(data), &workdata); err != nil {
		log.Printf("json.Unmarshal() failed (%s)", err)
		return 0, nil
	}

	if tag, ok := workdata["tag"]; ok && len(tag) > 0 {
		query = query + " tags:" + tag
	}

	api := s.w.GetAPI()
	if total, pushers, err = api.SearchPusher(query, from, size); err != nil {
		return 10 * counter, nil
	}
	s.pushs(sender, pushers, workdata["data"])
	for from = size; from < total; from = from + size {
		_, pushers, _ = api.SearchPusher(query, from, size)
		s.pushs(sender, pushers, workdata["data"])
	}
	return 0, nil
}

func (s PushAllSender) pushs(sender string, pushers []pusherLib.Pusher, data string) {
	api := s.w.GetAPI()
	for _, pusher := range pushers {
		api.Push(sender, pusher.ID, data)
	}
}
