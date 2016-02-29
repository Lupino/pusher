package senders

import (
	pusherLib "github.com/Lupino/pusher"
)

// PushAllSender a pushall sender to process pushall api
type PushAllSender struct {
	pusherRoot string
}

// NewPushAllSender new push all sender
func NewPushAllSender(pusherRoot string) PushAllSender {
	return PushAllSender{
		pusherRoot: pusherRoot,
	}
}

// GetName for the periodic funcName
func (PushAllSender) GetName() string {
	return "pushall"
}

// Send message to pusher then return sendlater
func (s PushAllSender) Send(sender, data string) (int, error) {
	var (
		err     error
		pushers []pusherLib.Pusher
		total   int
		from    = 0
		size    = 10
	)
	if total, pushers, err = GetPushersBySender(s.pusherRoot, sender, from, size); err != nil {
		return 10, nil
	}
	s.pushs(sender, pushers, data)
	for from = size; from < total; from = from + size {
		_, pushers, _ = GetPushersBySender(s.pusherRoot, sender, from, size)
		s.pushs(sender, pushers, data)
	}
	return 0, nil
}

func (s PushAllSender) pushs(sender string, pushers []pusherLib.Pusher, data string) {
	for _, pusher := range pushers {
		Push(s.pusherRoot, sender, pusher.ID, data)
	}
}
