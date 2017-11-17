package application

import (
	"net/http"
	"github.com/gorilla/mux"
	"github.com/cabernety/gopkg/httplib"
	"fmt"
	"github.com/BoxLinker/boxlinker-api"
)

func (a *Api) Monitor(w http.ResponseWriter, r *http.Request) {
	user := a.getUserInfo(r)
	serviceName := mux.Vars(r)["serviceName"]
	boxlinker.GetQueryParam(r, "start")
	boxlinker.GetQueryParam(r, "end")
	httplib.Get(fmt.Sprintf("%s/api/v1/query_range", a.config.Monitor.URL)).
		Param("query", "").
		Param("start", "").
		Param("end", "").
		Param("step", "15s")

}
