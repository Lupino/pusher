package senders

import (
	"fmt"
	"github.com/Lupino/pusher/utils"
	"github.com/Lupino/pusher/worker"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// HookSender a hook sender
type HookSender struct {
	name   string
	url    string
	secret string
	w      worker.Worker
}

// NewHookSender new hook sender
func NewHookSender(w worker.Worker, name, url, secret string) HookSender {
	return HookSender{
		name:   name,
		url:    url,
		secret: secret,
		w:      w,
	}
}

// GetName for the periodic funcName
func (s HookSender) GetName() string {
	return s.name
}

// Send message to pusher then return sendlater
func (s HookSender) Send(pusher, data string) (int, error) {
	var (
		rsp        *http.Response
		err        error
		form       = url.Values{}
		signParams = make(map[string]string)
		timestamp  = strconv.FormatInt(time.Now().Unix(), 10)
		req        *http.Request
		sign       string
	)

	form.Set("sender", s.name)
	form.Set("pusher", pusher)
	form.Set("data", data)

	req, _ = http.NewRequest("POST", s.url, strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("X-Request-Time", timestamp)

	signParams["timestamp"] = timestamp

	for key := range form {
		signParams[key] = form.Get(key)
	}

	sign = utils.HmacMD5(s.secret, signParams)
	req.Header.Add("X-Request-Signature", sign)

	if rsp, err = http.DefaultClient.Do(req); err != nil {
		log.Printf("http.DefaultClient.Do() failed (%s)", err)
		return 0, nil
	}
	defer rsp.Body.Close()
	if int(rsp.StatusCode/100) != 2 {
		err = fmt.Errorf("hook[%s] sender to pusher[%s] failed", s.name, pusher)
		return 0, nil
	}
	return 0, nil
}
