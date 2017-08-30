package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/cabernety/boxlinker/controller/manager"
	"github.com/cabernety/boxlinker/auth/builtin"
	api "github.com/cabernety/boxlinker/api/v1/user"

	"os"
	settings "github.com/cabernety/boxlinker/settings/user"
	"fmt"
)

var (
	flags = []cli.Flag{
		cli.BoolFlag{
			Name:   "debug, D",
			Usage:  "enable debug",
			EnvVar: "DEBUG",
		},
		cli.StringFlag{
			Name:   "listen, l",
			Value:  ":8080",
			Usage:  "server listen address",
			EnvVar: "LISTEN",
		},
		cli.StringFlag{
			Name:   "db-user",
			Value:  "root",
			EnvVar: "DB_USER",
		},
		cli.StringFlag{
			Name:   "db-password",
			Value:  "123456",
			EnvVar: "DB_PASSWORD",
		},
		cli.StringFlag{
			Name:   "db-host",
			Value:  "127.0.0.1",
			EnvVar: "DB_HOST",
		},
		cli.StringFlag{
			Name:   "db-port",
			Value:  "3306",
			EnvVar: "DB_PORT",
		},
		cli.StringFlag{
			Name:   "db-name",
			Value:  "boxlinker",
			EnvVar: "DB_NAME",
		},
		cli.StringFlag{
			Name:   "admin-name",
			Value:  "admin",
			EnvVar: "ADMIN_NAME",
		},
		cli.StringFlag{
			Name:   "admin-password",
			Value:  "Admin123456",
			EnvVar: "ADMIN_PASSWORD",
		},
		cli.StringFlag{
			Name:   "admin-email",
			Value:  "service@boxlinker.com",
			EnvVar: "ADMIN_EMAIL",
		},
		cli.StringFlag{
			Name: 	"user-password-salt",
			Value:	"arandomuserpasswordsalt",
			EnvVar: "USER_PASSWORD_SALT",
		},
		cli.StringFlag{
			Name:  "cookie-domain",
			Value: "localhost",
			EnvVar: "COOKIE_DOMAIN",
		},


	}
)

func main() {
	app := cli.NewApp()
	app.Name = "Boxlinker 用户服务"
	app.Usage = "Boxlinker 用户服务"
	app.Action = action
	app.Before = func(c *cli.Context) error {
		log.SetLevel(log.DebugLevel)
		if c.Bool("debug") {
			log.SetLevel(log.DebugLevel)
		}
		return nil
	}
	app.Flags = flags

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func action(c *cli.Context) error {

	settings.InitSettings(c)


	authenticator := builtin.NewAuthenticator()

	controllerManager, err := manager.NewManager(manager.ManagerOptions{
		Authenticator:	authenticator,
		DBUser: 		c.String("db-user"),
		DBPassword: 	c.String("db-password"),
		DBHost: 		c.String("db-host"),
		DBPort: 		c.Int("db-port"),
		DBName: 		c.String("db-name"),
	})

	if err != nil {
		return fmt.Errorf("New Manager: %s", err.Error())
	}

	if err := controllerManager.CheckAdminUser(); err != nil {
		return fmt.Errorf("CheckAdminUser: %v", err)
	}

	return api.NewApi(api.ApiOptions{
		Listen: c.String("listen"),
		Manager: controllerManager,
	}).Run()

}
