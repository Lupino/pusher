package pusher

import (
	"github.com/Lupino/pusher/utils"
	"net/http"
	"strconv"
	"time"
)

// Auth pusher api request
func (s SPusher) Auth(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	var appKey = req.Header.Get("X-App-Key")
	if s.key != appKey {
		sendJSONResponse(rw, http.StatusBadRequest, "err", "Invalid X-App-Key")
		return
	}
	var timestamp = req.Header.Get("X-Request-Time")
	var unixTimestamp, _ = strconv.ParseInt(timestamp, 10, 0)
	var reqTime = time.Unix(unixTimestamp, 0)
	var took = time.Since(reqTime)
	if took < 0 {
		took = -took
	}
	if took > 10*time.Minute {
		sendJSONResponse(rw, http.StatusBadRequest, "err", "Invalid X-Request-Time")
		return
	}

	var signParams = make(map[string]string)
	signParams["app_key"] = appKey
	signParams["timestamp"] = timestamp
	signParams["path"] = req.URL.Path
	var query = req.URL.Query()
	if query != nil {
		for key := range query {
			signParams[key] = query.Get(key)
		}
	}
	if req.Method == "POST" {
		req.ParseForm()
		if req.Form != nil {
			for key := range req.Form {
				signParams[key] = req.Form.Get(key)
			}

		}
	}
	var exceptSign = utils.HmacMD5(s.secret, signParams)
	var sign = req.Header.Get("X-Request-Signature")
	if sign != exceptSign {
		sendJSONResponse(rw, http.StatusBadRequest, "err", "Invalid X-Request-Signature")
		return
	}
	next(rw, req)
}
