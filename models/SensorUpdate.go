package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"

	"k8s.io/klog"
)

// SensorUpdate representing the mqtt sensor update message
type SensorUpdate struct {
	Schema    string      `json:"$schema,omitempty"`
	Timestamp string      `json:"timestamp"`
	Columns   interface{} `json:"columns"`
	Data      interface{} `json:"data"`
	Signature string      `json:"signature"`
	Meta      interface{} `json:"meta"`
}

// Insert will insert a new sensor update message into a sql database
func (s SensorUpdate) Insert(db *sql.DB, machine string, sensor string) error {
	result, err := db.Query("SELECT machine_sensor.id FROM machine_sensor JOIN sensor ON sensor.id = machine_sensor.sensor WHERE sensor.transmitted_id = $1 AND machine_sensor.machine = $2", sensor, machine)
	if err != nil {
		return err
	}

	defer func() {
		if err := result.Close(); err != nil {
			klog.Errorf("could not close result from datbase query (sensor update); %s\n", err)
		}
	}()

	var id int64

	// no result will be found
	if !result.Next() {
		return fmt.Errorf("could not find machine-sensor combination; machine id is %s and transmitted sensor id is: %s\n", machine, sensor)
	}

	if err := result.Scan(&id); err != nil {
		return fmt.Errorf("could not parse machine-sensor.id to int64: %s\n", err)
	}

	meta, err := json.Marshal(s.Meta)
	if err != nil {
		return fmt.Errorf("could not marshal meta: %s", err)
	}

	columns, err := json.Marshal(s.Columns)
	if err != nil {
		return fmt.Errorf("could not marshal columns: %s", err)
	}

	data, err := json.Marshal(s.Data)
	if err != nil {
		return fmt.Errorf("could not marshal data: %s", err)
	}

	match, err := regexp.MatchString("^[0-9]{4}-[]0-9]{2}-[0-9]{2}T[012][0-9]:[0-5][0-9]:[0-5][0-9].[0-9]*$", s.Timestamp)
	if err != nil {
		klog.Errorf("can not use regexp: %s\n", err)
	}

	if !match {
		klog.V(2).Infof("timestamp does not match")
		return nil
	}

	if _, err := db.Exec("INSERT INTO update_message (sensor_machine, timestamp, meta, column_definition, data, signature) VALUES ($1, $2, $3, $4, $5, $6)", id, s.Timestamp, meta, columns, data, s.Signature); err != nil {
		return fmt.Errorf("could not insert update_message data: %s\n", err)
	}

	return nil
}
