package senders

import (
	"encoding/json"
	"fmt"
	pusherLib "github.com/Lupino/pusher"
	"log"
	"net/http"
	"net/url"
)

// GetPusher from api
func GetPusher(host, pusher string) (p pusherLib.Pusher, err error) {
	var rsp *http.Response
	if rsp, err = http.Get(host + "/pusher/pushers/" + pusher + "/"); err != nil {
		log.Printf("http.Get() failed (%s)", err)
		return
	}
	defer rsp.Body.Close()
	if int(rsp.StatusCode/100) != 2 {
		err = fmt.Errorf("pusher[%s] not exists", pusher)
		return
	}
	var ret map[string]pusherLib.Pusher
	decoder := json.NewDecoder(rsp.Body)
	if err = decoder.Decode(&ret); err != nil {
		log.Printf("json.NewDecoder().Decode() failed (%s)", err)
		return
	}
	var ok bool
	if p, ok = ret["pusher"]; !ok {
		err = fmt.Errorf("pusher[%s] not exists", pusher)
		return
	}
	return
}

type getPushersBySenderResult struct {
	Pushers []pusherLib.Pusher `json:"pushers"`
	From    int                `json:"from"`
	Size    int                `json:"size"`
	Total   int                `json:"total"`
	Sender  string             `json:"sender"`
}

// GetPushersBySender from api
func GetPushersBySender(host, sender string, from, size int) (total int, pushers []pusherLib.Pusher, err error) {
	var rsp *http.Response
	var url = fmt.Sprintf("%s/pusher/%s/pushers/?from=%d&size=%d", host, sender, from, size)
	if rsp, err = http.Get(url); err != nil {
		log.Printf("http.Get() failed (%s)", err)
		return
	}
	defer rsp.Body.Close()
	if int(rsp.StatusCode/100) != 2 {
		err = fmt.Errorf("load sender[%s] failed", sender)
		return
	}
	var ret getPushersBySenderResult
	decoder := json.NewDecoder(rsp.Body)
	if err = decoder.Decode(&ret); err != nil {
		log.Printf("json.NewDecoder().Decode() failed (%s)", err)
		return
	}
	return ret.Total, ret.Pushers, nil
}
