// Package logic handles the logic of the programm.
// It will register handler on mqtt topics and contains the logic of the
// mediator
package logic

import (
	"database/sql"
	"encoding/json"
	"regexp"
	"strings"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"k8s.io/klog"

	"kosmos-mediator/models"
	mqttClient "kosmos-mediator/mqtt"
)

var topi string = "kosmos/analyses/+"            // mqtt topic
var rege string = "kosmos/analyses/[0-9a-zA-Z]+" // regex

// internal representation of the logic of the analytic result
type AnalysesResult struct {
	db       *sql.DB
	mqtt     *mqttClient.MqttWrapper
	regex    *regexp.Regexp
	sendChan chan<- models.MessageBase
}

// InitAnalyseResult initialise the analytic result logic
// and subscribe to the specific topic and set the handler
func InitAnalyseResult(db *sql.DB, mq *mqttClient.MqttWrapper, sendChan chan<- models.MessageBase) error {
	reg := regexp.MustCompile(rege)
	ar := AnalysesResult{regex: reg, db: db, mqtt: mq, sendChan: sendChan}
	if err := mq.Subscribe(topi, ar.handler); err != nil {
		return err
	}
	return nil
}

// handler is a mqtt handler comparing to https://godoc.org/github.com/eclipse/paho.mqtt.golang#MessageHandler
// this function will create an AnalyseResult model and store it into the database
// in the end the message will be published about the sendChan
func (ar AnalysesResult) handler(client MQTT.Client, msg MQTT.Message) {
	klog.Infof("Rec AnalyseResult handler: TOPIC: %s \n", msg.Topic())

	var analyse models.AnalyseResult

	if err := json.Unmarshal(msg.Payload(), &analyse); err != nil {
		klog.Errorf("could not unmarshal received message payload; %s\n", err)
		return
	}

	if !ar.regex.MatchString(msg.Topic()) {
		klog.Errorf("could not verify topic: %s\n", msg.Topic())
		return
	}

	topAr := strings.Split(msg.Topic(), "/")
	contract := topAr[2]

	if err := analyse.Insert(ar.db, contract); err != nil {
		klog.Errorf("could not insert analyse result into database: %s\n", err)
		return
	}

	ar.sendChan <- models.MessageBase{
		Machine:      analyse.Body.Calculated.Message.Machine,
		Sensor:       analyse.Body.Calculated.Message.Sensor,
		LastAnalyses: analyse.Body.From,
		Contract:     contract,
		Message:      msg.Payload(),
		MessageType:  models.Analyses,
		Model:        analyse.Body.Model,
	}

	klog.V(2).Info("analyse result is handled successfully")
}
