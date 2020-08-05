package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"k8s.io/klog"
)

type AnalyseResult struct {
	Schema     string      `json:"$schema,omitempty"`
	From       string      `json:"from"`
	Date       int64       `json:"date"`
	Signature  string      `json:"signature,omitempty"`
	Results    interface{} `json:"results"`
	Calculated struct {
		Received int64 `json:"received"`
		Message  struct {
			Machine string `json:"machine"`
			Sensor  string `json:"sensor"`
		}
	} `json:"calculated"`
}

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

// Insert will insert a new analyse result into the database
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

	tm := time.Unix(a.Date, 0)
	json, err := json.Marshal(a)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO analyse_result (contract, machine, sensor, time, result) VALUES ($1, $2, $3, $4, $5)", contract, a.Calculated.Message.Machine, a.Calculated.Message.Sensor, tm, json)

	return err
}
