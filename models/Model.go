package models

import (
	"database/sql"

	"k8s.io/klog"
)

// Model representing the database table model
type Model struct {
	Url string `json:"url"`
	Tag string `json:"tag"`
}

// InitialPipeline find the initial model in the pipeline
func (m Model) InitialPipeline(db *sql.DB, machine, sensor string) (Model, error) {
	id, err := GetMachineSensorId(db, machine, sensor)
	if err != nil {
		return Model{}, err
	}
	klog.V(2).Infof("id of machine-sensor %d", id)

	contract, err := GetContract(db, machine, id)
	if err != nil {
		return Model{}, err
	}
	klog.V(2).Infof("id of the matching contract: %s\n", contract)

	klog.V(2).Infof("parameters of the next query: contract = %s; machine_sensor: %d\n", contract, id)
	res, err := db.Query("SELECT url, tag FROM model JOIN next_analyses on model.id = next_analyses.next_model JOIN pipeline on next_analyses.id = pipeline.analyses WHERE contract = $1 AND machine_sensor = $2 AND previous_model is NULL", contract, id)
	if err != nil {
		return Model{}, err
	}
	defer func() {
		if err := res.Close(); err != nil {
			klog.Errorf("could not close result query: %s\n", err)
		}
	}()

	var url, tag string

	res.Next()
	err = res.Scan(&url, &tag)
	return Model{Url: url, Tag: tag}, err
}

// TestEnd test if the last analyses in the pipeline has be made
func (m Model) TestEnd(db *sql.DB, machine, sensor, contract string) (bool, error) {
	machSens, err := GetMachineSensorId(db, machine, sensor)
	if err != nil {
		klog.V(2).Infof("error in getMachineSensorId")
		return false, err
	}
	klog.V(2).Infof("machine-sensor id is %d\n", machSens)

	prevModel, err := m.getIdModel(db)
	if err != nil {
		return false, err
	}

	klog.V(2).Infof("parameter to use in query are: contract: %s; machine-sensor: %d, prevModel: %d", contract, machSens, prevModel)
	res, err := db.Query("SELECT EXISTS (SELECT next_model FROM next_analyses JOIN pipeline ON pipeline.analyses = next_analyses.id WHERE pipeline.contract = $1 AND next_analyses.machine_sensor = $2 AND next_analyses.previous_model = $3)", contract, machSens, prevModel)
	if err != nil {
		return false, err
	}

	defer func() {
		if err := res.Close(); err != nil {
			klog.Errorf("can not close rows: %s\n", err)
		}
	}()

	var ex bool

	for res.Next() {
		if err := res.Scan(&ex); err != nil {
			klog.Errorf("cannot scan err is: %s\n", err)
		}

	}
	return ex, nil
}

// Next query the next model based on the current model
func (m Model) Next(db *sql.DB, machine, sensor, contract string) (Model, error) {

	machSens, err := GetMachineSensorId(db, machine, sensor)
	if err != nil {
		return Model{}, err
	}

	prevModel, err := m.getIdModel(db)
	if err != nil {
		return Model{}, err
	}

	query, err := db.Query("SELECT model.url, model.tag FROM model JOIN next_analyses ON next_analyses.next_model = model.id JOIN pipeline ON pipeline.analyses = next_analyses.id WHERE pipeline.contract = $1 AND next_analyses.machine_sensor = $2 AND next_analyses.previous_model = $3", contract, machSens, prevModel)
	if err != nil {
		return Model{}, err
	}
	defer func() {
		if err := query.Close(); err != nil {
			klog.Errorf("could not close query: %s\n", err)
		}
	}()

	var url, tag string
	klog.Infof("url is %s and tag is %s\n", url, tag)

	query.Next()
	err = query.Scan(&url, &tag)
	klog.Infof("url is %s and tag is %s\n", url, tag)
	return Model{Url: url, Tag: tag}, err
}

// getIdModel find the id of a model based on the model url and tag
func (m Model) getIdModel(db *sql.DB) (int64, error) {

	klog.V(2).Infof("Query Parameter in getIdmodel: URL: %s, Tag: %s\n", m.Url, m.Tag)

	query, err := db.Query("SELECT id FROM model WHERE url = $1 AND tag = $2", m.Url, m.Tag)
	if err != nil {
		return -1, err
	}
	defer func() {
		if err := query.Close(); err != nil {
			klog.Errorf("cannot close query object :%s\n", err)
		}
	}()

	var id int64
	for query.Next() {
		err = query.Scan(&id)
		if err != nil {
			klog.Errorf("cannot scan err: %s\n", err)
		}
	}

	klog.V(2).Infof("id is: %d\n", id)

	return id, nil
}

// GetMachineSensorId find id of the machine_sensor table to a given machine and sensor
func GetMachineSensorId(db *sql.DB, machine, sensor string) (int64, error) {
	query, err := db.Query("SELECT machine_sensor.id FROM machine_sensor JOIN sensor ON sensor.id = machine_sensor.sensor WHERE transmitted_id = $1 AND machine = $2", sensor, machine)
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
	klog.V(2).Infof("query after: machine %s and machineSensor %d", machine, machineSensor)
	qu, err := db.Query("SELECT contract.id FROM contract JOIN machine_contract ON contract.id = machine_contract.contract JOIN machine_sensor ON machine_contract.machine = machine_sensor.machine WHERE machine_contract.machine = $1 AND machine_sensor.id = $2", machine, machineSensor)
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
