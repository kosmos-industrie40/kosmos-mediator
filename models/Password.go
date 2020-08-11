// Package models containing data models, which are used in different parts of the programm
package models

// Password defines the internal representation of the password configuration file
type Password struct {
	Mqtt struct {
		User     string `yaml:"user"`
		Password string `yaml:"password"`
	} `yaml:"mqtt"`
	Database struct {
		User     string `yaml:"user"`
		Password string `yaml:"password"`
	} `yaml:"database"`
}
