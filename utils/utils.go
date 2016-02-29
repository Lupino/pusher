package utils

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"strings"
)

// GenerateName for periodic job name base pusher and push data.
func GenerateName(pusher, data string) string {
	mac := hmac.New(md5.New, []byte(pusher))
	mac.Write([]byte(data))
	sum := mac.Sum(nil)

	return pusher + "_" + hex.EncodeToString(sum)
}

// ExtractPusher from periodic job name.
func ExtractPusher(name string) string {
	idx := strings.LastIndex(name, "_")
	if idx == -1 {
		return ""
	}
	return name[:idx]
}

// VerifyData where is the same with except name.
func VerifyData(expect, pusher, data string) bool {
	got := GenerateName(pusher, data)
	return expect == got
}
