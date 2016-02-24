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
	var info Info
	if info, err = GetInfo(pusher); err != nil {
		return
	}
	changed := false
	for _, sender := range senders {
		if info.AddSender(sender) {
			changed = true
		}
	}

	if changed {
		if err = SetInfo(info); err != nil {
			return
		}
	}
	return
}

func removeSender(pusher string, senders ...string) (err error) {
	var info Info
	if info, err = GetInfo(pusher); err != nil {
		return
	}
	changed := false
	for _, sender := range senders {
		if info.DelSender(sender) {
			changed = true
		}
	}

	if changed {
		if err = SetInfo(info); err != nil {
			return
		}
	}
	return
}

func hasSender(pusher, sender string) bool {
	var info Info
	var err error
	if info, err = GetInfo(pusher); err != nil {
		return false
	}
	return info.HasSender(sender)
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
		for _, s := range hit.Fields["senders"].([]interface{}) {
			if ss, ok := s.(string); ok {
				if ss == sender {
					pushers = append(pushers, hit.ID)
				}
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

func handleGetInfo(w http.ResponseWriter, req *http.Request, pusher string) {
	info, err := GetInfo(pusher)
	if err != nil {
		log.Printf("GetInfo() failed (%s)", err)
	}
	if info.ID != pusher {
		sendJSONResponse(w, http.StatusNotFound, "err", "pusher "+pusher+" not exists.")
		return
	}
	sendJSONResponse(w, http.StatusOK, "pusher", info)
}

func handleGetAllInfo(w http.ResponseWriter, req *http.Request) {
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
	var pushers []Info
	for _, pusher := range ret.Val() {
		info, _ := GetInfo(pusher)
		if info.ID == pusher {
			pushers = append(pushers, info)
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

func handleSearchInfo(w http.ResponseWriter, req *http.Request) {
	var qs = req.URL.Query()
	var err error
	var from, size int
	var total uint64
	var pushers []Info
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
		info, _ := GetInfo(hit.ID)
		if info.ID == hit.ID {
			pushers = append(pushers, info)
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

func handleAddInfo(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	var info = Info{}
	info.ID = req.Form.Get("pusher")
	info.Email = req.Form.Get("email")
	info.RealName = req.Form.Get("realname")
	info.NickName = req.Form.Get("nickname")
	info.PhoneNumber = req.Form.Get("phoneNumber")
	info.CreatedAt, _ = strconv.ParseInt(req.Form.Get("createdAt"), 10, 64)
	if info.ID == "" {
		sendJSONResponse(w, http.StatusOK, "err", "pusher is required.")
		return
	}

	if err := SetInfo(info); err != nil {
		log.Printf("SetInfo() failed(%s)", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	sendJSONResponse(w, http.StatusOK, "result", "OK")
}

func handleRemoveInfo(w http.ResponseWriter, req *http.Request, pusher string) {
	if err := DelInfo(pusher); err != nil {
		log.Printf("DelInfo() failed(%s)", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	sendJSONResponse(w, http.StatusOK, "result", "OK")
}

func handleUpdateInfo(w http.ResponseWriter, req *http.Request, pusher string) {
	var info Info
	var err error
	if info, err = GetInfo(pusher); err != nil {
		log.Printf("DelInfo() failed(%s)", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if info.ID != pusher {
		sendJSONResponse(w, http.StatusNotFound, "err", "pusher "+pusher+" not exists.")
		return
	}

	if req.Form.Get("email") != "" {
		info.Email = req.Form.Get("email")
	}
	if req.Form.Get("realname") != "" {
		info.RealName = req.Form.Get("realname")
	}
	if req.Form.Get("nickname") != "" {
		info.NickName = req.Form.Get("nickname")
	}
	if req.Form.Get("phoneNumber") != "" {
		info.PhoneNumber = req.Form.Get("phoneNumber")
	}
	if req.Form.Get("createdAt") != "" {
		info.CreatedAt, _ = strconv.ParseInt(req.Form.Get("createdAt"), 10, 64)
	}
	if err = SetInfo(info); err != nil {
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
	var realPushers []Info
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
		info, _ := GetInfo(pusher)
		if info.ID == pusher {
			realPushers = append(realPushers, info)
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
	router.HandleFunc("/pusher/pushers/{pusher}/", wapperPusherHandle(handleGetInfo)).Methods("GET")
	router.HandleFunc("/pusher/pushers/", handleGetAllInfo).Methods("GET")
	router.HandleFunc("/pusher/search/", handleSearchInfo).Methods("GET")

	router.HandleFunc("/pusher/pushers/", handleAddInfo).Methods("POST")
	router.HandleFunc("/pusher/pushers/{pusher}/", wapperPusherHandle(handleRemoveInfo)).Methods("DELETE")
	router.HandleFunc("/pusher/pushers/{pusher}/", wapperPusherHandle(handleUpdateInfo)).Methods("POST")

	router.HandleFunc("/pusher/{sender}/add", wapperSenderHandle(handleAddSender)).Methods("POST")
	router.HandleFunc("/pusher/{sender}/pushers/", wapperSenderHandle(handleGetPushersBySender)).Methods("GET")
	router.HandleFunc("/pusher/{sender}/delete", wapperSenderHandle(handleRemoveSender)).Methods("POST")
	router.HandleFunc("/pusher/{sender}/push", wapperSenderHandle(handlePush)).Methods("POST")
	router.HandleFunc("/pusher/{sender}/pushall", wapperSenderHandle(handlePushAll)).Methods("POST")
	return router
}
