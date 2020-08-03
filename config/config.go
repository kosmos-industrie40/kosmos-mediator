package config

import (
	"io"
	"os"

	yaml "gopkg.in/yaml.v2"
	"k8s.io/klog"

	"gitlab.inovex.de/proj-kosmos/intern-mqtt-db/models"
)

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

func handleConfiguration(handle io.Reader, conf *models.Configuration) error {
	decoder := yaml.NewDecoder(handle)
	decoder.SetStrict(true)

	if err := decoder.Decode(conf); err != nil {
		return err
	}

	return nil
}

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

func handlePassword(handle io.Reader, password *models.Password) error {
	decoder := yaml.NewDecoder(handle)
	decoder.SetStrict(true)

	if err := decoder.Decode(password); err != nil {
		return err
	}

	return nil
}
