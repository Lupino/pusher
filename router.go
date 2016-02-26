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

/**
 * @apiDefine SenderParam
 * @apiParam {String=sendmail, sendsms, customSenderName} sender Sender name.
 * @apiParam {String} pusher Pusher unique ID.
 */

/**
 * @apiDefine DataParam
 * @apiParam {Object} data Sender data.
 * @apiParam {Number} [schedat] when to sched the job.
 * @apiParamExample {json} MailSender data example:
 *     {
 *       "subject": "subject",
 *       "text": "this is the mail text, which use a `text/template` with some keyword {{.NickName}} {{.ID}}",
 *       "createdAt": 1456403493
 *     }
 * @apiParamExample {json} SMSSender data example:
 *     {
 *       "signName": "sms sign name",
 *       "template": "sms template",
 *       "params": "this is the sms template params, which use a `text/template` with some keyword {{.NickName}} {{.ID}}",
 *       "createdAt": 1456403493,
 *     }
 */
/**
 * @apiDefine PusherParam
 * @apiParam {String} pusher Pusher unique ID.
 * @apiParam {String} [email] Pusher email address.
 * @apiParam {String} [phoneNumber] Pusher phone number.
 * @apiParam {String} [nickname] Pusher nickname.
 * @apiParam {Number} [createdAt] Pusher created time.
 */

/**
 * @apiDefine PushResult
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "result": "OK",
 *       "name": "lupino_88bf72bd461965be993c0e6cee9cd061"
 *     }
 */

/**
 * @apiDefine ResultOK
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "result": "OK"
 *     }
 */

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

func push(sender, pusher, data, schedat string, force bool) (string, error) {
	if !hasSender(pusher, sender) {
		log.Printf("pusher[%s] not has sender[%s]", pusher, sender)
		if !force {
			return "", nil
		}
	}
	var opts = map[string]string{
		"args":    data,
		"schedat": schedat,
	}
	var name = generateName(pusher, data)
	if err := periodicClient.SubmitJob(PREFIX+sender, name, opts); err != nil {
		return "", err
	}
	return name, nil
}

func pushAll(sender, data, schedat string) (string, error) {
	var opts = map[string]string{
		"args":    data,
		"schedat": schedat,
	}
	var name = generateName(sender, data)
	if err := periodicClient.SubmitJob(PREFIX+"pushall", name, opts); err != nil {
		return "", err
	}
	return name, nil
}

/**
 * @api {post} /pusher/:sender/add Add a sender to an exists pusher.
 * @apiName addSender
 * @apiGroup Sender
 *
 * @apiUse SenderParam
 * @apiExample Example usage:
 * curl -i http://pusher_host/pusher/sendmail/add -d pusher=lupinno
 *
 *
 * @apiSuccess {String} result OK.
 *
 */
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

/**
 * @api {post} /pusher/:sender/delete delete sender from an exists pusher.
 * @apiName RemoveSender
 * @apiGroup Sender
 *
 * @apiUse SenderParam
 * @apiExample Example usage:
 * curl -i http://pusher_host/pusher/sendmail/delete -d pusher=lupinno
 *
 *
 * @apiSuccess {String} result OK.
 *
 */
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

/**
 * @api {post} /pusher/:sender/push Push message to pusher which has a sender.
 * @apiName push
 * @apiGroup Push
 *
 * @apiUse SenderParam
 * @apiUse DataParam
 * @apiParam {Boolean} [force=false] force push.
 *
 * @apiExample Example usage:
 * curl -i http://pusher_host/pusher/sendmail/push \
 *      -d pusher=lupino \
 *      -d data='{"subject": "subject", "text": "text"}'
 *
 * @apiSuccess {String} result OK.
 * @apiSuccess {String} name The periodic job name.
 * @apiUse PushResult
 *
 */
func handlePush(w http.ResponseWriter, req *http.Request, sender string) {
	f := new(pushForm)
	errs := binding.Bind(req, f)
	if errs.Handle(w) {
		return
	}

	var (
		name string
		err  error
	)

	if name, err = push(sender, f.Pusher, f.Data, f.SchedAt, f.Force); err != nil {
		log.Printf("push() failed (%s)", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	sendJSONResponse(w, http.StatusOK, "", map[string]string{"name": name, "result": "OK"})
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

/**
 * @api {post} /pusher/:sender/pushall Push message to all pusher which has a sender.
 * @apiName pushall
 * @apiGroup Push
 *
 * @apiParam {String=sendmail, sendsms, customSenderName} sender Sender name.
 * @apiUse DataParam
 *
 * @apiExample Example usage:
 * curl -i http://pusher_host/pusher/sendmail/pushall \
 *      -d data='{"subject": "subject", "text": "text"}'
 *
 * @apiSuccess {String} result OK.
 * @apiSuccess {String} name The periodic job name.
 * @apiUse PushResult
 *
 */
func handlePushAll(w http.ResponseWriter, req *http.Request, sender string) {
	f := new(pushAllForm)
	errs := binding.Bind(req, f)
	if errs.Handle(w) {
		return
	}

	var (
		name string
		err  error
	)
	if name, err = pushAll(sender, f.Data, f.SchedAt); err != nil {
		log.Printf("pushAll() failed (%s)", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, http.StatusOK, "", map[string]string{"name": name, "result": "OK"})
}

/**
 * @api {post} /pusher/:sender/cancelpush Cancel message push which not done.
 * @apiName cancelpush
 * @apiGroup Push
 *
 * @apiParam {String=pushall, sendmail, sendsms, customSenderName} sender Sender name.
 * @apiParam {String} name The periodic job name.
 *
 * @apiExample Example usage:
 * curl -i http://pusher_host/pusher/sendmail/cancelpush \
 *      -d name=xxxxxx
 *
 * @apiSuccess {String} result OK.
 * @apiUse ResultOK
 *
 */
func handleCancelPush(w http.ResponseWriter, req *http.Request, sender string) {
	req.ParseForm()
	var (
		name = req.Form.Get("name")
		err  error
	)
	if err = periodicClient.RemoveJob(PREFIX+sender, name); err != nil {
		log.Printf("periodic.Client.RemoveJob() failed (%s)", err)
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

/**
 * @api {get} /pusher/pushers/:pusher/ Get pusher information
 * @apiName GetPusher
 * @apiGroup Pusher
 *
 * @apiParam {String} pusher Pusher unique ID.
 * @apiExample Example usage:
 * curl -i http://pusher_host/pusher/pushers/4711/
 *
 * @apiSuccess {Object} pusher Pusher object.
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "pusher": {
 *         "id": "lupino",
 *         "email": "example@example.com",
 *         "nickname": "Lupino",
 *         "phoneNumber": "12345678901",
 *         "senders": [ "sendmail", "sendsms" ],
 *         "createdAt": 1456403493
 *       }
 *     }
 *
 * @apiError {String} err pusher <code>pusher</code> not exists.
 * @apiErrorExample Response (example):
 *     HTTP/1.1 404 Not Found
 *     {
 *       "err": "pusher 4711 not exists."
 *     }
 *
 */
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

/**
 * @api {get} /pusher/pushers/ Get pusher list
 * @apiName GetPusherList
 * @apiGroup Pusher
 *
 * @apiParam {Number} [from=0] describe how much and which part of the return pusher list
 * @apiParam {Number} [size=10] describe how much and which part of the return pusher list
 * @apiExample Example usage:
 * curl -i http://pusher_host/pusher/pushers/?from=0&size=20
 *
 *
 * @apiSuccess {String} pushers Pusher object list.
 * @apiSuccess {Number} total total pushers.
 * @apiSuccess {Number} from describe how much and which part of the return pusher list
 * @apiSuccess {Number} size describe how much and which part of the return pusher list
 *
 */
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

/**
 * @api {get} /pusher/search/ Search pushers
 * @apiName SearchPusher
 * @apiGroup Pusher
 *
 * @apiParam {String} q search keyword.
 * @apiParam {Number} [from=0] describe how much and which part of the return pusher list
 * @apiParam {Number} [size=10] describe how much and which part of the return pusher list
 * @apiExample Example usage:
 * curl -i http://pusher_host/pusher/search/?q=sendmail&from=0&size=20
 *
 *
 * @apiSuccess {String} pushers Pusher object list.
 * @apiSuccess {Number} total total pushers.
 * @apiSuccess {Number} from describe how much and which part of the return pusher list
 * @apiSuccess {Number} size describe how much and which part of the return pusher list
 * @apiSuccess {String} q search keyword.
 *
 */
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

/**
 * @api {post} /pusher/pushers/ Create a new pusher
 * @apiName AddPusher
 * @apiGroup Pusher
 *
 * @apiUse PusherParam
 * @apiExample Example usage:
 * curl -i http://pusher_host/pusher/pushers/ \
 *      -d pusher=lupino \
 *      -d email=lmjubuntu@gmail.com \
 *      -d phoneNumber=12345678901 \
 *      -d nickname=xxx \
 *      -d createdAt=1456403493
 *
 *
 * @apiSuccess {String} result OK.
 * @apiUse ResultOK
 *
 * @apiError {String} err pusher is required.
 */
func handleAddPusher(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	var p = Pusher{}
	p.ID = req.Form.Get("pusher")
	p.Email = req.Form.Get("email")
	p.NickName = req.Form.Get("nickname")
	p.PhoneNumber = req.Form.Get("phoneNumber")
	p.CreatedAt, _ = strconv.ParseInt(req.Form.Get("createdAt"), 10, 64)
	if p.ID == "" {
		sendJSONResponse(w, http.StatusNotAcceptable, "err", "pusher is required.")
		return
	}

	if err := SetPusher(p); err != nil {
		log.Printf("SetPusher() failed(%s)", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	sendJSONResponse(w, http.StatusOK, "result", "OK")
}

/**
 * @api {DELETE} /pusher/pushers/:pusher/ Remove an exists pusher
 * @apiName ReomvePusher
 * @apiGroup Pusher
 *
 * @apiParam {String} pusher Pusher unique ID.
 * @apiExample Example usage:
 * curl -i -XDELETE http://pusher_host/pusher/pushers/4711/
 *
 * @apiSuccess {String} result OK.
 * @apiUse ResultOK
 *
 */
func handleRemovePusher(w http.ResponseWriter, req *http.Request, pusher string) {
	if err := DelPusher(pusher); err != nil {
		log.Printf("DelPusher() failed(%s)", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	sendJSONResponse(w, http.StatusOK, "result", "OK")
}

/**
 * @api {post} /pusher/pushers/:pusher/ Update an exists pusher
 * @apiName UpdatePusher
 * @apiGroup Pusher
 *
 * @apiUse PusherParam
 * @apiExample Example usage:
 * curl -i http://pusher_host/pusher/pushers/lupino/ \
 *      -d email=lmjubuntu@gmail.com \
 *      -d phoneNumber=12345678901 \
 *      -d nickname=xxx \
 *      -d createdAt=1456403493
 *
 *
 * @apiSuccess {String} result OK.
 * @apiUse ResultOK
 *
 * @apiError {String} err pusher <code> pusher </code> not exists.
 *
 */
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

/**
 * @api {get} /pusher/:sender/pushers/ Get pusher list by sender
 * @apiName GetPusherListBySender
 * @apiGroup Pusher
 *
 * @apiParam {String} sender Sender of pusher.
 * @apiParam {Number} [from=0] describe how much and which part of the return pusher list
 * @apiParam {Number} [size=10] describe how much and which part of the return pusher list
 * @apiExample Example usage:
 * curl -i http://pusher_host/pusher/sendmail/pushers/?from=0&size=20
 *
 * @apiSuccess {String} pushers Pusher object list.
 * @apiSuccess {Number} total total pushers.
 * @apiSuccess {Number} from describe how much and which part of the return pusher list
 * @apiSuccess {Number} size describe how much and which part of the return pusher list
 * @apiSuccess {String} sender Sender of pusher
 *
 */
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
	router.HandleFunc("/pusher/{sender}/cancelpush", wapperSenderHandle(handleCancelPush)).Methods("POST")
	router.HandleFunc("/pusher/{sender}/pushall", wapperSenderHandle(handlePushAll)).Methods("POST")
	return router
}
