package client

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

// PusherClient for pusher server
type PusherClient struct {
	host   string
	key    string
	secret string
}

// New create new pusher client
func New(host, key, secret string) PusherClient {
	return PusherClient{host: host, key: key, secret: secret}
}

func (client PusherClient) signParams(req *http.Request, path string, params url.Values) {
	var signParams = make(map[string]string)
	signParams["path"] = path
	req.Header.Add("X-App-Key", client.key)
	signParams["app_key"] = client.key
	var timestamp = strconv.FormatInt(time.Now().Unix(), 10)
	req.Header.Add("X-Request-Time", timestamp)
	signParams["timestamp"] = timestamp
	for key := range params {
		signParams[key] = params.Get(key)
	}
	var sign = utils.HmacMD5(client.secret, signParams)
	req.Header.Add("X-Request-Signature", sign)
}

func (client PusherClient) signPath(req *http.Request, path string) {
	client.signParams(req, path, url.Values{})
}

// GetPusher from client
func (client PusherClient) GetPusher(pusher string) (p pusherLib.Pusher, err error) {
	var rsp *http.Response
	var path = "/pusher/pushers/" + pusher + "/"
	var req, _ = http.NewRequest("GET", "http://"+client.host+path, nil)
	if len(client.key) > 0 {
		client.signPath(req, path)
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

type getAllPusherResult struct {
	Pushers []pusherLib.Pusher `json:"pushers"`
	From    int                `json:"from"`
	Size    int                `json:"size"`
	Total   int                `json:"total"`
}

// GetPusherList from client
func (client PusherClient) GetPusherList(from, size int) (total int, pushers []pusherLib.Pusher, err error) {
	var rsp *http.Response
	var path = "/pusher/pushers/"
	var query = url.Values{}
	query.Add("from", strconv.Itoa(from))
	query.Add("size", strconv.Itoa(size))

	var url = fmt.Sprintf("http://%s%s?%s", client.host, path, query.Encode())

	var req, _ = http.NewRequest("GET", url, nil)
	if len(client.key) > 0 {
		client.signParams(req, path, query)
	}
	if rsp, err = http.DefaultClient.Do(req); err != nil {
		log.Printf("http.DefaultClient.Do() failed (%s)", err)
		return
	}
	defer rsp.Body.Close()
	if int(rsp.StatusCode/100) != 2 {
		err = fmt.Errorf("get all pusher list failed")
		return
	}
	var ret getAllPusherResult
	decoder := json.NewDecoder(rsp.Body)
	if err = decoder.Decode(&ret); err != nil {
		log.Printf("json.NewDecoder().Decode() failed (%s)", err)
		return
	}
	return ret.Total, ret.Pushers, nil
}

type searchPusherResult struct {
	Pushers []pusherLib.Pusher `json:"pushers"`
	From    int                `json:"from"`
	Size    int                `json:"size"`
	Total   int                `json:"total"`
	Q       string             `json:"q"`
}

// SearchPusher from client
func (client PusherClient) SearchPusher(q string, from, size int) (total int, pushers []pusherLib.Pusher, err error) {
	var rsp *http.Response
	var path = "/pusher/search/"
	var query = url.Values{}
	query.Add("q", q)
	query.Add("from", strconv.Itoa(from))
	query.Add("size", strconv.Itoa(size))

	var url = fmt.Sprintf("http://%s%s?%s", client.host, path, query.Encode())

	var req, _ = http.NewRequest("GET", url, nil)
	if len(client.key) > 0 {
		client.signParams(req, path, query)
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

// AddPusher create a new pusher
func (client PusherClient) AddPusher(pusher pusherLib.Pusher) (err error) {
	var rsp *http.Response
	var path = "/pusher/pushers/"
	var form = url.Values{}
	form.Add("pusher", pusher.ID)
	form.Add("email", pusher.Email)
	form.Add("nickname", pusher.NickName)
	form.Add("phoneNumber", pusher.PhoneNumber)
	form.Add("CreatedAt", strconv.FormatInt(pusher.CreatedAt, 10))

	var url = fmt.Sprintf("http://%s%s", client.host, path)

	var req, _ = http.NewRequest("POST", url, strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if len(client.key) > 0 {
		client.signParams(req, path, form)
	}
	if rsp, err = http.DefaultClient.Do(req); err != nil {
		log.Printf("http.DefaultClient.Do() failed (%s)", err)
		return
	}
	defer rsp.Body.Close()
	if int(rsp.StatusCode/100) != 2 {
		err = fmt.Errorf("create pusher failed")
		return
	}
	return nil
}

// RemovePusher remove an exists pusher
func (client PusherClient) RemovePusher(pusher string) (err error) {
	var rsp *http.Response
	var path = fmt.Sprintf("/pusher/pushers/%s/", pusher)
	var url = fmt.Sprintf("http://%s%s", client.host, path)

	var req, _ = http.NewRequest("DELETE", url, nil)
	if len(client.key) > 0 {
		client.signPath(req, path)
	}
	if rsp, err = http.DefaultClient.Do(req); err != nil {
		log.Printf("http.DefaultClient.Do() failed (%s)", err)
		return
	}
	defer rsp.Body.Close()
	if int(rsp.StatusCode/100) != 2 {
		err = fmt.Errorf("remove pusher (%s) failed", pusher)
		return
	}
	return nil
}

// UpdatePusher update an exists pusher
func (client PusherClient) UpdatePusher(pusher string, updated map[string]string) (err error) {
	var rsp *http.Response
	var path = fmt.Sprintf("/pusher/pushers/%s/", pusher)
	var form = url.Values{}
	for k, v := range updated {
		form.Add(k, v)
	}

	var url = fmt.Sprintf("http://%s%s", client.host, path)

	var req, _ = http.NewRequest("POST", url, strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if len(client.key) > 0 {
		client.signParams(req, path, form)
	}
	if rsp, err = http.DefaultClient.Do(req); err != nil {
		log.Printf("http.DefaultClient.Do() failed (%s)", err)
		return
	}
	defer rsp.Body.Close()
	if int(rsp.StatusCode/100) != 2 {
		err = fmt.Errorf("update pusher (%s) failed", pusher)
		return
	}
	return nil
}

// RemoveTag remove an exists pusher tag
func (client PusherClient) RemoveTag(pusher, tag string) (err error) {
	var rsp *http.Response
	var path = fmt.Sprintf("/pusher/pushers/%s/%s/", pusher, tag)
	var url = fmt.Sprintf("http://%s%s", client.host, path)

	var req, _ = http.NewRequest("DELETE", url, nil)
	if len(client.key) > 0 {
		client.signPath(req, path)
	}
	if rsp, err = http.DefaultClient.Do(req); err != nil {
		log.Printf("http.DefaultClient.Do() failed (%s)", err)
		return
	}
	defer rsp.Body.Close()
	if int(rsp.StatusCode/100) != 2 {
		err = fmt.Errorf("remove pusher (%s) tag (%s) failed", pusher, tag)
		return
	}
	return nil
}

// AddTag add a tag to an exists pusher
func (client PusherClient) AddTag(pusher, tag string) (err error) {
	var rsp *http.Response
	var path = fmt.Sprintf("/pusher/pushers/%s/%s/", pusher, tag)
	var url = fmt.Sprintf("http://%s%s", client.host, path)

	var req, _ = http.NewRequest("POST", url, nil)
	if len(client.key) > 0 {
		client.signPath(req, path)
	}
	if rsp, err = http.DefaultClient.Do(req); err != nil {
		log.Printf("http.DefaultClient.Do() failed (%s)", err)
		return
	}
	defer rsp.Body.Close()
	if int(rsp.StatusCode/100) != 2 {
		err = fmt.Errorf("remove pusher (%s) tag (%s) failed", pusher, tag)
		return
	}
	return nil
}

// AddSender add a sender to an exists pusher
func (client PusherClient) AddSender(pusher, sender string) (err error) {
	var rsp *http.Response
	var path = fmt.Sprintf("/pusher/%s/add", sender)

	var form = url.Values{}
	form.Add("pusher", pusher)

	var url = fmt.Sprintf("http://%s%s", client.host, path)

	var req, _ = http.NewRequest("POST", url, strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if len(client.key) > 0 {
		client.signParams(req, path, form)
	}
	if rsp, err = http.DefaultClient.Do(req); err != nil {
		log.Printf("http.DefaultClient.Do() failed (%s)", err)
		return
	}
	defer rsp.Body.Close()
	if int(rsp.StatusCode/100) != 2 {
		err = fmt.Errorf("remove pusher (%s) sender (%s) failed", pusher, sender)
		return
	}
	return nil
}

// RemoveSender remove sender from an exists pusher
func (client PusherClient) RemoveSender(pusher, sender string) (err error) {
	var rsp *http.Response
	var path = fmt.Sprintf("/pusher/%s/remove", sender)

	var form = url.Values{}
	form.Add("pusher", pusher)

	var url = fmt.Sprintf("http://%s%s", client.host, path)

	var req, _ = http.NewRequest("POST", url, strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if len(client.key) > 0 {
		client.signParams(req, path, form)
	}
	if rsp, err = http.DefaultClient.Do(req); err != nil {
		log.Printf("http.DefaultClient.Do() failed (%s)", err)
		return
	}
	defer rsp.Body.Close()
	if int(rsp.StatusCode/100) != 2 {
		err = fmt.Errorf("remove pusher (%s) sender (%s) failed", pusher, sender)
		return
	}
	return nil
}

type pushResult struct {
	Name   string `json:"name"`
	Result string `json:"result"`
}

// Push message to pusher server by client
func (client PusherClient) Push(sender, pusher, data, schedat string) (name string, err error) {
	var rsp *http.Response
	var form = url.Values{}
	form.Set("pusher", pusher)
	form.Set("data", data)
	form.Set("schedat", schedat)

	var path = fmt.Sprintf("/pusher/%s/push", sender)
	var url = fmt.Sprintf("http://%s%s", client.host, path)

	var req, _ = http.NewRequest("POST", url, strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if len(client.key) > 0 {
		client.signParams(req, path, form)
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
	var ret pushResult
	decoder := json.NewDecoder(rsp.Body)
	if err = decoder.Decode(&ret); err != nil {
		log.Printf("json.NewDecoder().Decode() failed (%s)", err)
		return
	}
	return ret.Name, nil
}

// CancelPush cancel push by a push name
func (client PusherClient) CancelPush(sender, name string) (err error) {
	var rsp *http.Response
	var form = url.Values{}
	form.Set("name", name)

	var path = fmt.Sprintf("/pusher/%s/cancelpush", sender)
	var url = fmt.Sprintf("http://%s%s", client.host, path)

	var req, _ = http.NewRequest("POST", url, strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if len(client.key) > 0 {
		client.signParams(req, path, form)
	}
	if rsp, err = http.DefaultClient.Do(req); err != nil {
		log.Printf("http.DefaultClient.Do() failed (%s)", err)
		return
	}
	defer rsp.Body.Close()
	if int(rsp.StatusCode/100) != 2 {
		err = fmt.Errorf("cancel push sender[%s] push (%s) failed", sender, name)
		return
	}
	return nil
}

// PushAll message to pusher server by client
func (client PusherClient) PushAll(sender, data, tag, schedat string) (name string, err error) {
	var rsp *http.Response
	var form = url.Values{}
	form.Set("tag", tag)
	form.Set("data", data)
	form.Set("schedat", schedat)

	var path = fmt.Sprintf("/pusher/%s/pushall", sender)
	var url = fmt.Sprintf("http://%s%s", client.host, path)

	var req, _ = http.NewRequest("POST", url, strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if len(client.key) > 0 {
		client.signParams(req, path, form)
	}
	if rsp, err = http.DefaultClient.Do(req); err != nil {
		log.Printf("http.DefaultClient.Do() failed (%s)", err)
		return
	}
	defer rsp.Body.Close()
	if int(rsp.StatusCode/100) != 2 {
		err = fmt.Errorf("pushall sender(%s) tag (%s) failed", sender)
		return
	}
	var ret pushResult
	decoder := json.NewDecoder(rsp.Body)
	if err = decoder.Decode(&ret); err != nil {
		log.Printf("json.NewDecoder().Decode() failed (%s)", err)
		return
	}
	return ret.Name, nil
}
