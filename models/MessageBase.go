package models

// MessageBase will be used to transmit data between
// the receiving and publishing part of the program
type MessageBase struct {
	Machine      string
	Sensor       string
	LastAnalyses string
	Message      []byte
	Contract     string
}
