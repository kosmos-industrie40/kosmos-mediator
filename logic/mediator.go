package logic

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"k8s.io/klog"

	"gitlab.inovex.de/proj-kosmos/intern-mqtt-db/models"
	mqttClient "gitlab.inovex.de/proj-kosmos/intern-mqtt-db/mqtt"
)

// Mediator find the next model based on the MessageBase and sends an SendMsg message
func Mediator(db *sql.DB, mq mqttClient.MqttWrapper, sendChan <-chan models.MessageBase) {
	for {
		base := <-sendChan
		var (
			model models.Model
			err   error
			typ   string
		)

		switch base.MessageTyp {
		// handle messages based on analytics
		case models.Analyses:
			mo := base.Model
			exists, err := mo.TestEnd(db, base.Machine, base.Sensor, base.Contract)
			if err != nil {
				klog.Errorf("could not test if the current model is the last one: %s\n", err)
				continue
			}

			if !exists {
				continue
			}

			model, err = mo.Next(db, base.Machine, base.Sensor, base.Contract)
			if err != nil {
				klog.Errorf("cannot find the next analytic model message will not be further processed: %s\n", err)
				continue
			}
			typ = "analyse_result"
		// handle messages based on update
		case models.Update:
			model, err = model.InitialPipeline(db, base.Machine, base.Sensor, base.Contract)
			if err != nil {
				klog.Errorf("cannot find the next analytic model message will not be further processed: %s\n", err)
				continue
			}
			typ = "sensor_update"
		default:
			klog.Errorf("Unexpected MessageType, message will not be futher processed")
			continue
		}

		// url escaping
		encode := url.QueryEscape(model.Url)
		// repalce / with - to precent several subtopics
		modelUrl := strings.ReplaceAll(encode, "/", "-")

		msg := models.SendMsg{
			Contract: base.Contract,
			Type:     typ,
			Payload:  base.Message,
		}

		bytes, err := json.Marshal(msg)
		if err != nil {
			klog.Errorf("cannot create json of the message: %s\n", err)
			continue
		}

		mqMsg := mqttClient.Msg{
			Topic: fmt.Sprintf("kosmos/analytics/%s/%s", modelUrl, model.Tag),
			Msg:   bytes,
		}

		if err := mq.Publish(mqMsg); err != nil {
			klog.Errorf("error in publishing the mqtt message: %s\n", err)
			continue
		}
	}
}
