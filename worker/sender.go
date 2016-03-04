package worker

// Sender interface for pusher
type Sender interface {
	// GetName for the periodic funcName
	GetName() string
	// Send message to pusher then return sendlater
	// if err != nil job fail
	// if sendlater > 0 send later
	// if sendlater == 0 send done
	Send(pusher, data string, counter int) (sendlater int, err error)
}
