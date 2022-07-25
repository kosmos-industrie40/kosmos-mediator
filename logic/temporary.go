package logic

import (
	"encoding/json"
	"regexp"
	"strings"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"k8s.io/klog"

	"kosmos-mediator/models"
	mqttClient "kosmos-mediator/mqtt"
)

// representing temporarry data type
type Temporary struct {
	regexUp  *regexp.Regexp
	regexAna *regexp.Regexp
	sendChan chan<- models.MessageBase
}

var topics = [2]string{
	"kosmos/machine-data/+/sensor/+/update/temporary",
	"kosmos/analyses/+/temporary",
}

var regexs = [2]string{
	"kosmos/machine-data/[a-zA-Z0-9]+/sensor/[a-zA-Z0-9]+/update/temporary",
	"kosmos/analyses/[a-zA-Z0-9]+/temporary",
}

// InitTemporary initialise the temporary type and subscribe to the expected */temporary topics
func InitTemporary(mq *mqttClient.MqttWrapper, sendChan chan<- models.MessageBase) error {
	msg := Temporary{
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
func (t Temporary) handler(client MQTT.Client, msg MQTT.Message) {
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
			MessageType:  models.Update,
		}
	// handle analyses message temporary message
	case t.regexAna.MatchString(msg.Topic()):
		klog.V(2).Infof("analyses result message temp")

		topAr := strings.Split(msg.Topic(), "/")
		contract := topAr[2]

		var analyses models.AnalyseResult

		if err := json.Unmarshal(msg.Payload(), &analyses); err != nil {
			klog.Errorf("can not unmarshal received message payload; %s\n", err)
		}

		t.sendChan <- models.MessageBase{
			LastAnalyses: analyses.Body.From,
			Machine:      analyses.Body.Calculated.Message.Machine,
			Sensor:       analyses.Body.Calculated.Message.Sensor,
			Message:      msg.Payload(),
			Contract:     contract,
			MessageType:  models.Analyses,
			Model:        analyses.Body.Model,
		}
	// unexpected topic
	default:
		klog.Errorf("could not verify topic: %s\n", msg.Topic())

	}
}
