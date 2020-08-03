package models

type Configuration struct {
	Database struct {
		Address  string `yaml:"address"`
		Port     int    `yaml:"port"`
		Database string `yaml:"database"`
	} `yaml:"database"`
	Mqtt struct {
		Address string `yaml:"address"`
		Port    int    `yaml:"port"`
		Tls     bool   `yaml:"tls"`
	} `yaml:"mqtt"`
	Webserver struct {
		Address string `yaml:"address"`
		Port    int    `yaml:"port"`
	} `yaml:"webserver"`
}
