package main

import (
	"flag"
	"github.com/Lupino/go-periodic"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"gopkg.in/redis.v3"
	"io/ioutil"
	"log"
	"net/http"
)

var periodicPort string
var redisHost string
var redisClient *redis.Client
var periodicClient *periodic.Client

// PREFIX the perfix key of pusher.
var PREFIX = "pusher:"

func init() {
	flag.StringVar(&periodicPort, "periodic_port", "unix:///tmp/periodic.sock", "the periodic server port.")
	flag.StringVar(&redisHost, "redis_host", "localhost:6379", "the redis server host.")
	flag.Parse()

	redisClient = redis.NewClient(&redis.Options{
		Addr: redisHost,
	})

	if err := redisClient.Ping().Err(); err != nil {
		log.Fatal(err)
	}

	periodicClient = periodic.NewClient()
	if err := periodicClient.Connect(periodicPort); err != nil {
		log.Fatal(err)
	}
}

func addPusher(group string, pusher ...string) error {
	return redisClient.SAdd(PREFIX+group, pusher...).Err()
}

func removePusher(group string, pusher ...string) error {
	return redisClient.SRem(PREFIX+group, pusher...).Err()
}

func hasPusher(group, pusher string) bool {
	return redisClient.SIsMember(PREFIX+group, pusher).Val()
}

func getPushers(group string) ([]string, error) {
	return redisClient.SMembers(PREFIX + group).Result()
}

func push(group, pusher, body string) error {
	if !hasPusher(group, pusher) {
		log.Printf("pusher[%s] not in group[%s]", group, pusher)
		return nil
	}
	var opts = map[string]string{
		"args": body,
	}
	if err := periodicClient.SubmitJob(group, pusher, opts); err != nil {
		return err
	}
	return nil
}

func pushAll(group, body string) error {
	var pushers, _ = getPushers(group)
	var opts = map[string]string{
		"args": body,
	}
	for _, pusher := range pushers {
		periodicClient.SubmitJob(group, pusher, opts)
	}
	return nil
}

func handleAddPusher(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	pusher := req.Form.Get("pusher")
	vars := mux.Vars(req)
	group := vars["group"]
	if err := addPusher(group, pusher); err != nil {
		log.Printf("addPusher() failed (%s)", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if _, err := w.Write([]byte("OK")); err != nil {
		log.Fatal(err)
	}
}

func handleRemovePusher(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	pusher := req.Form.Get("pusher")
	vars := mux.Vars(req)
	group := vars["group"]
	if err := removePusher(group, pusher); err != nil {
		log.Printf("removePusher() failed (%s)", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if _, err := w.Write([]byte("OK")); err != nil {
		log.Fatal(err)
	}
}

func handlePush(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	pusher := req.Form.Get("pusher")
	vars := mux.Vars(req)
	group := vars["group"]
	var (
		bodyBytes []byte
		err       error
	)
	if bodyBytes, err = ioutil.ReadAll(req.Body); err != nil {
		log.Printf("ioutil.ReadAll() failed (%s)", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err := push(group, pusher, string(bodyBytes)); err != nil {
		log.Printf("push() failed (%s)", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if _, err := w.Write([]byte("OK")); err != nil {
		log.Fatal(err)
	}
}

func handlePushAll(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	group := vars["group"]
	var (
		bodyBytes []byte
		err       error
	)
	if bodyBytes, err = ioutil.ReadAll(req.Body); err != nil {
		log.Printf("ioutil.ReadAll() failed (%s)", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err := pushAll(group, string(bodyBytes)); err != nil {
		log.Printf("push() failed (%s)", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if _, err := w.Write([]byte("OK")); err != nil {
		log.Fatal(err)
	}
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/pusher/{group}/add", handleAddPusher).Methods("POST")
	router.HandleFunc("/pusher/{group}/delete", handleRemovePusher).Methods("POST")
	router.HandleFunc("/pusher/{group}/push", handlePush).Methods("POST")
	router.HandleFunc("/pusher/{group}/pushall", handlePushAll).Methods("POST")

	n := negroni.New(negroni.NewRecovery(), negroni.NewLogger())
	n.UseHandler(router)
	n.Run(":3000")
}
