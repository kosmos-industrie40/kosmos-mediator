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

var topic string = "kosmos/machine-data/+/sensor/+/update" // mqtt topic
var regex string = "kosmos/machine-data/*/sensor/*/update" // regex (has to be updated TODO)

type SensorUpdate struct {
	db       *sql.DB
	mqtt     *mqttClient.MqttWrapper
	regex    *regexp.Regexp
	sendChan chan<- MessageBase
}

func InitSensorUpdate(db *sql.DB, mq *mqttClient.MqttWrapper, sendChan chan<- MessageBase) error {
	regex := regexp.MustCompile(regex)
	su := SensorUpdate{regex: regex, db: db, mqtt: mq, sendChan: sendChan}
	if err := mq.Subscribe(topic, su.sensorHandler); err != nil {
		return err
	}
	return nil
}

func (su SensorUpdate) sensorHandler(client MQTT.Client, msg MQTT.Message) {
	klog.Infof("Rec SensorHandler: TOPIC: %s \n", msg.Topic())

	var sensorData models.SensorUpdate

	if err := json.Unmarshal(msg.Payload(), &sensorData); err != nil {
		klog.Errorf("Couldn't unmarshal received message payload: %s \n", err)
		return
	}

	if !su.regex.MatchString(msg.Topic()) {
		klog.Errorf("could not verify topic: %s\n", msg.Topic())
		return
	}

	topicSliced := strings.Split(msg.Topic(), "/")
	machineID := topicSliced[2]
	sensorID := topicSliced[4]

	err := sensorData.Insert(su.db, machineID, sensorID)
	if err != nil {
		klog.Errorf("could not insert data into db: %s\n", err)
		return
	}

	su.sendChan <- MessageBase{
		Machine:      machineID,
		Sensor:       sensorID,
		LastAnalyses: "",
	}
}
