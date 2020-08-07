// Package logic handles the logic of the programm.
// It will register handler on mqtt topics and contains the logic of the
// mediator
package logic

type MessageBase struct {
	Machine      string
	Sensor       string
	LastAnalyses string
}
