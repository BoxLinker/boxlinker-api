package main

import (
	"github.com/urfave/cli"
	"github.com/Sirupsen/logrus"
	"os"
	api "github.com/cabernety/boxlinker/api/v1/registry"
	"fmt"
)

var flags = []cli.Flag{
	cli.StringFlag{
		Name: "listen, l",
		Value: ":8080",
		EnvVar: "LISTEN",
	},
	cli.BoolFlag{
		Name: "debug, D",
		EnvVar: "DEBUG",
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
	app.Flags = flags

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func action(c *cli.Context) error {

	a := &api.Api{
		Listen: c.String("listen"),
	}

	return fmt.Errorf("Run Api err: %v", a.Run())
}