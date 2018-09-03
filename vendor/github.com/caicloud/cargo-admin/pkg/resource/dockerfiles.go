package resource

import (
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/caicloud/cargo-admin/pkg/api/admin/types"

	"github.com/caicloud/nirvana/log"

	"gopkg.in/yaml.v2"
)

const TemplatePath = "/app/dockerfiles.yml"

var groups []*types.DockerfileGroup
var once sync.Once

func GetDockerfiles() ([]*types.DockerfileGroup, error) {
	once.Do(
		func() {
			log.Infof("read dockerfile templates from %s", TemplatePath)
			data, err := ioutil.ReadFile(TemplatePath)
			if err != nil {
				log.Errorf("read template file %s error: %v", TemplatePath, err)
				return
			}
			result, err := getDockerfiles(data)
			if err != nil {
				log.Errorf("unmarshal yaml template file %s error: %v", TemplatePath, err)
				return
			}
			groups = result
			log.Info("read dockerfile templates succeed")
		},
	)

	if groups == nil || len(groups) == 0 {
		return nil, fmt.Errorf("get dockerfiles error")
	}

	return groups, nil
}

func getDockerfiles(data []byte) ([]*types.DockerfileGroup, error) {
	dockerfiles := make([]*types.DockerfileGroup, 0)
	err := yaml.Unmarshal(data, &dockerfiles)
	if err != nil {
		log.Errorf("unmarshal yaml template file %s error: %v", TemplatePath, err)
		return nil, err
	}

	return dockerfiles, nil
}
