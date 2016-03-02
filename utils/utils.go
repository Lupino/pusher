package utils

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"sort"
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

// HmacMD5 sign pusher request
func HmacMD5(slot string, params map[string]string) string {
	mac := hmac.New(md5.New, []byte(slot))
	var keys []string
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		mac.Write([]byte(key))
		mac.Write([]byte(params[key]))
	}
	sum := mac.Sum(nil)

	return strings.ToUpper(hex.EncodeToString(sum))
}
