// build +unit
package config

import (
	"strings"
	"testing"

	yaml "gopkg.in/yaml.v2"

	"gitlab.inovex.de/proj-kosmos/intern-mqtt-db/models"
)

func TestConfigPasswords(t *testing.T) {
	var password models.Password
	password.Database.Password = "abc"
	password.Database.User = "bcd"
	password.Mqtt.Password = "pppp"
	password.Mqtt.User = "asdf"

	bytes, err := yaml.Marshal(password)
	var conf models.Password
	if err != nil {
		t.Errorf("could not parse configuration to yaml: %v\n", err)
	}

	if err := handlePassword(strings.NewReader(string(bytes)), &conf); err != nil {
		t.Errorf("could not unparse from yaml to configuration")
	}

	if conf != password {
		t.Errorf("conf != con\n\t%v\n\t%v\n", password, conf)
	}
}

func TestConfigConfiguration(t *testing.T) {
	var configuration models.Configuration
	configuration.Database.Address = "127.0.0.1"
	configuration.Database.Database = "postgres"
	configuration.Database.Port = 789
	configuration.Mqtt.Address = "127.0.0.1"
	configuration.Mqtt.Port = 5432

	bytes, err := yaml.Marshal(configuration)
	var conf models.Configuration
	if err != nil {
		t.Errorf("could not parse configuration to yaml: %v\n", err)
	}

	if err := handleConfiguration(strings.NewReader(string(bytes)), &conf); err != nil {
		t.Errorf("could not unparse from yaml to configuration")
	}

	if conf != configuration {
		t.Errorf("conf != con\n\t%v\n\t%v\n", configuration, conf)
	}
}
