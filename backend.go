package pusher

import (
	"github.com/Lupino/go-periodic"
	"gopkg.in/redis.v3"
)

var redisClient *redis.Client
var periodicClient *periodic.Client

// SetBackend server
func SetBackend(rc *redis.Client, pc *periodic.Client) {
	redisClient = rc
	periodicClient = pc
}
