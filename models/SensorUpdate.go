package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"k8s.io/klog"
)

// SensorBodyUpdate representing the mqtt sensor update message
type SensorBodyUpdate struct {
	Schema string `json:"$schema,omitempty"`
	Body   struct {
		Timestamp string      `json:"timestamp"`
		Columns   interface{} `json:"columns"`
		Data      interface{} `json:"data"`
		Meta      interface{} `json:"meta,omitempty"`
	} `json:"body"`
	Signature SignatureObject `json:"signature,omitempty"`
}

// Insert will insert a new sensor update message into a sql database
func (s SensorBodyUpdate) Insert(db *sql.DB, machine string, sensor string) error {
	result, err := db.Query("SELECT machine_sensors.id FROM machine_sensors JOIN sensors ON sensors.id = machine_sensors.sensor WHERE sensors.transmitted_id = $1 AND machine_sensors.machine = $2", sensor, machine)
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
		return fmt.Errorf("could not find machine-sensor combination; machine id is %s and transmitted sensor id is: %s", machine, sensor)
	}

	if err := result.Scan(&id); err != nil {
		return fmt.Errorf("could not parse machine-sensor.id to int64: %s", err)
	}

	meta, err := json.Marshal(s.Body.Meta)
	if err != nil {
		return fmt.Errorf("could not marshal meta: %s", err)
	}

	columns, err := json.Marshal(s.Body.Columns)
	if err != nil {
		return fmt.Errorf("could not marshal columns: %s", err)
	}

	data, err := json.Marshal(s.Body.Data)
	if err != nil {
		return fmt.Errorf("could not marshal data: %s", err)
	}

	signatureJson, err := json.Marshal(s.Signature)
	if err != nil {
		return fmt.Errorf("could not marshal the signature object: %s", err)
	}

	tm, err := time.Parse(time.RFC3339, s.Body.Timestamp)
	if err != nil {
		klog.Errorf("timestamp can not be parsed: %s\n", err)
	}

	if _, err := db.Exec("INSERT INTO update_message (sensor_machine, timestamp, meta, column_definition, data, signature) VALUES ($1, $2, $3, $4, $5, $6)", id, tm, meta, columns, data, signatureJson); err != nil {
		return fmt.Errorf("could not insert update_message data: %s", err)
	}

	return nil
}

// SensorUpdate representing the mqtt sensor update message
type SensorUpdate struct {
	Schema    string          `json:"$schema,omitempty"`
	Timestamp string          `json:"timestamp"`
	Columns   interface{}     `json:"columns"`
	Data      interface{}     `json:"data"`
	Signature SignatureObject `json:"signature"`
	Meta      interface{}     `json:"meta"`
}

// Insert will insert a new sensor update message into a sql database
func (s SensorUpdate) Insert(db *sql.DB, machine string, sensor string) error {
	result, err := db.Query("SELECT machine_sensors.id FROM machine_sensors JOIN sensors ON sensors.id = machine_sensors.sensor WHERE sensors.transmitted_id = $1 AND machine_sensors.machine = $2", sensor, machine)
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
		return fmt.Errorf("could not find machine-sensor combination; machine id is %s and transmitted sensor id is: %s", machine, sensor)
	}

	if err := result.Scan(&id); err != nil {
		return fmt.Errorf("could not parse machine-sensor.id to int64: %s", err)
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

	signatureJson, err := json.Marshal(s.Signature)
	if err != nil {
		return fmt.Errorf("could not marshal the signature object: %s", err)
	}

	tm, err := time.Parse(time.RFC3339, s.Timestamp)
	if err != nil {
		klog.Errorf("timestamp can not be parsed: %s\n", err)
	}

	if _, err := db.Exec("INSERT INTO update_message (sensor_machine, timestamp, meta, column_definition, data, signature) VALUES ($1, $2, $3, $4, $5, $6)", id, tm, meta, columns, data, signatureJson); err != nil {
		return fmt.Errorf("could not insert update_message data: %s", err)
	}

	return nil
}
