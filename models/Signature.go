package models

type SignatureObject struct {
	Signature string `json:"signature"`
	Meta      struct {
		Algorithm    string `json:"algorithm"`
		Date         string `json:"date"`
		SerialNumber string `json:"serialNumber"`
	} `json:"meta"`
}
