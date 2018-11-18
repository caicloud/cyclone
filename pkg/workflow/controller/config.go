package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	api_v1 "k8s.io/api/core/v1"
)

const (
	// The key of config file in ConfigMap
	ConfigFileKey = "workflow-controller.json"

	// Keys of resources images and coordinator image in config file
	GitResolverImage   = "git-resolver"
	ImageResolverImage = "image-resolver"
	KvResolverImage    = "kv-resolver"
	CoordinatorImage   = "coordinator"
)

type ControllerConfig struct {
	// Images that used in controller, such as resource resolvers.
	Images map[string]string `json:"images"`
	// Logging configuration, such as log level.
	Logging LoggingConfig `json:"logging"`
	// The PVC used to transfer artifacts in WorkflowRun
	PVC string `json:"pvc"`
}

type LoggingConfig struct {
	Level string `json:"level"`
}

var Config ControllerConfig

func ReloadConfig(cm *api_v1.ConfigMap) error {
	data, ok := cm.Data[ConfigFileKey]
	if !ok {
		fmt.Errorf("ConfigMap '%s' doesn't have data key '%s'", cm.Name, ConfigFileKey)
	}
	err := json.Unmarshal([]byte(data), &Config)
	if err != nil {
		log.WithField("data", data).Debug("Unmarshal config data error: ", err)
		return err
	}

	if !validate(&Config) {
		return fmt.Errorf("validate config failed")
	}

	InitLogger(&Config.Logging)
	return nil
}

func LoadConfig(configPath *string, config *ControllerConfig) error {
	log.WithField("file", *configPath).Info("Start load configure file")

	data, err := ioutil.ReadFile(*configPath)
	if err != nil {
		log.Error("Load conf failed: ", err)
		return err
	}
	data = trimComments(data)

	err = json.Unmarshal(data, config)
	if err != nil {
		log.Errorf("Parse config error: ", err)
		return err
	}

	if !validate(&Config) {
		return fmt.Errorf("validate config failed")
	}

	return nil
}

func trimComments(data []byte) (data1 []byte) {

	var line []byte

	data1 = data[:0]
	for {
		pos := bytes.IndexByte(data, '\n')
		if pos < 0 {
			line = data
		} else {
			line = data[:pos+1]
		}
		data1 = append(data1, trimCommentsLine(line)...)
		if pos < 0 {
			return
		}
		data = data[pos+1:]
	}
}

func trimCommentsLine(line []byte) []byte {

	n := len(line)
	quoteCount := 0
	for i := 0; i < n; i++ {
		c := line[i]
		switch c {
		case '\\':
			i++
		case '"':
			quoteCount++
		case '#':
			if (quoteCount & 1) == 0 {
				return line[:i]
			}
		}
	}
	return line
}

// validate validates some required configurations.
func validate(config *ControllerConfig) bool {
	for _, k := range []string{GitResolverImage, ImageResolverImage, KvResolverImage, CoordinatorImage} {
		_, ok := config.Images[k]
		if !ok {
			return false
		}
	}

	return true
}
