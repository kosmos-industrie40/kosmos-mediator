package config

import (
	"io"
	"os"

	yaml "gopkg.in/yaml.v2"
	"k8s.io/klog"

	"gitlab.inovex.de/proj-kosmos/intern-mqtt-db/models"
)

// ParseConfiguration parse a yaml file and returns the configuration in the configuration data model.
func ParseConfigurations(path string, configurations *models.Configuration) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	defer func() {
		err := file.Close()
		if err != nil {
			klog.Errorf("could not close file; err: %v", err)
		}
	}()

	return handleConfiguration(file, configurations)
}

// handleConfiguration is used to provide a better possibility to test this functionality
// (not mocking file open operations)
// this function will decode the open file to the configuration data model
func handleConfiguration(handle io.Reader, conf *models.Configuration) error {
	decoder := yaml.NewDecoder(handle)
	decoder.SetStrict(true)

	if err := decoder.Decode(conf); err != nil {
		return err
	}

	return nil
}

// ParsePassword parse a yaml file and returns the passowrd-user combinations in the passsword data model.
func ParsePassword(path string, password *models.Password) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	defer func() {
		err := file.Close()
		if err != nil {
			klog.Errorf("could not close file; err: %v", err)
		}
	}()

	return handlePassword(file, password)
}

// handlePassword is used to provide a better possibility to test this functionality
// (not mocking file open operations)
// this function will decode the open file to the password data model
func handlePassword(handle io.Reader, password *models.Password) error {
	decoder := yaml.NewDecoder(handle)
	decoder.SetStrict(true)

	if err := decoder.Decode(password); err != nil {
		return err
	}

	return nil
}
