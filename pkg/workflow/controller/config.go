package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	api_v1 "k8s.io/api/core/v1"
)

// The key of config file in ConfigMap
const ConfigFileKey = "workflow-controller.json"

type ControllerConfig struct {
	Logging LoggingConfig `json:"logging"`
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

	InitLogger(&Config.Logging)
	return nil
}

func LoadConfig(configPath *string, config interface{}) error {
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
	}

	return err
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
