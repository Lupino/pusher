package worker

import (
	"encoding/json"
	"fmt"
	pusherLib "github.com/Lupino/pusher"
	"github.com/Lupino/pusher/utils"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// API of pusher server
type API struct {
	host   string
	key    string
	secret string
}

// GetPusher from api
func (api API) GetPusher(pusher string) (p pusherLib.Pusher, err error) {
	var rsp *http.Response
	var path = "/pusher/pushers/" + pusher + "/"
	var req, _ = http.NewRequest("GET", "http://"+api.host+path, nil)
	if len(api.key) > 0 {
		var signParams = make(map[string]string)
		signParams["path"] = path
		req.Header.Add("X-App-Key", api.key)
		signParams["app_key"] = api.key
		var timestamp = strconv.FormatInt(time.Now().Unix(), 10)
		req.Header.Add("X-Request-Time", timestamp)
		signParams["timestamp"] = timestamp
		var sign = utils.HmacMD5(api.secret, signParams)
		req.Header.Add("X-Request-Signature", sign)
	}
	if rsp, err = http.DefaultClient.Do(req); err != nil {
		log.Printf("http.DefaultClient.Do() failed (%s)", err)
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
func (api API) SearchPusher(q string, from, size int) (total int, pushers []pusherLib.Pusher, err error) {
	var rsp *http.Response
	var path = fmt.Sprintf("/pusher/search/")
	var query = url.Values{}
	query.Add("q", q)
	query.Add("from", strconv.Itoa(from))
	query.Add("size", strconv.Itoa(size))

	var url = fmt.Sprintf("http://%s%s?%s", api.host, path, query.Encode())

	var req, _ = http.NewRequest("GET", url, nil)
	if len(api.key) > 0 {
		var signParams = make(map[string]string)
		signParams["path"] = path
		req.Header.Add("X-App-Key", api.key)
		signParams["app_key"] = api.key
		var timestamp = strconv.FormatInt(time.Now().Unix(), 10)
		req.Header.Add("X-Request-Time", timestamp)
		signParams["timestamp"] = timestamp
		for key := range query {
			signParams[key] = query.Get(key)
		}
		var sign = utils.HmacMD5(api.secret, signParams)
		req.Header.Add("X-Request-Signature", sign)
	}
	if rsp, err = http.DefaultClient.Do(req); err != nil {
		log.Printf("http.DefaultClient.Do() failed (%s)", err)
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
func (api API) Push(sender, pusher, data string) (err error) {
	var rsp *http.Response
	var form = url.Values{}
	form.Set("pusher", pusher)
	form.Set("data", data)

	var path = fmt.Sprintf("/pusher/%s/push", sender)
	var url = fmt.Sprintf("http://%s%s", api.host, path)

	var req, _ = http.NewRequest("POST", url, strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if len(api.key) > 0 {
		var signParams = make(map[string]string)
		signParams["path"] = path
		req.Header.Add("X-App-Key", api.key)
		signParams["app_key"] = api.key
		var timestamp = strconv.FormatInt(time.Now().Unix(), 10)
		req.Header.Add("X-Request-Time", timestamp)
		signParams["timestamp"] = timestamp
		for key := range form {
			signParams[key] = form.Get(key)
		}
		var sign = utils.HmacMD5(api.secret, signParams)
		req.Header.Add("X-Request-Signature", sign)
	}
	if rsp, err = http.DefaultClient.Do(req); err != nil {
		log.Printf("http.DefaultClient.Do() failed (%s)", err)
		return
	}
	defer rsp.Body.Close()
	if int(rsp.StatusCode/100) != 2 {
		err = fmt.Errorf("push sender[%s] pusher[%s] failed", sender, pusher)
		return
	}
	return nil
}
