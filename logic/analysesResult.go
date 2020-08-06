package logic

import (
	"database/sql"
	"encoding/json"
	"regexp"
	"strings"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"k8s.io/klog"

	"gitlab.inovex.de/proj-kosmos/intern-mqtt-db/models"
	mqttClient "gitlab.inovex.de/proj-kosmos/intern-mqtt-db/mqtt"
)

var topi string = "kosmos/analyses/+"            // mqtt topic
var rege string = "kosmos/analyses/[0-9a-zA-Z]+" // regex

type AnalysesResult struct {
	db       *sql.DB
	mqtt     *mqttClient.MqttWrapper
	regex    *regexp.Regexp
	sendChan chan<- MessageBase
}

// InitAnalyseResult initialise the analyse result
func InitAnalyseResult(db *sql.DB, mq *mqttClient.MqttWrapper, sendChan chan<- MessageBase) error {
	reg := regexp.MustCompile(rege)
	ar := AnalysesResult{regex: reg, db: db, mqtt: mq, sendChan: sendChan}
	if err := mq.Subscribe(topi, ar.handler); err != nil {
		return err
	}
	return nil
}

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

	ar.sendChan <- MessageBase{
		Machine:      analyse.Calculated.Message.Machine,
		Sensor:       analyse.Calculated.Message.Sensor,
		LastAnalyses: analyse.From,
	}

	klog.V(2).Info("analyse result is handeld successfully")
}
