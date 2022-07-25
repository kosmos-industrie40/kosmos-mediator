// Package models containing data models, which are used in different parts of the programm
package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"k8s.io/klog"
)

type AnalyseResultBody struct {
	From       string `json:"from"`
	Timestamp  string `json:"timestamp"`
	Type       string `json:"type"`
	Calculated struct {
		Received string `json:"received"`
		Message  struct {
			Machine string `json:"machine"`
			Sensor  string `json:"sensor"`
		}
	} `json:"calculated"`
	Results interface{} `json:"results"`
	Model   Model       `json:"model"`
}

// AnalyseResult representing the AnalyseResult message
type AnalyseResult struct {
	Schema    string            `json:"$schema,omitempty"`
	Body      AnalyseResultBody `json:"body"`
	Signature SignatureObject   `json:"signature,omitempty"`
}

// testExists will test in the database in a defined table if an defined value exists in the defined column
func testExists(db *sql.DB, table, column, value string) (bool, error) {
	dbResult, err := db.Query(fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM %s WHERE %s = $1 LIMIT 1)", table, column), value)
	if err != nil {
		return false, err
	}

	defer func() {
		if err := dbResult.Close(); err != nil {
			klog.Errorf("could not close result from database query (analyse result); %s\n", err)
		}
	}()

	if !dbResult.Next() {
		return false, fmt.Errorf("unexpected db failure; no result found")
	}

	var res bool
	if err := dbResult.Scan(&res); err != nil {
		return false, err
	}

	return res, nil
}

// Insert will insert a new analyse result into a sql database
func (a AnalyseResult) Insert(db *sql.DB, contract string) error {
	var machineExist, sensorExist, contractExist bool
	var err error

	if machineExist, err = testExists(db, "machines", "id", a.Body.Calculated.Message.Machine); err != nil {
		return err
	}

	if sensorExist, err = testExists(db, "sensors", "transmitted_id", a.Body.Calculated.Message.Sensor); err != nil {
		return err
	}

	if contractExist, err = testExists(db, "contracts", "id", contract); err != nil {
		return err
	}

	if !(machineExist && sensorExist && contractExist) {
		return fmt.Errorf("the result is made on a unknown contract, machine or sensor")
	}

	js, err := json.Marshal(a)
	if err != nil {
		return err
	}

	tm, err := time.Parse(time.RFC3339, a.Body.Timestamp)
	if err != nil {
		klog.Errorf("timestamp can not be parsed: %s\n", err)
	}

	machineSensorId, err := GetMachineSensorId(db, a.Body.Calculated.Message.Machine, a.Body.Calculated.Message.Sensor)
	if err != nil {
		klog.Errorf("could not get machine sensor id: %s\n", err)
		return err
	}

	klog.V(2).Infof("Machine sensor id: %d\n", machineSensorId)

	contractMachineSensorId, err := GetContractMachineSensor(db, contract, machineSensorId)
	if err != nil {
		klog.Errorf("Could not get contract machine sensor: %s\n", err)
		return err
	}

	klog.V(2).Infof("Contract machine sensor id: %s\n", contractMachineSensorId)

	_, err = db.Exec(
		"INSERT INTO analysis_result (contract_machine_sensor, time, result) VALUES ($1, $2, $3)",
		contractMachineSensorId,
		tm,
		js)

	return err
}
