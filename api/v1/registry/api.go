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
	//registryModels "github.com/BoxLinker/boxlinker-api/controller/models/registry"
	"github.com/BoxLinker/boxlinker-api/pkg/registry/authn"
	"strings"
	"sort"
	"github.com/BoxLinker/boxlinker-api/pkg/registry/authz"
	"time"
	"net"
)

type Api struct {
	Listen string
	Manager manager.RegistryManager
	Authenticator authn.Authenticator
	Authorizers []authz.Authorizer
	Config *Config
}
type ApiConfig struct {
	Listen string
	Manager manager.RegistryManager
	BasicAuthURL string
	ConfigFilePath string
}
func NewApi(ac *ApiConfig) (*Api, error) {

	config, err := LoadConfig(ac.ConfigFilePath)
	if err != nil {
		return nil, err
	}

	a := &Api{
		Listen: ac.Listen,
		Manager: ac.Manager,
		Config: config,
		Authorizers: []authz.Authorizer{},
	}
	// authenticator
	a.Authenticator = &authn.DefaultAuthenticator{
		BasicAuthURL: ac.BasicAuthURL,
	}

	//if err := ac.Manager.SaveACL(&registryModels.ACL{
	//	Account: "*",
	//	Name: "library/*",
	//	Actions: "*",
	//}); err != nil {
	//	return nil, err
	//}

	// authorizes
	if config.ACL != nil {
		staticAuthorizer, err := authz.NewACLAuthorizer(config.ACL)
		if err != nil {
			return nil, err
		}
		a.Authorizers = append(a.Authorizers, staticAuthorizer)
	}

	mysqlAuthorizer, err := authz.NewACLMysqlAuthorizer(authz.ACLMysqlConfig{
		Manager: ac.Manager,
		CacheTTL: time.Second * 60,
	})
	if err != nil {
		return nil, err
	}
	a.Authorizers = append(a.Authorizers, mysqlAuthorizer)

	return a, nil
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

type authzResult struct {
	scope authScope
	authorizedActions []string
}

type authRequest struct {
	RemoteConnAddr string
	RemoteAddr     string
	RemoteIP       net.IP
	User           string
	Password       authn.PasswordString
	Account        string
	Service        string
	Scopes         []authScope
	Labels         authn.Labels
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

func (a *Api) authorizeScope(ai *authz.AuthRequestInfo) ([]string, error) {

	for i, authorizer := range a.Authorizers {
		result, err := authorizer.Authorize(ai)
		logrus.Infof("Authz %s %s -> %s, %v", authorizer.Name(), *ai, result, err)
		if err != nil {
			if err == authz.NoMatch {
				continue
			}
			err = fmt.Errorf("authz #%d returned error: %s", i+1, err)
			logrus.Errorf("%s: %s", *ai, err)
			return nil, err
		}
		return result, nil
	}
	logrus.Warningf("%s did not match any authz rule", *ai)
	return nil, nil
}

func (a *Api) authorize(ar *authRequest) ([]authzResult, error) {
	ares := []authzResult{}
	for _, scope := range ar.Scopes {
		ai := &authz.AuthRequestInfo{
			Account: ar.Account,
			Type: scope.Type,
			Name: scope.Name,
			Service: ar.Service,
			Actions: scope.Actions,
		}
		actions, err := a.authorizeScope(ai)
		if err != nil {
			return nil, err
		}
		ares = append(ares, authzResult{scope: scope, authorizedActions: actions})
	}
	return ares, nil
}

// POST 	/v1/registry/auth
func (a *Api) DoRegistryAuth(w http.ResponseWriter, r *http.Request){
	ar, err := prepareRequest(r)
	ares := []authzResult{}
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
	if len(ar.Scopes) > 0 {
		ares, err = a.authorize(ar)
		if err != nil {
			http.Error(w, fmt.Sprintf("Authorization failed (%s)", err), http.StatusInternalServerError)
			return
		}
	} else {
		// Authentication-only request ("docker login"), pass through.
	}

	//t := &a.Config.Token
	token, err := a.Config.GenerateToken(ar, ares)
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