// Package models containing data models, which are used in different parts of the programm
package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"

	"k8s.io/klog"
)

// AnalyseResult representing the AnalyseResult message
type AnalyseResult struct {
	Schema     string      `json:"$schema,omitempty"`
	From       string      `json:"from"`
	Timestamp  string      `json:"timestamp"`
	Signature  string      `json:"signature,omitempty"`
	Results    interface{} `json:"results"`
	Calculated struct {
		Received int64 `json:"received"`
		Message  struct {
			Machine string `json:"machine"`
			Sensor  string `json:"sensor"`
		}
	} `json:"calculated"`
	Model Model `json:"model"`
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

	if machineExist, err = testExists(db, "machine", "id", a.Calculated.Message.Machine); err != nil {
		return err
	}

	if sensorExist, err = testExists(db, "sensor", "transmitted_id", a.Calculated.Message.Sensor); err != nil {
		return err
	}

	if contractExist, err = testExists(db, "contract", "id", contract); err != nil {
		return err
	}

	if !(machineExist && sensorExist && contractExist) {
		return fmt.Errorf("the result is made on a unknown contract, machine or sensor")
	}

	js, err := json.Marshal(a)
	if err != nil {
		return err
	}
	match, err := regexp.MatchString("^[0-9]{4}-[]0-9]{2}-[0-9]{2}T[012][0-9]:[0-5][0-9]:[0-5][0-9].[0-9]*$", a.Timestamp)
	if err != nil {
		klog.Errorf("can not use regexp: %s\n", err)
		return err
	}

	if !match  {
		klog.V(2).Infof("timestamp does not match")
		return nil
	}
	_, err = db.Exec("INSERT INTO analyse_result (contract, machine, sensor, time, result) VALUES ($1, $2, $3, $4, $5)", contract, a.Calculated.Message.Machine, a.Calculated.Message.Sensor, a.Timestamp, js)

	return err
}
