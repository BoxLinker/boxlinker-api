package registry

import (
	"net/http"
	"github.com/rs/cors"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"fmt"
	"io/ioutil"
	"github.com/Sirupsen/logrus"
	"encoding/json"
)

type Api struct {
	Listen string
}

type RegistryCallback struct {
	Events [] struct{
		Id 			string 	`json:"id"`
		Timestamp 	string 	`json:"timestamp"`
		Action 		string 	`json:"action"`
		Target 		struct{
			MediaType 		string 		`json:"mediaType"`
			Size 			int64 		`json:"size"`
			Digest 			string 		`json:"digest"`
			Length 			int64 		`json:"length"`
			Repository 		string 		`json:"repository"`
			Url 			string 		`json:"url"`
			Tag 			string 		`json:"tag"`
		} `json:"target"`
		Request 	struct{
			Id 		string 		`json:"id"`
			Addr 	string 		`json:"addr"`
			Host	string 		`json:"host"`
			Method 	string 		`json:"method"`
			UserAgent string 	`json:"useragent"`
		} 	`json:"request"`
		Source 		struct{
			Addr 	string 	`json:"addr"`
			InstanceID string `json:"instanceID"`
		} 	`json:"source"`
	}	`json:"events"`
}
// POST 	/v1/registry/auth
// GET		/v1/registry/images?current_page=1&page_count=10
// GET		/v1/registry/image/:id
// POST		/v1/registry/image
// PUT		/v1/registry/image/:id
// DELETE	/v1/registry/image/:id
// PUT		/v1/registry/image/:id/privilege?private={true|false}

// POST		/v1/registry/event
func (a *Api) RegistryEvent(w http.ResponseWriter, r *http.Request){
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("read body: %v", err.Error()), http.StatusInternalServerError)
		return
	}
	events := &RegistryCallback{}
	if err := json.Unmarshal(b, events); err != nil {
		http.Error(w, fmt.Sprintf("Unmarshal body: %v", err.Error()), http.StatusInternalServerError)
		return
	}
	// 确认镜像以及 tag 是否存在，如果不存在创建镜像记录
	// 创建 image:tag action 记录

	fmt.Printf("r.Body:>\n %+v", events)
}

func (a * Api) Run() error {
	cs := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "DELETE", "PUT", "OPTIONS"},
		AllowedHeaders: []string{"Origin", "Content-Type", "Accept", "token", "X-Requested-With", "X-Access-Token"},
	})

	globalMux := http.NewServeMux()

	eventRouter := mux.NewRouter()
	eventRouter.HandleFunc("/v1/registry/event", a.RegistryEvent).Methods("POST")
	globalMux.Handle("/v1/registry/event", eventRouter)

	s := &http.Server{
		Addr: a.Listen,
		Handler: context.ClearHandler(cs.Handler(globalMux)),
	}

	logrus.Debugf("Server run: %s", a.Listen)

	return s.ListenAndServe()
}