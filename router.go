package pusher

import (
	"github.com/blevesearch/bleve"
	"github.com/gorilla/mux"
	"github.com/mholt/binding"
	"github.com/unrolled/render"
	"log"
	"net/http"
	"strconv"
)

var r = render.New()

func sendJSONResponse(w http.ResponseWriter, status int, key string, data interface{}) {
	if key == "" {
		r.JSON(w, status, data)
	} else {
		r.JSON(w, status, map[string]interface{}{key: data})
	}
}

func addSender(pusher string, senders ...string) (err error) {
	var p Pusher
	if p, err = GetPusher(pusher); err != nil {
		return
	}
	changed := false
	for _, sender := range senders {
		if p.AddSender(sender) {
			changed = true
		}
	}

	if changed {
		if err = SetPusher(p); err != nil {
			return
		}
	}
	return
}

func removeSender(pusher string, senders ...string) (err error) {
	var p Pusher
	if p, err = GetPusher(pusher); err != nil {
		return
	}
	changed := false
	for _, sender := range senders {
		if p.DelSender(sender) {
			changed = true
		}
	}

	if changed {
		if err = SetPusher(p); err != nil {
			return
		}
	}
	return
}

func hasSender(pusher, sender string) bool {
	var p Pusher
	var err error
	if p, err = GetPusher(pusher); err != nil {
		return false
	}
	return p.HasSender(sender)
}

func searchPushers(sender string, size, from int) (uint64, []string, error) {
	query := bleve.NewQueryStringQuery(sender)
	searchRequest := bleve.NewSearchRequestOptions(query, size, from, false)
	searchRequest.Fields = []string{"senders"}
	searchResult, err := index.Search(searchRequest)
	if err != nil {
		log.Printf("bleve.Index.Search() failed(%s)", err)
		return 0, nil, err
	}

	var pushers []string
	for _, hit := range searchResult.Hits {
		if values, ok := hit.Fields["senders"].([]interface{}); ok {
			for _, s := range values {
				if ss, ok := s.(string); ok {
					if ss == sender {
						pushers = append(pushers, hit.ID)
					}
				}
			}
		} else {
			value, ok := hit.Fields["senders"].(string)
			if ok && value == sender {
				pushers = append(pushers, hit.ID)
			}
		}
	}

	return searchResult.Total, pushers, nil
}

func getAllPushersBySender(sender string) ([]string, error) {
	var total uint64
	var pushers []string
	var err error
	if total, _, err = searchPushers(sender, 2, 0); err != nil {
		return nil, err
	}

	_, pushers, err = searchPushers(sender, int(total), 0)
	return pushers, err
}

func push(sender, pusher, data, schedat string, force bool) error {
	if !hasSender(pusher, sender) {
		log.Printf("pusher[%s] not has sender[%s]", pusher, sender)
		if !force {
			return nil
		}
	}
	var opts = map[string]string{
		"args":    data,
		"schedat": schedat,
	}
	if err := periodicClient.SubmitJob(PREFIX+sender, generateName(pusher, data), opts); err != nil {
		return err
	}
	return nil
}

func pushAll(sender, data, schedat string) error {
	var pushers, _ = getAllPushersBySender(sender)
	var opts = map[string]string{
		"args":    data,
		"schedat": schedat,
	}
	for _, pusher := range pushers {
		periodicClient.SubmitJob(PREFIX+sender, generateName(pusher, data), opts)
	}
	return nil
}

func handleAddSender(w http.ResponseWriter, req *http.Request, sender string) {
	req.ParseForm()
	pusher := req.Form.Get("pusher")
	if err := addSender(pusher, sender); err != nil {
		log.Printf("addSender() failed (%s)", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	sendJSONResponse(w, http.StatusOK, "result", "OK")
}

func handleRemoveSender(w http.ResponseWriter, req *http.Request, sender string) {
	req.ParseForm()
	pusher := req.Form.Get("pusher")
	if err := removeSender(pusher, sender); err != nil {
		log.Printf("removeSender() failed (%s)", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	sendJSONResponse(w, http.StatusOK, "result", "OK")
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
			Required: false,
		},
		&f.Force: binding.Field{
			Form:     "force",
			Required: false,
		},
	}
}

func handlePush(w http.ResponseWriter, req *http.Request, sender string) {
	f := new(pushForm)
	errs := binding.Bind(req, f)
	if errs.Handle(w) {
		return
	}

	if err := push(sender, f.Pusher, f.Data, f.SchedAt, f.Force); err != nil {
		log.Printf("push() failed (%s)", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	sendJSONResponse(w, http.StatusOK, "result", "OK")
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
			Required: false,
		},
	}
}

func handlePushAll(w http.ResponseWriter, req *http.Request, sender string) {
	f := new(pushAllForm)
	errs := binding.Bind(req, f)
	if errs.Handle(w) {
		return
	}

	if err := pushAll(sender, f.Data, f.SchedAt); err != nil {
		log.Printf("pushAll() failed (%s)", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, http.StatusOK, "result", "OK")
}

func wapperSenderHandle(handle func(http.ResponseWriter, *http.Request, string)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		sender := vars["sender"]
		handle(w, req, sender)
	}
}

func handleGetPusher(w http.ResponseWriter, req *http.Request, pusher string) {
	p, err := GetPusher(pusher)
	if err != nil {
		log.Printf("GetPusher() failed (%s)", err)
	}
	if p.ID != pusher {
		sendJSONResponse(w, http.StatusNotFound, "err", "pusher "+pusher+" not exists.")
		return
	}
	sendJSONResponse(w, http.StatusOK, "pusher", p)
}

func handleGetAllPusher(w http.ResponseWriter, req *http.Request) {
	var qs = req.URL.Query()
	var err error
	var from, size int
	if from, err = strconv.Atoi(qs.Get("from")); err != nil {
		from = 0
	}

	if size, err = strconv.Atoi(qs.Get("size")); err != nil {
		size = 10
	}

	if size > 100 {
		size = 100
	}

	var stop = from + size

	ret := redisClient.ZRange(secondIndex, int64(from), int64(stop))
	if ret.Err() != nil {
		log.Printf("redis.Client.ZRange() failed (%s)", ret.Err())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	var pushers []Pusher
	for _, pusher := range ret.Val() {
		p, _ := GetPusher(pusher)
		if p.ID == pusher {
			pushers = append(pushers, p)
		}
	}

	total := redisClient.ZCard(secondIndex).Val()

	sendJSONResponse(w, http.StatusOK, "", map[string]interface{}{
		"pushers": pushers,
		"total":   total,
		"from":    from,
		"size":    size,
	})
}

func handleSearchPusher(w http.ResponseWriter, req *http.Request) {
	var qs = req.URL.Query()
	var err error
	var from, size int
	var total uint64
	var pushers []Pusher
	var q = qs.Get("q")
	if from, err = strconv.Atoi(qs.Get("from")); err != nil {
		from = 0
	}

	if size, err = strconv.Atoi(qs.Get("size")); err != nil {
		size = 10
	}

	if size > 100 {
		size = 100
	}

	if q == "" {
		sendJSONResponse(w, http.StatusOK, "err", "q is required.")
		return
	}

	query := bleve.NewQueryStringQuery(q)
	searchRequest := bleve.NewSearchRequestOptions(query, size, from, false)
	searchResult, err := index.Search(searchRequest)
	if err != nil {
		log.Printf("bleve.Index.Search() failed(%s)", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	for _, hit := range searchResult.Hits {
		p, _ := GetPusher(hit.ID)
		if p.ID == hit.ID {
			pushers = append(pushers, p)
		}
	}

	total = searchResult.Total

	sendJSONResponse(w, http.StatusOK, "", map[string]interface{}{
		"pushers": pushers,
		"total":   total,
		"from":    from,
		"size":    size,
		"q":       q,
	})
}

func handleAddPusher(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	var p = Pusher{}
	p.ID = req.Form.Get("pusher")
	p.Email = req.Form.Get("email")
	p.RealName = req.Form.Get("realname")
	p.NickName = req.Form.Get("nickname")
	p.PhoneNumber = req.Form.Get("phoneNumber")
	p.CreatedAt, _ = strconv.ParseInt(req.Form.Get("createdAt"), 10, 64)
	if p.ID == "" {
		sendJSONResponse(w, http.StatusOK, "err", "pusher is required.")
		return
	}

	if err := SetPusher(p); err != nil {
		log.Printf("SetPusher() failed(%s)", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	sendJSONResponse(w, http.StatusOK, "result", "OK")
}

func handleRemovePusher(w http.ResponseWriter, req *http.Request, pusher string) {
	if err := DelPusher(pusher); err != nil {
		log.Printf("DelPusher() failed(%s)", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	sendJSONResponse(w, http.StatusOK, "result", "OK")
}

func handleUpdatePusher(w http.ResponseWriter, req *http.Request, pusher string) {
	var p Pusher
	var err error
	if p, err = GetPusher(pusher); err != nil {
		log.Printf("DelPusher() failed(%s)", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if p.ID != pusher {
		sendJSONResponse(w, http.StatusNotFound, "err", "pusher "+pusher+" not exists.")
		return
	}

	if req.Form.Get("email") != "" {
		p.Email = req.Form.Get("email")
	}
	if req.Form.Get("realname") != "" {
		p.RealName = req.Form.Get("realname")
	}
	if req.Form.Get("nickname") != "" {
		p.NickName = req.Form.Get("nickname")
	}
	if req.Form.Get("phoneNumber") != "" {
		p.PhoneNumber = req.Form.Get("phoneNumber")
	}
	if req.Form.Get("createdAt") != "" {
		p.CreatedAt, _ = strconv.ParseInt(req.Form.Get("createdAt"), 10, 64)
	}
	if err = SetPusher(p); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	sendJSONResponse(w, http.StatusOK, "result", "OK")
}

func handleGetPushersBySender(w http.ResponseWriter, req *http.Request, sender string) {
	var qs = req.URL.Query()
	var err error
	var from, size int
	var pushers []string
	var realPushers []Pusher
	var total uint64
	if from, err = strconv.Atoi(qs.Get("from")); err != nil {
		from = 0
	}

	if size, err = strconv.Atoi(qs.Get("size")); err != nil {
		size = 10
	}

	if size > 100 {
		size = 100
	}

	total, pushers, err = searchPushers(sender, size, from)
	for _, pusher := range pushers {
		p, _ := GetPusher(pusher)
		if p.ID == pusher {
			realPushers = append(realPushers, p)
		}
	}

	sendJSONResponse(w, http.StatusOK, "", map[string]interface{}{
		"pushers": realPushers,
		"total":   total,
		"from":    from,
		"size":    size,
		"sender":  sender,
	})
}

func wapperPusherHandle(handle func(http.ResponseWriter, *http.Request, string)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		pusher := vars["pusher"]
		handle(w, req, pusher)
	}
}

// NewRouter return new pusher router
func NewRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/pusher/pushers/{pusher}/", wapperPusherHandle(handleGetPusher)).Methods("GET")
	router.HandleFunc("/pusher/pushers/", handleGetAllPusher).Methods("GET")
	router.HandleFunc("/pusher/search/", handleSearchPusher).Methods("GET")

	router.HandleFunc("/pusher/pushers/", handleAddPusher).Methods("POST")
	router.HandleFunc("/pusher/pushers/{pusher}/", wapperPusherHandle(handleRemovePusher)).Methods("DELETE")
	router.HandleFunc("/pusher/pushers/{pusher}/", wapperPusherHandle(handleUpdatePusher)).Methods("POST")

	router.HandleFunc("/pusher/{sender}/add", wapperSenderHandle(handleAddSender)).Methods("POST")
	router.HandleFunc("/pusher/{sender}/pushers/", wapperSenderHandle(handleGetPushersBySender)).Methods("GET")
	router.HandleFunc("/pusher/{sender}/delete", wapperSenderHandle(handleRemoveSender)).Methods("POST")
	router.HandleFunc("/pusher/{sender}/push", wapperSenderHandle(handlePush)).Methods("POST")
	router.HandleFunc("/pusher/{sender}/pushall", wapperSenderHandle(handlePushAll)).Methods("POST")
	return router
}
