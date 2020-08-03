package logic

import (
	"database/sql"
	"encoding/json"
	"regexp"
	"strings"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"k8s.io/klog"

	"gitlab.inovex.de/proj-kosmos/mqtt-database/models"
	"gitlab.inovex.de/proj-kosmos/mqtt-database/mqtt"
)

type SensorUpdate struct {
	db    *sql.DB
	mqtt  *mqtt.MqttWrapper
	regex *regexp.Regexp
}

func Init(db *sql.DB, mqtt *mqtt.MqttWrapper) SensorUpdate {
	regex := regexp.MustCompile("kosmos/machine-data/*/sensor/*/update")
	return SensorUpdate{regex: regex, db: db, mqtt: mqtt}
}

func (su *SensorUpdate) SensorHandler(client MQTT.Client, msg MQTT.Message) {
	//implements what happens with received messages subscribed to before

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
}
