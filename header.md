Pusher Object
=============

```go
type Pusher struct {
	ID          string   `json:"id"`
	Email       string   `json:"email"`
	NickName    string   `json:"nickname"`
	PhoneNumber string   `json:"phoneNumber"`
	Senders     []string `json:"senders"`
	Tags        []string `json:"tags"`
	CreatedAt   int64    `json:"createdAt"`
}
```
