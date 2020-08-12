package models

import (
	"database/sql"
)

// Model representing the database table model
type Model struct {
	Url string `json:"url"`
	Tag string `json:"tag"`
}

func (m Model) TestEnd(db *sql.DB, machine, sensor, contract string) (bool, error) {
	machSens, err := getMachineSensorId(db, machine, sensor)
	if err != nil {
		return false, err
	}

	prevModel, err := m.getIdModel(db)
	if err != nil {
		return false, err
	}
	query, err := db.Query("SELECT EXISTS (SELECT 1 FROM model JOIN next_analyses ON next_analyses.next_model = model.id JOIN pipeline ON pipeline.next_analyses = next_analyses.id WHERE pipeline.contract = $1 AND next_analyses.machine_sensor = $2 AND next_analyses.previous_model = $3", contract, machSens, prevModel)
	if err != nil {
		return false, err
	}

	var exists bool

	query.Next()
	err = query.Scan(&exists)
	return exists, err
}

// Next query the next model based on the current model
func (m Model) Next(db *sql.DB, machine, sensor, contract string) (Model, error) {

	machSens, err := getMachineSensorId(db, machine, sensor)
	if err != nil {
		return Model{}, err
	}

	prevModel, err := m.getIdModel(db)
	if err != nil {
		return Model{}, err
	}

	query, err := db.Query("SELECT model.url, model.tag FROM model JOIN next_analyses ON next_analyses.next_model = model.id JOIN pipeline ON pipeline.next_analyses = next_analyses.id WHERE pipeline.contract = $1 AND next_analyses.machine_sensor = $2 AND next_analyses.previous_model = $3", contract, machSens, prevModel)
	if err != nil {
		return Model{}, err
	}

	var url, tag string

	query.Next()
	err = query.Scan(&url, &tag)
	return Model{Url: url, Tag: tag}, err
}

// getIdModel find the id of a model based on the model url and tag
func (m Model) getIdModel(db *sql.DB) (int64, error) {
	query, err := db.Query("SELECT id FROM model WHERE url = $1 AND tag = $2", m.Url, m.Tag)
	if err != nil {
		return -1, err
	}

	var id int64
	query.Next()
	err = query.Scan(&id)

	return id, err
}

// getMachineSensorId find id of the machine_sensor table to a given machine and sensor
func getMachineSensorId(db *sql.DB, machine, sensor string) (int64, error) {
	query, err := db.Query("SELECT machine_sensor.id FROM machine_sensor JOIN sensor ON sensor.id = machine_sensor.sensor WHERE transmitted_id = $1 AND machine = $2", sensor, machine)
	if err != nil {
		return -1, err
	}

	var id int64
	query.Next()
	err = query.Scan(&id)
	return id, err
}
