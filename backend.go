package pusher

import (
	"github.com/Lupino/go-periodic"
	"github.com/blevesearch/bleve"
	"gopkg.in/redis.v3"
)

var redisClient *redis.Client
var periodicClient *periodic.Client

var index bleve.Index

// SetBackend server
func SetBackend(rc *redis.Client, pc *periodic.Client, idx bleve.Index) {
	redisClient = rc
	periodicClient = pc
	index = idx
}
