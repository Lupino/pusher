package main

import (
	"flag"
	"github.com/Lupino/go-periodic"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/mholt/binding"
	"gopkg.in/redis.v3"
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

func push(group, pusher, data, schedat string) error {
	if !hasPusher(group, pusher) {
		log.Printf("pusher[%s] not in group[%s]", group, pusher)
		return nil
	}
	var opts = map[string]string{
		"args":    data,
		"schedat": schedat,
	}
	if err := periodicClient.SubmitJob(group, pusher, opts); err != nil {
		return err
	}
	return nil
}

func pushAll(group, data, schedat string) error {
	var pushers, _ = getPushers(group)
	var opts = map[string]string{
		"args":    data,
		"schedat": schedat,
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

type pushForm struct {
	Pusher  string
	Data    string
	SchedAt string
}

func (f *pushForm) FieldMap(_ *http.Request) binding.FieldMap {
	return binding.FieldMap{
		&f.Pusher: binding.Field{
			Form:     "pusher",
			Required: true,
		},
		&f.Data: binding.Field{
			Form:     "data",
			Required: true,
		},
		&f.SchedAt: binding.Field{
			Form:     "schedat",
			Required: true,
		},
	}
}

func handlePush(w http.ResponseWriter, req *http.Request) {
	f := new(pushForm)
	errs := binding.Bind(req, f)
	if errs.Handle(w) {
		return
	}

	vars := mux.Vars(req)
	group := vars["group"]
	var (
		bodyBytes []byte
		err       error
	)
	if err := push(group, f.Pusher, f.Data, f.SchedAt); err != nil {
		log.Printf("push() failed (%s)", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if _, err := w.Write([]byte("OK")); err != nil {
		log.Fatal(err)
	}
}

type pushAllForm struct {
	Data    string
	SchedAt string
}

func (f *pushAllForm) FieldMap(_ *http.Request) binding.FieldMap {
	return binding.FieldMap{
		&f.Data: binding.Field{
			Form:     "data",
			Required: true,
		},
		&f.SchedAt: binding.Field{
			Form:     "schedat",
			Required: true,
		},
	}
}

func handlePushAll(w http.ResponseWriter, req *http.Request) {
	f := new(pushAllForm)
	errs := binding.Bind(req, f)
	if errs.Handle(w) {
		return
	}

	vars := mux.Vars(req)
	group := vars["group"]
	if err := pushAll(group, f.Data, f.SchedAt); err != nil {
		log.Printf("pushAll() failed (%s)", err)
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
