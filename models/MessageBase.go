package models

// MessageTyp is an enum with the values Analyses = 1 and SensorUpdate = 2
type MessageTyp int

// can be used by the enumeration above
const (
	Analyses = 1
	Update   = 2
)

// MessageBase will be used to transmit data between
// the receiving and publishing part of the program
type MessageBase struct {
	Machine      string
	Sensor       string
	LastAnalyses string
	Message      []byte
	Contract     string
	MessageTyp   MessageTyp
	Model        Model
}
