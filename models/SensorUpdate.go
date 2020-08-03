package models

import (
	"database/sql"
	"fmt"

	"k8s.io/klog"
)

type SensorUpdate struct {
	Schema    string      `json:"$schema,omitempty"`
	Timestamp int64       `json:"timestamp"`
	Columns   interface{} `json:"columns"`
	Data      interface{} `json:"data"`
	Signature string      `json:"signature"`
	Meta      interface{} `json:"meta:omitempty"`
}

func (s SensorUpdate) Insert(db *sql.DB, machine string, sensor string) error {
	klog.Infof("identify sensor id")
	result, err := db.Query("SELECT id FROM machine_sensor JOIN sensor ON sensor.id = machine_sensor.sensor WHERE sensor.transmitted_id = $1 AND machine_sensor.machine = $2", sensor, machine)
	if err != nil {
		klog.Errorf("could not query id from machine_sensor; %s\n", err)
	}

	defer func() {
		if err := result.Close(); err != nil {
			klog.Errorf("could not close result from datbase query (sensor update); %s\n", err)
		}
	}()

	var id int64

	// no result will be found
	if !result.Next() {
		//TODO error handling; we want to store all sensor data?
		return fmt.Errorf("could not find machine-sensor combination")
	}

	if err := result.Scan(&id); err != nil {
		return fmt.Errorf("could not parse machine-sensor.id to int64: %s\n", err)
	}

	klog.Infof("insert sensor update data into database")
	if _, err := db.Exec(fmt.Sprintf("INSERT INTO update_message (sensor_machine, timestamp, meta, column, data, signature) VALUES ($1, $2, $3, $4, $5, $6)"), id, s.Timestamp, s.Meta, s.Columns, s.Data, s.Signature); err != nil {
		return fmt.Errorf("could not insert update_message data: %s\n", err)
	}

	return nil
}
