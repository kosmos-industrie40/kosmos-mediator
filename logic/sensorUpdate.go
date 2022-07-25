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

var topic = "kosmos/machine-data/+/sensor/+/update"           // mqtt topic
var regex = "kosmos/machine-data/[^/,]+/sensor/[^/,]+/update" // regex

type SensorUpdate struct {
	db       *sql.DB
	mqtt     *mqttClient.MqttWrapper
	regex    *regexp.Regexp
	sendChan chan<- models.MessageBase
}

// InitSensorUpdate initialises the SensorUpdate logic
// subscribe to a mqtt topic and set the handler of this topic
func InitSensorUpdate(db *sql.DB, mq *mqttClient.MqttWrapper, sendChan chan<- models.MessageBase) error {
	regex := regexp.MustCompile(regex)
	su := SensorUpdate{regex: regex, db: db, mqtt: mq, sendChan: sendChan}
	if err := mq.Subscribe(topic, su.sensorHandler); err != nil {
		return err
	}
	return nil
}

// sensorHandler is a mqtt handler comparing to https://godoc.org/github.com/eclipse/paho.mqtt.golang#MessageHandler
// will create an SensorUpdate model and write this into the database
// in the end the message will be published about the sendChan
func (su SensorUpdate) sensorHandler(client MQTT.Client, msg MQTT.Message) {
	klog.Infof("Rec SensorHandler: TOPIC: %s \n", msg.Topic())

	var sensorBodyData models.SensorBodyUpdate
	var sensorData models.SensorUpdate

	if err := json.Unmarshal(msg.Payload(), &sensorBodyData); err != nil {
		klog.Errorf("Couldn't unmarshal received message payload to SensorBodyUpdate: %s \n", err)
	}

	if sensorBodyData.Body.Timestamp == "" {
		if err := json.Unmarshal(msg.Payload(), &sensorData); err != nil {
			klog.Errorf("Couldn't unmarshal received message payload: %s \n", err)
			return
		}

		sensorBodyData.Body.Columns = sensorData.Columns
		sensorBodyData.Body.Data = sensorData.Data
		sensorBodyData.Body.Meta = sensorData.Meta
		sensorBodyData.Schema = sensorData.Schema
		sensorBodyData.Signature = sensorData.Signature

	}

	if !su.regex.MatchString(msg.Topic()) {
		klog.Errorf("could not verify topic: %s\n", msg.Topic())
		return
	}

	topicSliced := strings.Split(msg.Topic(), "/")
	machineID := topicSliced[2]
	sensorID := topicSliced[4]

	err := sensorBodyData.Insert(su.db, machineID, sensorID)
	if err != nil {
		klog.Errorf("could not insert sensor data into db: %s\n", err)
		return
	}

	data, err := json.Marshal(sensorBodyData)
	if err != nil {
		klog.Errorf("cannot marshal sensorBodyData")
		return
	}

	su.sendChan <- models.MessageBase{
		Machine:      machineID,
		Sensor:       sensorID,
		LastAnalyses: "",
		Contract:     "",
		Message:      data,
		MessageType:  models.Update,
	}

	klog.V(2).Info("sensor update is handled successfully")
}
