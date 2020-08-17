package models

//SendMsg representing the message, which will be published to the mqtt broker
type SendMsg struct {
	Contract string      `json:"contract"`
	Type     string      `json:"type"`
	Payload  interface{} `json:"payload"`
}
