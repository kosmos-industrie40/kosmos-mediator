package models

import (
	"database/sql"
	"fmt"

	"k8s.io/klog"
)

// Model representing the database table model
type Model struct {
	Url string `json:"url"`
	Tag string `json:"tag"`
}

// InitialPipeline find the initial model in the pipeline
func (m Model) InitialPipelines(db *sql.DB, machine, sensor string) (string, []Model, error) {
	// Create array to store all Models in which can be found for the contract specified by machine and sensor ID
	var result []Model = []Model{}

	// Get Machine-Sensor ID
	machineSensorId, err := GetMachineSensorId(db, machine, sensor)
	if err != nil {
		return "", result, err
	}
	klog.V(2).Infof("id of machine-sensor %d", machineSensorId)

	// Get respective Contract
	contractId, err := GetContract(db, machine, machineSensorId)
	if err != nil {
		return "", result, err
	}

	klog.V(2).Infof("id of the contract: %s\n", contractId)

	// Get contract machine sensor to get pipelines
	contractMachineSensorID, err := GetContractMachineSensor(db, contractId, machineSensorId)
	if err != nil {
		klog.Errorf("Couldnt get contract machine sensor for contract %s, and machine sensor %d", contractId, machineSensorId)
		return contractId, []Model{}, err
	}
	// Get Pipeline to get analysis
	pipelineId, err := GetPipelineId(db, contractMachineSensorID)
	if err != nil {
		klog.Errorf("Couldnt get pipeline from contract machine sensor %s", contractMachineSensorID)
		return contractId, []Model{}, err
	}

	// Choose Pipeline startpoint Look at WHERE Prev, Next, execute
	//res, err := db.Query("SELECT url, tag FROM models JOIN next_analyses on models.id = next_analyses.next_model JOIN pipeline on next_analyses.id = pipeline.analyses WHERE contract = $1 AND machine_sensor = $2 AND previous_model is NULL", contractId, machineSensorId)
	klog.V(2).Infof("parameters of the next query: pipelineId = %s\n", pipelineId)
	rows, err := db.Query("SELECT container FROM models JOIN analysis on models.id = analysis.execute JOIN pipelines on analysis.pipeline = pipelines.id WHERE pipelines.id=$1 AND prev_model is NULL", pipelineId)
	if err != nil {
		return contractId, []Model{}, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			klog.Errorf("could not close result query: %s\n", err)
		}
	}()

	// Go through all analysis object that full fill the specified criteria
	for rows.Next() {
		var singleContainerId string
		// Get next result from rows
		err = rows.Scan(&singleContainerId)
		if err != nil {
			klog.Errorf("Could not find initial model in database for pipeline: %s", pipelineId)
			return contractId, []Model{}, err
		}
		url, tag, err := GetUrlAndTagFromContainer(db, singleContainerId)
		if err != nil {
			klog.Errorf("Could not get url and tag for container with ID: %s", singleContainerId)
			return contractId, []Model{}, err
		}
		result = append(result, Model{Url: url, Tag: tag})
	}
	klog.V(2).Infof("Found the following entrypoints: %s", result)
	return contractId, result, err
}

// TestEnd test if the last analyses in the pipeline has be made
func (m Model) TestEnd(db *sql.DB, machine, sensor, contract string) (bool, error) {
	// Get PipelineId Through machine and sensor combination
	machSens, err := GetMachineSensorId(db, machine, sensor)
	if err != nil {
		klog.V(2).Infof("error in getMachineSensorId")
		return false, err
	}
	klog.V(2).Infof("machine-sensor id is %d\n", machSens)
	conMachSens, err := GetContractMachineSensor(db, contract, machSens)
	if err != nil {
		klog.V(2).Infof("error in GetContractMachineSensor")
		return false, err
	}
	pipelineId, err := GetPipelineId(db, conMachSens)
	if err != nil {
		klog.V(2).Infof("error in GetPipelineId")
		return false, err
	}
	klog.V(2).Infof("pipeline id is %s\n", pipelineId)
	// Get current Model to check if it was the last one
	currentModel, err := m.getIdModel(db)
	if err != nil {
		return false, err
	}

	klog.V(2).Infof("parameter to use in query are: pipelineId: %s; currModel: %d", pipelineId, currentModel)
	res, err := db.Query("SELECT EXISTS (SELECT analysis.pipeline FROM analysis JOIN pipelines ON analysis.pipeline=pipelines.id WHERE pipelines.id=$1 AND analysis.execute=$2 AND analysis.next_model is NULL);", pipelineId, currentModel)
	if err != nil {
		return false, err
	}

	defer func() {
		if err := res.Close(); err != nil {
			klog.Errorf("can not close rows: %s\n", err)
		}
	}()

	var isEnd bool

	for res.Next() {
		if err := res.Scan(&isEnd); err != nil {
			klog.Errorf("cannot scan err is: %s\n", err)
		}

	}
	return isEnd, nil

}

// Next query the next model based on the current model
func (m Model) Next(db *sql.DB, machine, sensor, contract string) (Model, error) {

	machSens, err := GetMachineSensorId(db, machine, sensor)
	if err != nil {
		return Model{}, err
	}
	conMachSens, err := GetContractMachineSensor(db, contract, machSens)
	if err != nil {
		klog.V(2).Infof("error in GetContractMachineSensor")
		return Model{}, err
	}
	pipelineId, err := GetPipelineId(db, conMachSens)
	if err != nil {
		klog.V(2).Infof("error in GetPipelineId")
		return Model{}, err
	}
	klog.V(2).Infof("pipeline id is %s\n", pipelineId)
	// Get current Model to check for the next one
	currentModel, err := m.getIdModel(db)
	if err != nil {
		return Model{}, err
	}

	query, err := db.Query("SELECT containers.url, containers.tag FROM containers JOIN models ON models.container = containers.id JOIN analysis ON analysis.next_model = models.id JOIN pipelines ON analysis.pipeline = pipelines.id WHERE pipelines.id = $1 AND analysis.execute = $2", pipelineId, currentModel)
	if err != nil {
		return Model{}, err
	}
	defer func() {
		if err := query.Close(); err != nil {
			klog.Errorf("could not close query: %s\n", err)
		}
	}()

	var url, tag string
	query.Next()
	err = query.Scan(&url, &tag)
	klog.Infof("url is %s and tag is %s\n", url, tag)
	return Model{Url: url, Tag: tag}, err

}

// getIdModel find the id of a model based on the model url and tag
func (m Model) getIdModel(db *sql.DB) (int64, error) {

	klog.V(2).Infof("Query Parameter in getIdModel: URL: %s, Tag: %s\n", m.Url, m.Tag)

	query, err := db.Query("SELECT models.id FROM containers JOIN models on containers.id = models.container WHERE url = $1 AND tag = $2", m.Url, m.Tag)
	if err != nil {
		return -1, err
	}
	defer func() {
		if err := query.Close(); err != nil {
			klog.Errorf("cannot close query object :%s\n", err)
		}
	}()

	var id int64 = -1
	found := false

	for query.Next() {
		found = true
		err = query.Scan(&id)
		if err != nil {
			klog.Errorf("cannot scan err: %s\n", err)
			return -1, err
		}
	}

	if !found {
		return -1, fmt.Errorf("no result found")
	}

	klog.V(2).Infof("id is: %d\n", id)

	return id, nil
}

// GetMachineSensorId find id of the machine_sensor table to a given machine and sensor
func GetMachineSensorId(db *sql.DB, machine, sensor string) (int64, error) {

	query, err := db.Query(
		"SELECT machine_sensors.id FROM machine_sensors JOIN sensors ON sensors.id = machine_sensors.sensor WHERE transmitted_id = $1 AND machine = $2", sensor, machine)

	if err != nil {
		return -1, err
	}

	defer func() {
		if err := query.Close(); err != nil {
			klog.Errorf("cannot close query object :%s\n", err)
		}
	}()

	var id int64
	query.Next()
	err = query.Scan(&id)
	return id, err
}

// GetContract get the contract of a machine and machineSensor combination
func GetContract(db *sql.DB, machine string, machineSensor int64) (string, error) {
	klog.V(2).Infof("query after: machineSensor %d", machineSensor)
	qu, err := db.Query("SELECT contract_machine_sensors.contract FROM contract_machine_sensors WHERE contract_machine_sensors.machine_sensor = $1", machineSensor)
	if err != nil {
		return "", err
	}

	defer func() {
		if err := qu.Close(); err != nil {
			klog.Errorf("cannot close query response: %s\n", err)
		}
	}()

	var contract string

	for qu.Next() {
		if err := qu.Scan(&contract); err != nil {
			return "", err
		}
	}

	return contract, nil
}

// GetContractMachineSensor get the contract machineSensor from a contract and machine sensor combination
func GetContractMachineSensor(db *sql.DB, contractId string, machineSensor int64) (string, error) {

	klog.V(2).Infof("query after: machineSensor %d", machineSensor)

	qu, err := db.Query("SELECT contract_machine_sensors.id FROM contract_machine_sensors WHERE contract_machine_sensors.contract = $1 AND contract_machine_sensors.machine_sensor = $2", contractId, machineSensor)
	if err != nil {
		return "", err
	}

	defer func() {
		if err := qu.Close(); err != nil {
			klog.Errorf("cannot close query response: %s\n", err)
		}
	}()

	var contract string

	for qu.Next() {
		if err := qu.Scan(&contract); err != nil {
			return "", err
		}
	}

	return contract, nil
}

func GetPipelineId(db *sql.DB, contractMachineSensorId string) (string, error) {
	klog.V(2).Infof("query for pipeline after: contractMachineSensorId %s", contractMachineSensorId)
	qu, err := db.Query("SELECT id FROM pipelines WHERE contract_machine_sensor = $1", contractMachineSensorId)
	if err != nil {
		return "", err
	}

	defer func() {
		if err := qu.Close(); err != nil {
			klog.Errorf("cannot close query response: %s\n", err)
		}
	}()

	var pipelineId string

	for qu.Next() {
		if err := qu.Scan(&pipelineId); err != nil {
			return "", err
		}
	}

	return pipelineId, nil
}

func GetUrlAndTagFromContainer(db *sql.DB, containerId string) (string, string, error) {
	klog.V(2).Infof("Query for url and tag after: containerId %s", containerId)
	qu, err := db.Query("SELECT url, tag FROM containers WHERE id = $1", containerId)
	if err != nil {
		return "", "", err
	}

	defer func() {
		if err := qu.Close(); err != nil {
			klog.Errorf("cannot close query response: %s\n", err)
		}
	}()

	var url, tag string

	for qu.Next() {
		if err := qu.Scan(&url, &tag); err != nil {
			return "", "", err
		}
	}

	return url, tag, nil
}
