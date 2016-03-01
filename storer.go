package pusher

// Storer interface for store pusher data
type Storer interface {
	Set(Pusher) error
	Get(string) (Pusher, error)
	Del(string) error
	Search(string, int, int) (uint64, []Pusher, error)
	GetAll(from, size int) (uint64, []Pusher, error)
}
