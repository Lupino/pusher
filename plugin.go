package pusher

// Plugin defined pusher plugin
type Plugin interface {
	// GetGroupName for the periodic funcName
	GetGroupName() string
	// Do the job then return schedlater
	// if err != nil job fail
	// if schedlater > 0 job sched later
	// if schedlater == 0 job done
	Do(pusher, data string) (schedlater int, err error)
}
