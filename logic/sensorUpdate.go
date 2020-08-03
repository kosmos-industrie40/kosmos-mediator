package logic

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"k8s.io/klog"

	"gitlab.inovex.de/proj-kosmos/mqtt-database/database"
	"gitlab.inovex.de/proj-kosmos/mqtt-database/models"
	mqttClient "gitlab.inovex.de/proj-kosmos/mqtt-database/mqtt"
)

type SensorUpdate struct {
	Db   *database.DbWrapper
	Mqtt *mqttClient.MqttWrapper
}

func (su *SensorUpdate) SensorHandler(client MQTT.Client, msg MQTT.Message) {
	//implements what happens with received messages subscribed to before

	klog.Infof("Rec SensorHandler: TOPIC: %s \n", msg.Topic())

	var sensorData models.SensorUpdate

	if err := json.Unmarshal(msg.Payload(), &sensorData); err != nil {
		klog.Errorf("Couldn't unmarshal received message payload: %s \n", err)
	}

	topicSliced := strings.Split(msg.Topic(), "/") // kosmos/machine-data/<machineID>/sensor/<sensorID>/update
	var machineID = topicSliced[2]
	var sensorID, _ = strconv.Atoi(topicSliced[4])

	if err := su.AddSensorUpdate(machineID, sensorID, sensorData); err != nil {
		klog.Errorf("could not insert message into db: %s\n", err)
	}

	klog.Infof("Inserted received Sensor Update: %s\n", msg.Topic())

}

func (su *SensorUpdate) AddSensorUpdate(machineID string, sensorID int, msg models.SensorUpdate) error {

	if machineID == "" || sensorID < 0 {
		return fmt.Errorf("machine id or sensor id not specified")
	}
	if len(msg.Columns) != len(msg.Data) {
		return fmt.Errorf("Number of columns ist not equal to number of data_sets")
	}

	//Check if contract for machine exists

	var read_contract string
	var read_contract_intefrace interface{} = read_contract
	read_contract_queryResult := []*interface{}{&read_contract_intefrace}

	if err := su.Db.Query("machine_contract", []string{"contract"}, []string{"machine"}, []interface{}{machineID}, read_contract_queryResult); err != nil {
		return err
	}

	if read_contract_intefrace.(string) == "" {
		return fmt.Errorf("There is no contract for machine_id: %s \n", machineID)
	}

	// Check if sensor exists for specified machine
	var read_sensor int64
	var read_sensor_interface interface{} = read_sensor
	read_sensor_queryResult := []*interface{}{&read_sensor_interface}

	if err := su.Db.Query("machine_sensor", []string{"sensor"}, []string{"machine", "sensor"}, []interface{}{machineID, sensorID}, read_sensor_queryResult); err != nil {
		return err
	}
	fmt.Printf("content %d type %T \n", read_sensor_interface, read_sensor_interface)

	if read_sensor_interface.(int64) == 0 {
		return fmt.Errorf("There is no sensor_id: %d for machine_id: %s \n", sensorID, machineID)
	}

	//add message and get it's id
	sensor_message_id, err := su.Db.InsertReturn("sensor_message", []string{"sensor", "timestamp"}, []interface{}{sensorID, time.Unix(msg.Timestamp, 0)}, "id")

	if err != nil {
		return err
	}

	for _, meta := range msg.Metadata {
		if err := su.Db.Insert("sensor_meta", []string{"sensor_message", "name", "description", "type", "value"}, []interface{}{sensor_message_id, meta.Name, meta.Description, meta.Type, meta.Value}); err != nil {
			return err
		}
	}

	//if columns do not exist create them and save their id to sensor_columns_ids for later referencing
	var sensor_columns_ids []int64
	for _, column := range msg.Columns {

		var sensor_column_id int64
		var columnIdInterface interface{} = sensor_column_id
		queryResult := []*interface{}{&columnIdInterface}

		if err := su.Db.Query("sensor_column", []string{"id"}, []string{"name", "description", "type"}, []interface{}{column.Name, column.Description, column.Type}, queryResult); err != nil {
			return err
		}

		if columnIdInterface.(int64) == 0 { //if colum. does exist get it's current id
			newColumnID, err := su.Db.InsertReturn("sensor_column", []string{"name", "description", "type"}, []interface{}{column.Name, column.Description, column.Type}, "id")
			if err != nil {
				return err
			}

			sensor_columns_ids = append(sensor_columns_ids, int64(newColumnID.(int)))

		} else {
			sensor_columns_ids = append(sensor_columns_ids, columnIdInterface.(int64))
		}
	}

	//write all data with their corresponding column_id to the database
	for dataset_idx, dataset := range msg.Data {
		for _, data := range dataset {
			if err := su.Db.Insert("sensor_data", []string{"sensor_message", "sensor_column", "value"}, []interface{}{sensor_message_id, sensor_columns_ids[dataset_idx], data}); err != nil {
				return err
			}
		}
	}

	return nil
}
