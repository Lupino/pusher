package pusher

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"strings"
)

// PREFIX the perfix key of pusher.
const PREFIX = "pusher:"
const secondIndex = "pusher:secondIndex"

func generateName(pusher, data string) string {
	mac := hmac.New(md5.New, []byte(pusher))
	mac.Write([]byte(data))
	sum := mac.Sum(nil)

	return pusher + "_" + hex.EncodeToString(sum)
}

func extractPusher(name string) string {
	idx := strings.LastIndex(name, "_")
	if idx == -1 {
		return ""
	}
	return name[:idx]
}

func verifyData(expect, pusher, data string) bool {
	got := generateName(pusher, data)
	return expect == got
}
