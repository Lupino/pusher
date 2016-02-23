package senders

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

var apiRoot = "http://gw.api.taobao.com/router/rest"

// SMSSender a alidayu sms sender
type SMSSender struct {
	appKey    string
	appSecret string
	signName  string
	template  string
}

// NewSMSSender new alidayu sms sender
func NewSMSSender(key, secret, signName, template string) SMSSender {
	return SMSSender{
		appKey:    key,
		appSecret: secret,
		signName:  signName,
		template:  template,
	}
}

type smsObject struct {
	PhoneNumber string `json:"phoneNumber"`
	Params      string `json:"params"`
}

// GetName for the periodic funcName
func (SMSSender) GetName() string {
	return "sendsms"
}

// Send message to pusher then return sendlater
func (s SMSSender) Send(pusher, data string) (int, error) {
	var (
		sms smsObject
		err error
	)
	if err = json.Unmarshal([]byte(data), &sms); err != nil {
		log.Printf("json.Unmarshal() failed (%s)", err)
		return 0, nil
	}
	if err = s.SendSMS(sms.PhoneNumber, sms.Params); err != nil {
		log.Printf("senders.SMSSender.SendSMS() failed (%s)", err)
		return 0, nil
	}
	return 0, nil
}

// SendSMS message
func (s SMSSender) SendSMS(phoneNumber, smsParams string) error {
	params := make(map[string]string)
	params["method"] = "alibaba.aliqin.fc.sms.num.send"
	params["app_key"] = s.appKey
	params["timestamp"] = time.Now().Format("2006-01-02 15:04:05")
	params["format"] = "json"
	params["v"] = "2.0"
	params["sign_method"] = "hmac"
	params["sms_type"] = "normal"
	params["sms_free_sign_name"] = s.signName
	params["rec_num"] = phoneNumber
	params["sms_param"] = smsParams
	params["sms_template_code"] = s.template
	params["sign"] = hmacMd5(s.appSecret, params)

	var form = url.Values{}
	for key, value := range params {
		form.Set(key, value)
	}

	rsp, err := http.PostForm(apiRoot, form)
	if err != nil {
		log.Printf("http.PostForm() failed (%s)", err)
		return err
	}
	defer rsp.Body.Close()

	decoder := json.NewDecoder(rsp.Body)
	var ret map[string]interface{}
	if err = decoder.Decode(&ret); err != nil {
		log.Printf("json.NewDecoder().Decode() failed (%s)", err)
		return err
	}
	errRsp, ok := ret["error_response"]
	if !ok {
		return nil
	}
	errRet, ok := errRsp.(map[string]string)
	if !ok {
		return fmt.Errorf("Unknow error")
	}
	return fmt.Errorf("%s", errRet["sub_code"])
}

func hmacMd5(slot string, params map[string]string) string {
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
