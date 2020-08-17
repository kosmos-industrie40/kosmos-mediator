package main

import (
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/klog"

	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"os"

	"gitlab.inovex.de/proj-kosmos/intern-mqtt-db/config"
	"gitlab.inovex.de/proj-kosmos/intern-mqtt-db/logic"
	"gitlab.inovex.de/proj-kosmos/intern-mqtt-db/models"
	mq "gitlab.inovex.de/proj-kosmos/intern-mqtt-db/mqtt"
)

var cli struct {
	password      string
	configuration string
}

func init() {
	klog.InitFlags(nil)
	flag.StringVar(&cli.password, "pass", "examplePassword.yaml", "is the path to the password configuration")
	flag.StringVar(&cli.configuration, "config", "exampleConfiguration.yaml", "is the path to the configuration file")
}

func main() {
	flag.Parse()

	var pas models.Password
	var conf models.Configuration

	// parse configuration and password
	if err := config.ParseConfigurations(cli.configuration, &conf); err != nil {
		klog.Errorf("could not parse configuration: %s\n", err)
		os.Exit(1)
	}

	if err := config.ParsePassword(cli.password, &pas); err != nil {
		klog.Errorf("could not parse password: %s\n", err)
		os.Exit(1)
	}

	conStr := fmt.Sprintf("host=%s user=%s password=%s port=%d sslmode=disable dbname=%s", conf.Database.Address, pas.Database.User, pas.Database.Password, conf.Database.Port, conf.Database.Database)
	db, err := sql.Open("postgres", conStr)
	if err != nil {
		klog.Errorf("could not connect to database: %s\n", err)
	}

	sendChan := make(chan models.MessageBase, 100)

	mqtt := mq.MqttWrapper{}
	if err := mqtt.Init(pas.Mqtt.User, pas.Mqtt.Password, conf.Mqtt.Address, conf.Mqtt.Port, conf.Mqtt.Tls); err != nil {
		klog.Errorf("cannot connect with mqtt broker: %s\n", err)
		os.Exit(1)
	}

	go logic.Mediator(db, mqtt, sendChan)

	if err := logic.InitSensorUpdate(db, &mqtt, sendChan); err != nil {
		klog.Errorf("can not subscribe sensor update: %s\n", err)
		os.Exit(1)
	}

	if err := logic.InitAnalyseResult(db, &mqtt, sendChan); err != nil {
		klog.Errorf("can not subscribe sensor update: %s\n", err)
		os.Exit(1)
	}

	if err := logic.InitTemporary(&mqtt, sendChan); err != nil {
		klog.Errorf("can not subscribe to temporary topics: %s\n", err)
		os.Exit(1)
	}

	// enable monitoring
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(fmt.Sprintf("%s:%d", conf.Webserver.Address, conf.Webserver.Port), nil); err != nil {
		klog.Errorf("can not start webserver; %s\n", err)
		os.Exit(1)
	}
}
