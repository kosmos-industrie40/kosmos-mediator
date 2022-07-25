package logic

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"k8s.io/klog"

	"kosmos-mediator/models"
	mqttClient "kosmos-mediator/mqtt"
)

// Mediator find the next model based on the MessageBase and sends an SendMsg message
func Mediator(db *sql.DB, mq mqttClient.MqttWrapper, sendChan <-chan models.MessageBase) {
	for {
		base := <-sendChan
		var (
			model models.Model
			//err   error
			typ string
		)

		switch base.MessageType {
		// handle messages based on analytics
		case models.Analyses:
			klog.V(2).Infof("handle analyse case")
			previousModel := base.Model
			isEnd, err := previousModel.TestEnd(db, base.Machine, base.Sensor, base.Contract)
			if err != nil {
				klog.Errorf("could not test if the current model is the last one: %s\n", err)
				continue
			}

			// The previous model was the last model for the pipeline therefore the pipeline is done.
			// Todo this does not allow an pipeline to use a model as an intermediate step which is also used as an end step
			if isEnd {
				klog.V(2).Infof("Pipeline end found! (Last Model: %s:%s)", previousModel.Url, previousModel.Tag)
				continue
			}

			model, err = previousModel.Next(db, base.Machine, base.Sensor, base.Contract)
			if err != nil {
				klog.Errorf("cannot find the next analytic model message will not be further processed: %s\n", err)
				continue
			}
			typ = "analyse_result"
			if err = forwardMqttMessage(mq, base, model, typ); err != nil {
				var baseJson []byte
				baseJson, marshalErr := json.Marshal(base)
				if err != nil {
					klog.Errorf("Error getting baseJson while handeling error %s.\n New error is \n %s", err, marshalErr)
				}
				klog.Errorf("Could not forward Message: %s due to %s", baseJson, err)
				continue
			}

		// handle messages based on update
		case models.Update:
			// Get all models which are at the start of piepelines for that contract
			contractId, modelResults, err := model.InitialPipelines(db, base.Machine, base.Sensor)
			if err != nil {
				klog.Errorf("cannot find the next analytic model message will not be further processed: %s\n", err)
				continue
			}
			// Specify the message specifics and forward message to respective pipeline startpoints
			base.Contract = contractId
			typ := "sensor_update"
			for _, modelResult := range modelResults {
				if err = forwardMqttMessage(mq, base, modelResult, typ); err != nil {
					var baseJson []byte
					baseJson, marshalErr := json.Marshal(base)
					if err != nil {
						klog.Errorf("Error getting baseJson while handeling error %s.\n New error is \n %s", err, marshalErr)
					}
					klog.Errorf("Could not forward Message: %s due to %s", baseJson, err)
					continue
				}
				klog.V(2).Infof("Started pipeline with URL: %s, Tag: %s", modelResult.Url, modelResult.Tag)
			}

		default:
			klog.Errorf("Unexpected MessageType, message will not be further processed")
			continue
		}
	}
}

// forwarMqttMessage forwards the message which was received last to the next analyses model
func forwardMqttMessage(mq mqttClient.MqttWrapper, base models.MessageBase, model models.Model, typ string) (err error) {
	// Get old message in order to forward it to the correct recipient
	var oldMsg interface{}
	if err := json.Unmarshal(base.Message, &oldMsg); err != nil {
		klog.Errorf("cannot unmarshal the previous received message")
		return err
	}
	// Create mesage from models.MessageBase
	mqttMsgBody := models.SendMsg{
		Body: models.SendBody{
			Contract: base.Contract,
			Type:     typ,
			Payload:  oldMsg,
			Machine:  base.Machine,
			Sensor:   base.Sensor,
		},
	}
	bytes, err := json.Marshal(mqttMsgBody)
	if err != nil {
		klog.Errorf("cannot create json of the message: %s\n", err)
		return err
	}

	// url escaping
	// encode := url.QueryEscape(model.Url)
	// replace / with - to prevent several subtopics
	// modelUrl := strings.ReplaceAll(model.Url, "/", "-")

	// Build final MQTT message
	mqMsg := mqttClient.Msg{
		Topic: fmt.Sprintf("kosmos/analytics/%s/%s", model.Url, model.Tag),
		Msg:   bytes,
	}
	// Publish MQTT message
	if err := mq.Publish(mqMsg); err != nil {
		klog.Errorf("error in publishing the mqtt message: %s\n", err)
		return err
	}
	return nil
}
