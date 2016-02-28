package main

// Start Command eg : ./push 0 20000 localhost:6000 60

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	lg         *log.Logger
	httpClient *http.Client
	t          int
)

const testContent = "{\"subject\": \"subject\", \"text\": \"text\"}"

func init() {
	httpTransport := &http.Transport{
		Dial: func(netw, addr string) (net.Conn, error) {
			deadline := time.Now().Add(30 * time.Second)
			c, err := net.DialTimeout(netw, addr, 20*time.Second)
			if err != nil {
				return nil, err
			}

			c.SetDeadline(deadline)
			return c, nil
		},
		DisableKeepAlives: false,
	}
	httpClient = &http.Client{
		Transport: httpTransport,
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	infoLogfi, err := os.OpenFile("./push.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	lg = log.New(infoLogfi, "", log.LstdFlags|log.Lshortfile)

	begin, err := strconv.Atoi(os.Args[1])
	if err != nil {
		panic(err)
	}
	length, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic(err)
	}

	t, err = strconv.Atoi(os.Args[4])
	if err != nil {
		panic(err)
	}

	setup()

	num := runtime.NumCPU() * 8
	lg.Printf("start routine num:%d", num)

	l := length / num
	b, e := begin, begin+l
	time.AfterFunc(time.Duration(t)*time.Second, stop)
	for i := 0; i < num; i++ {
		go startPush(b, e)
		b += l
		e += l
	}
	if b < begin+length {
		go startPush(b, begin+length)
	}

	time.Sleep(9999 * time.Hour)
}

func stop() {
	os.Exit(-1)
}

func startPush(b, e int) {
	lg.Printf("start Push from %d to %d", b, e)
	bodys := make([]url.Values, e-b)
	for i := 0; i < e-b; i++ {
		body := url.Values{}
		body.Set("data", testContent+strconv.Itoa(b))
		body.Set("pusher", "lupino")
		bodys[i] = body
	}

	for {
		for i := 0; i < len(bodys); i++ {
			resp, err := httpPost(fmt.Sprintf("http://%s/pusher/sendmail/push", os.Args[3]), "application/x-www-form-urlencoded", strings.NewReader(bodys[i].Encode()))
			if err != nil {
				lg.Printf("post error (%v)", err)
				continue
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				lg.Printf("post error (%v)", err)
				return
			}
			resp.Body.Close()

			lg.Printf("response %s", string(body))
			//time.Sleep(50 * time.Millisecond)
		}
	}
}

func setup() {
	var form = url.Values{}
	form.Set("pusher", "lupino")
	form.Set("email", "example.com")
	_, err := httpPost(fmt.Sprintf("http://%s/pusher/pushers/", os.Args[3]), "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		lg.Printf("post error (%v)", err)
	}

	var form2 = url.Values{}
	form2.Set("pusher", "lupino")
	_, err = httpPost(fmt.Sprintf("http://%s/pusher/sendmail/add", os.Args[3]), "application/x-www-form-urlencoded", strings.NewReader(form2.Encode()))
	if err != nil {
		lg.Printf("post error (%v)", err)
	}
}

func httpPost(url string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
