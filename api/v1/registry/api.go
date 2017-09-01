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
	"github.com/BoxLinker/boxlinker-api/controller/manager"
	"github.com/BoxLinker/boxlinker-api/pkg/registry/authn"
	"github.com/BoxLinker/boxlinker-api/pkg/registry/tools"
	"strings"
	"sort"
)

type Api struct {
	Listen string
	Manager manager.RegistryManager
	Authenticator authn.Authenticator
	Config *tools.Config
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

type authScope struct {
	Type string
	Name string
	Actions []string
}

type authRequest struct {
	User string
	Password authn.PasswordString
	Account string
	Service string
	Scopes []authScope
}


func prepareRequest(r *http.Request) (*authRequest, error) {
	ar := &authRequest{}
	user, pass, ok := r.BasicAuth()
	if ok {
		ar.User = user
		ar.Password = authn.PasswordString(pass)
	}
	ar.Account = r.FormValue("account")
	if ar.Account == "" {
		ar.Account = ar.User
	} else if ar.Account != "" && ar.Account != ar.User {
		return nil, fmt.Errorf("user and account are not same (%q and %q)", ar.User, ar.Account)
	}

	ar.Service = r.FormValue("service")

	if err := r.ParseForm(); err != nil {
		return nil, fmt.Errorf("invalid form value: %s", err)
	}
	if r.FormValue("scope") != "" {
		for _, scopeStr := range r.Form["scope"] {
			var scope authScope
			parts := strings.Split(scopeStr, ":")
			switch len(parts) {
			case 3:
				scope = authScope{
					Type: parts[0],
					Name: parts[1],
					Actions: strings.Split(parts[2], ","),
				}
			case 4:
				scope = authScope{
					Type: parts[0],
					Name: parts[1] + ":" + parts[2],
					Actions: strings.Split(parts[3], ","),
				}
			default:
				return nil, fmt.Errorf("invalid scope (%q)", scopeStr)
			}
			sort.Strings(scope.Actions)
			ar.Scopes = append(ar.Scopes, scope)
		}
	}

	return ar, nil
}

// POST 	/v1/registry/auth
func (a *Api) DoRegistryAuth(w http.ResponseWriter, r *http.Request){
	ar, err := prepareRequest(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Bad Request: %s", err), http.StatusBadRequest)
		return
	}
	logrus.Debugf("Auth request: %+v", ar)
	authResult, _, err := a.Authenticator.Authenticate(ar.User, ar.Password)
	if err != nil {
		http.Error(w, fmt.Sprintf("Authentication failed (%s)", err), http.StatusUnauthorized)
		return
	}
	if !authResult {
		logrus.Warnf("Auth failed (user:%s)", ar.User)
		w.Header()["WWW-Authenticate"] = []string{fmt.Sprintf(`Basic realm="%s"`, a.Config.Token.Issuer)}
		http.Error(w, "Auth Failed.", http.StatusUnauthorized)
		return
	}

	// authorize based on scopes



	t := &a.Config.Token
	token, err := a.Config.GenerateToken(t.Issuer, ar.Account, ar.Service, t.Expiration)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate token (%s)", err), http.StatusInternalServerError)
		return
	}

	result, _ := json.Marshal(&map[string]string{"token": token})
	logrus.Debugf("generate token: %s", string(result))
	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}
// GET		/v1/registry/images?current_page=1&page_count=10
// GET		/v1/registry/image/:id
// POST		/v1/registry/image
// PUT		/v1/registry/image/:id
// DELETE	/v1/registry/image/:id
// PUT		/v1/registry/image/:id/privilege?private={1|0}

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
	eventRouter.HandleFunc("/v1/registry/auth", a.DoRegistryAuth).Methods("GET")
	eventRouter.HandleFunc("/v1/registry/event", a.RegistryEvent).Methods("POST")
	globalMux.Handle("/v1/registry/", eventRouter)

	s := &http.Server{
		Addr: a.Listen,
		Handler: context.ClearHandler(cs.Handler(globalMux)),
	}

	logrus.Infof("Server run: %s", a.Listen)

	return s.ListenAndServe()
}