package pusher

import (
	"github.com/gorilla/mux"
	"github.com/mholt/binding"
	"log"
	"net/http"
)

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

func push(group, pusher, data, schedat string, force bool) error {
	if !force && !hasPusher(group, pusher) {
		log.Printf("pusher[%s] not in group[%s]", group, pusher)
		return nil
	}
	var opts = map[string]string{
		"args":    data,
		"schedat": schedat,
	}
	if err := periodicClient.SubmitJob(PREFIX+group, generateName(pusher, data), opts); err != nil {
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
		periodicClient.SubmitJob(PREFIX+group, generateName(pusher, data), opts)
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
	Force   bool
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
		&f.Force: binding.Field{
			Form:     "force",
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
	if err := push(group, f.Pusher, f.Data, f.SchedAt, f.Force); err != nil {
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

// NewRouter return new pusher router
func NewRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/pusher/{group}/add", handleAddPusher).Methods("POST")
	router.HandleFunc("/pusher/{group}/delete", handleRemovePusher).Methods("POST")
	router.HandleFunc("/pusher/{group}/push", handlePush).Methods("POST")
	router.HandleFunc("/pusher/{group}/pushall", handlePushAll).Methods("POST")
	return router
}
