package user

import (
	"net/http"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/cabernety/boxlinker/controller/manager"
	mAuth "github.com/cabernety/boxlinker/controller/middleware/auth"
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
)

type ApiOptions struct {
	Listen string
	Manager manager.Manager
}

type Api struct {
	listen string
	manager manager.Manager
}

func NewApi(config ApiOptions) *Api {
	return &Api{
		listen: config.Listen,
		manager: config.Manager,
	}
}
// get 	/v1/user/auth/token
// post /v1/user/auth/login
// post	/v1/user/auth/reg
// put	/v1/user/account/list
// put	/v1/user/account/:id/changepassword
// get	/v1/user/account/:id
func (a *Api) Run() error {
	cs := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "DELETE", "PUT", "OPTIONS"},
		AllowedHeaders: []string{"Origin", "Content-Type", "Accept", "token", "X-Requested-With", "X-Access-Token"},
	})
	// middleware
	apiAuthRequired := mAuth.NewAuthRequired(a.manager)

	globalMux := http.NewServeMux()

	loginRegRouter := mux.NewRouter()
	loginRegRouter.HandleFunc("/v1/user/auth/login", a.Login).Methods("POST")
	loginRegRouter.HandleFunc("/v1/user/auth/reg", a.Reg).Methods("POST")
	globalMux.Handle("/v1/user/auth/", loginRegRouter)

	accountRouter := mux.NewRouter()
	accountRouter.HandleFunc("/v1/user/account/list", a.GetUsers).Methods("GET")
	accountRouter.HandleFunc("/v1/user/account/changepassword", a.ChangePassword).Methods("PUT")
	accountRouter.HandleFunc("/v1/user/account/userinfo", a.GetUser).Methods("GET")
	accountAuthRouter := negroni.New()
	accountAuthRouter.Use(negroni.HandlerFunc(apiAuthRequired.HandlerFuncWithNext))
	accountAuthRouter.UseHandler(accountRouter)
	globalMux.Handle("/v1/user/account/", accountAuthRouter)

	s := &http.Server{
		Addr:    a.listen,
		Handler: context.ClearHandler(cs.Handler(globalMux)),
	}

	log.Infof("Server listen on: %s", a.listen)

	return s.ListenAndServe()
}
