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
	if rsp, err = http.Get("http://" + host + "/pusher/pushers/" + pusher + "/"); err != nil {
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

type searchPusherResult struct {
	Pushers []pusherLib.Pusher `json:"pushers"`
	From    int                `json:"from"`
	Size    int                `json:"size"`
	Total   int                `json:"total"`
	Q       string             `json:"q"`
}

// SearchPusher from api
func SearchPusher(host, q string, from, size int) (total int, pushers []pusherLib.Pusher, err error) {
	var rsp *http.Response
	var url = fmt.Sprintf("http://%s/pusher/search/?q=%s&from=%d&size=%d", host, q, from, size)
	if rsp, err = http.Get(url); err != nil {
		log.Printf("http.Get() failed (%s)", err)
		return
	}
	defer rsp.Body.Close()
	if int(rsp.StatusCode/100) != 2 {
		err = fmt.Errorf("search pusher [%s] failed", q)
		return
	}
	var ret searchPusherResult
	decoder := json.NewDecoder(rsp.Body)
	if err = decoder.Decode(&ret); err != nil {
		log.Printf("json.NewDecoder().Decode() failed (%s)", err)
		return
	}
	return ret.Total, ret.Pushers, nil
}

// Push message to pusher server by api
func Push(host, sender, pusher, data string) (err error) {
	var rsp *http.Response
	var form = url.Values{}
	form.Set("pusher", pusher)
	form.Set("data", data)
	if rsp, err = http.PostForm("http://"+host+"/pusher/"+sender+"/push", form); err != nil {
		log.Printf("http.PostForm() failed (%s)", err)
		return
	}
	defer rsp.Body.Close()
	if int(rsp.StatusCode/100) != 2 {
		err = fmt.Errorf("push sender[%s] pusher[%s] failed", sender, pusher)
		return
	}
	return nil
}
