package models

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
