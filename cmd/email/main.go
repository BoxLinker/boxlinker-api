package main

import (
	"github.com/urfave/cli"
	log "github.com/Sirupsen/logrus"
	"os"
	api "github.com/cabernety/boxlinker/api/v1/email"
	"github.com/cabernety/boxlinker/controller/amqp"
)

var flags = []cli.Flag{
	cli.StringFlag{
		Name: "listen, l",
		Value: ":8080",
		EnvVar: "LISTEN",
	},
	cli.BoolFlag{
		Name: "test",
		EnvVar: "TEST",
	},
	cli.BoolFlag{
		Name: "debug, D",
		EnvVar: "DEBUG",
	},

	cli.StringFlag{
		Name: "mail-host",
		Value: "smtp.exmail.qq.com:25",
		EnvVar: "MAIL_HOST",
	},
	cli.StringFlag{
		Name: "mail-user",
		Value: "service@boxlinker.com",
		EnvVar: "MAIL_USER",
	},
	cli.StringFlag{
		Name: "mail-user-title",
		Value: "Boxlinker",
		EnvVar: "MAIL_USER_TITLE",
	},
	cli.StringFlag{
		Name: "mail-password",
		Value: "Just4fun",
		EnvVar: "MAIL_PASSWORD",
	},
	cli.StringFlag{
		Name: "mail-type",
		Value: "html",
		EnvVar: "MAIL_TYPE",
	},

	cli.StringFlag{
		Name: "rabbitmq-uri",
		Value: "amqp://guest:guest@localhost:5672/",
		EnvVar: "RABBITMQ_URI",
	},
	cli.StringFlag{
		Name: "rabbitmq-exchange",
		Value: "test-exchange",
		EnvVar: "RABBITMQ_EXCHANGE",
	},
	cli.StringFlag{
		Name: "rabbitmq-exchange-type",
		Usage: "Exchange type - direct|fanout|topic|x-custom",
		Value: "fanout",
		EnvVar: "RABBITMQ_EXCHANGE_TYPE",
	},
	cli.StringFlag{
		Name: "rabbitmq-queue-name",
		Value: "test-queue-name",
		EnvVar: "RABBITMQ_QUEUE_NAME",
	},
	cli.StringFlag{
		Name: "rabbitmq-consumer-tag",
		Usage: "AMQP consumer tag (should not be blank)",
		Value: "boxlinker-email",
		EnvVar: "RABBITMQ_CONSUMER_TAG",
	},
	cli.StringFlag{
		Name: "rabbitmq-binding-key",
		Value: "boxlinker-email-amqp-binding-key",
		EnvVar: "RABBITMQ_BINDING_KEY",
	},


}

func main(){
	app := cli.NewApp()
	app.Name = "Boxlinker Email server"
	app.Before = func(c *cli.Context) error {
		if c.Bool("debug") {
			log.SetLevel(log.DebugLevel)
		}
		return nil
	}
	app.Action = action
	app.Flags = flags

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func action(c *cli.Context) error {

	notifyMsg := make(chan []byte)

	amqpConsumer := &amqp.Consumer{
		URI: c.String("rabbitmq-uri"),
		Exchange: c.String("rabbitmq-exchange"),
		ExchangeType: c.String("rabbitmq-exchange-type"),
		QueueName: c.String("rabbitmq-queue-name"),
		Tag: c.String("rabbitmq-consumer-tag"),
		BindingKey: c.String("rabbitmq-binding-key"),
		NotifyMsg: notifyMsg,
	}

	if err := amqpConsumer.Run(); err != nil {
		return err
	}


	return api.NewApi(api.ApiOptions{
		Listen: c.String("listen"),
		AMQPConsumer: amqpConsumer,
		EmailOption: api.EmailOption{
			User: c.String("mail-user"),
			UserTitle: c.String("mail-user-title"),
			Host: c.String("mail-host"),
			Password: c.String("mail-password"),
		},
		TestMode: c.Bool("test"),
	}).Run()


}


