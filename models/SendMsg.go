package models

//SendMsg representing the message, which will be published to the mqtt broker
type SendMsg struct {
	Body SendBody `json:"body"`
}

// SendBody is the body in a message
type SendBody struct {
	Contract string      `json:"contract"`
	Type     string      `json:"type"`
	Machine  string      `json:"machine"`
	Sensor   string      `json:"sensor"`
	Payload  interface{} `json:"payload"`
}
