package main

import (
	"github.com/urfave/cli"
	"github.com/Sirupsen/logrus"
	"os"
	cmd "github.com/BoxLinker/boxlinker-api/cmd"
	registryModels "github.com/BoxLinker/boxlinker-api/controller/models/registry"
	api "github.com/BoxLinker/boxlinker-api/api/v1/registry"
	"fmt"
	"github.com/BoxLinker/boxlinker-api/controller/models"
	"github.com/BoxLinker/boxlinker-api/controller/manager"
	"github.com/BoxLinker/boxlinker-api/pkg/registry/authn"
	"errors"
	"github.com/BoxLinker/boxlinker-api/pkg/registry/tools"
)

var flags = []cli.Flag{
	cli.StringFlag{
		Name: "basic-auth-url",
		Value: "http://localhost:8080/v1/user/auth/basicAuth",
		EnvVar: "BASIC_AUTH_URL",
	},

	cli.StringFlag{
		Name: "config-file",
		Value: "./auth_config.yml",
		EnvVar: "CONFIG_FILE",
	},
}

func main(){
	app := cli.NewApp()
	app.Name = "Boxlinker Email server"
	app.Before = func(c *cli.Context) error {
		if c.Bool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
		}
		return nil
	}
	app.Action = action
	app.Flags = append(flags, append(cmd.DBFlags, cmd.SharedFlags...)...)

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}


func action(c *cli.Context) error {

	configFilePath := c.String("config-file")
	if len(configFilePath) == 0 {
		return errors.New("no config file provided")
	}

	config, err := tools.LoadConfig(configFilePath)
	if err != nil {
		return err
	}

	basicAuthURL := c.String("basic-auth-url")
	if len(basicAuthURL) == 0 {
		return errors.New("basic-auth-url is required")
	}

	engine, err := models.NewEngine(models.GetDBOptions(c), registryModels.Tables())
	if err != nil {
		return fmt.Errorf("new db engine err: %v", err)
	}

	controllerManager, err := manager.NewRegistryManager(engine)
	if err != nil {
		return fmt.Errorf("new controller manager err: %v", err)
	}

	// authenticator
	authenticator := &authn.DefaultAuthenticator{
		BasicAuthURL: basicAuthURL,
	}

	a := &api.Api{
		Listen: c.String("listen"),
		Manager: controllerManager,
		Authenticator: authenticator,
		Config: config,
	}

	return fmt.Errorf("Run Api err: %v", a.Run())
}