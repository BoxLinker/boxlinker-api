package application

import (
	"net/http"
	"github.com/gorilla/mux"
	"github.com/cabernety/gopkg/httplib"
	"fmt"
	"github.com/BoxLinker/boxlinker-api"
	"time"
	"io/ioutil"
	"encoding/json"
	"github.com/Sirupsen/logrus"
)

type prometheusResult struct {
	Status string `json:"status"`
	Data struct{
		ResultType string `json:"resultType"`
		Result []struct{
			Metric map[string]interface{} `json:"metric"`
			Values [][]interface{} `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

func (a *Api) Monitor(w http.ResponseWriter, r *http.Request) {
	user := a.getUserInfo(r)
	serviceName := mux.Vars(r)["serviceName"]
	start := boxlinker.GetQueryParam(r, "start")
	end := boxlinker.GetQueryParam(r, "end")

	if _, err := time.Parse("2006-01-02T15:04:05.000Z", start); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_PARAM_ERR, nil, "start param err")
		return
	}
	if _, err := time.Parse("2006-01-02T15:04:05.000Z", end); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_PARAM_ERR, nil, "end param err")
		return
	}

	res, err := httplib.Get(fmt.Sprintf("%s/api/v1/query_range", a.config.Monitor.URL)).
		Param("query", fmt.Sprintf(
			"container_memory_usage_bytes{container_name=\"%s\",namespace=\"%s\"}",
			serviceName, user.Name)).
		Param("start", start).
		Param("end", end).
		Param("step", "15s").Response()
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, err.Error())
		return
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, err.Error())
		return
	}
	var result prometheusResult
	if err := json.Unmarshal(b, &result); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, err.Error())
		return
	}

	if result.Status == "success" && result.Data.ResultType == "matrix" {
		if len(result.Data.Result) == 1 {
			boxlinker.Resp(w, boxlinker.STATUS_OK, result.Data.Result[0].Values)
			return
		}
	}
	logrus.Warnf("prometheus search err, result -> \n %+v", result)

	boxlinker.Resp(w, boxlinker.STATUS_FAILED, result)

}
