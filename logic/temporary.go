package logic

import (
	"encoding/json"
	"regexp"
	"strings"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"k8s.io/klog"

	"gitlab.inovex.de/proj-kosmos/intern-mqtt-db/models"
	mqttClient "gitlab.inovex.de/proj-kosmos/intern-mqtt-db/mqtt"
)

// representing temporarry data type
type Temprorary struct {
	regexUp  *regexp.Regexp
	regexAna *regexp.Regexp
	sendChan chan<- models.MessageBase
}

var topics = [2]string{
	"kosmos/machine-data/+/sensor/+/update/temprary",
	"kosmos/analyses/+/temporary",
}

var regexs = [2]string{
	"kosmos/machine-data/[a-zA-Z0-9]+/sensor/[a-zA-Z0-9]+/update/temprary",
	"kosmos/analyses/[a-zA-Z0-9]+/temporary",
}

// IInitTemprorary initialise the temporary type and subscribe to the expected */temporrary topics
func InitTemprorary(mq *mqttClient.MqttWrapper, sendChan chan<- models.MessageBase) error {
	msg := Temprorary{
		regexUp:  regexp.MustCompile(regexs[0]),
		regexAna: regexp.MustCompile(regexs[1]),
		sendChan: sendChan,
	}

	if err := mq.Subscribe(topics[0], msg.handler); err != nil {
		return err
	}
	if err := mq.Subscribe(topics[1], msg.handler); err != nil {
		return err
	}

	return nil
}

// handler callback functions which handle incomming messages
func (t Temprorary) handler(client MQTT.Client, msg MQTT.Message) {
	klog.Infof("receive temp on topic %s\n", msg.Topic())

	switch {

	// handle sensor message temporary message
	case t.regexUp.MatchString(msg.Topic()):
		klog.V(2).Infof("update sensor message temp")
		topicSliced := strings.Split(msg.Topic(), "/")
		machineID := topicSliced[2]
		sensorID := topicSliced[4]
		t.sendChan <- models.MessageBase{
			Machine:      machineID,
			Sensor:       sensorID,
			LastAnalyses: "",
			Message:      msg.Payload(),
			Contract:     "",
		}

		topAr := strings.Split(msg.Topic(), "/")
		contract := topAr[2]

		var analyses models.AnalyseResult

		if err := json.Unmarshal(msg.Payload(), &analyses); err != nil {
			klog.Errorf("can not unmarshal received message payload; %s\n", err)
		}

		t.sendChan <- models.MessageBase{
			LastAnalyses: analyses.From,
			Machine:      analyses.Calculated.Message.Machine,
			Sensor:       analyses.Calculated.Message.Sensor,
			Message:      msg.Payload(),
			Contract:     contract,
		}

	// handle analyses message temporary message
	case t.regexAna.MatchString(msg.Topic()):
		klog.V(2).Infof("analyses result message temp")

	// unexpected topic
	default:
		klog.Errorf("could not verify topic: %s\n", msg.Topic())

	}
}
