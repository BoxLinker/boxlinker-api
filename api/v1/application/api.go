package application

import (
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"fmt"
	"github.com/BoxLinker/boxlinker-api/controller/manager"
	tAuth "github.com/BoxLinker/boxlinker-api/controller/middleware/auth_token"
	userModels "github.com/BoxLinker/boxlinker-api/controller/models/user"
	"net/http"
	"github.com/Sirupsen/logrus"
	"github.com/BoxLinker/boxlinker-api"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/codegangsta/negroni"
	"k8s.io/client-go/kubernetes"
)

type Api struct {
	config *Config
	manager manager.ApplicationManager
	clientSet *kubernetes.Clientset
}

type ApiConfig struct {
	Config *Config
	ControllerManager manager.ApplicationManager
	ClientSet *kubernetes.Clientset
}

func NewApi(config ApiConfig) (*Api, error) {
	return &Api{
		config: config.Config,
		manager: config.ControllerManager,
		clientSet: config.ClientSet,
	}, nil
}

type Config struct {
	Server struct {
		Addr string `yaml:"addr,omitempty"`
		Debug bool `yaml:"debug"`
	}    `yaml:"server,omitempty"`
	DB struct{
		Host string `yaml:"host,omitempty"`
		Port int `yaml:"port,omitempty"`
		User string `yaml:"user,omitempty"`
		Password string `yaml:"password,omitempty"`
		Name string `yaml:"name,omitempty"`
	} `yaml:"db,omitempty"`
	Auth struct{
		TokenAuthUrl string `yaml:"tokenAuthUrl,omitempty"`
		BasicAuthUrl string `yaml:"basicAuthUrl,omitempty"`
	} `yaml:"auth,omitempty"`
	K8S struct{
		KubeConfig string `yaml:"kubeconfig"`
	} `yaml:"k8s"`
}

func LoadConfig(cPath string) (*Config, error) {
	contents, err := ioutil.ReadFile(cPath)
	if err != nil {
		return nil, err
	}
	c := &Config{}

	if err := yaml.Unmarshal(contents, c); err != nil {
		return nil, fmt.Errorf("load config file err: %s", err)
	}

	return c, nil
}

func (a *Api) Run() error {
	cs := boxlinker.Cors
	// middleware
	apiAuthRequired := tAuth.NewAuthTokenRequired(a.config.Auth.TokenAuthUrl)

	globalMux := http.NewServeMux()

	serviceRouter := mux.NewRouter()
	serviceRouter.HandleFunc("/v1/application/auth/service", a.CreateService).Methods("POST")

	authRouter := negroni.New()
	authRouter.Use(negroni.HandlerFunc(apiAuthRequired.HandlerFuncWithNext))
	authRouter.UseHandler(serviceRouter)
	globalMux.Handle("/v1/application/auth/", authRouter)

	s := &http.Server{
		Addr: a.config.Server.Addr,
		Handler: context.ClearHandler(cs.Handler(globalMux)),
	}

	logrus.Infof("Server run: %s", a.config.Server.Addr)

	return s.ListenAndServe()
}

func (a *Api) getUserInfo(r *http.Request) *userModels.User {
	us := r.Context().Value("user")
	if us == nil {
		return nil
	}
	ctx := us.(map[string]interface{})
	if ctx == nil || ctx["uid"] == nil {
		return nil
	}
	return &userModels.User{
		Id: ctx["uid"].(string),
		Name: ctx["username"].(string),
	}
}

func int32Ptr(i int32) *int32 { return &i }