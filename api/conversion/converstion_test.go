/*
Copyright 2017 caicloud authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package conversion

import (
	"testing"

	newapi "github.com/caicloud/cyclone/pkg/api"
)

var Pipeline = &newapi.Pipeline{
	Name:        "pipeline1",
	Description: "first test pipeline",
	Owner:       "robin",
	Build: &newapi.Build{
		BuilderImage: &newapi.BuilderImage{
			Image: "caicloud.io/golang:1.8",
			EnvVars: []newapi.EnvVar{
				newapi.EnvVar{
					Name:  "GOROOT",
					Value: "/usr/local/go",
				},
				newapi.EnvVar{
					Name:  "GOPATH",
					Value: "/Users/robin/gocode",
				},
			},
		},
		Stages: &newapi.BuildStages{
			Package: &newapi.PackageStage{
				GeneralStage: newapi.GeneralStage{
					Command: []string{"go build ."},
				},
				Outputs: []string{"./outputs/cyclone"},
			},
			ImageBuild: &newapi.ImageBuildStage{
				BuildInfos: []*newapi.ImageBuildInfo{
					&newapi.ImageBuildInfo{
						ContextDir:     ".",
						DockerfilePath: "Dockerfile.server",
					},
				},
			},
			IntegrationTest: &newapi.IntegrationTestStage{
				IntegrationTestSet: &newapi.IntegrationTestSet{
					Command: []string{"go test"},
				},
				Services: []newapi.Service{
					newapi.Service{
						Name:    "mongo",
						Image:   "mongo:3.0.5",
						Command: []string{"mongod --smallfiles"},
					},
					newapi.Service{
						Name:  "kafka",
						Image: "wurstmeister/kafka:0.10.1.0",
						EnvVars: []newapi.EnvVar{
							newapi.EnvVar{
								Name:  "KAFKA_ADVERTISED_HOST_NAME",
								Value: "kafkahost",
							},
							newapi.EnvVar{
								Name:  "KAFKA_ADVERTISED_PORT",
								Value: "9092",
							},
							newapi.EnvVar{
								Name:  "KAFKA_LOG_DIRS",
								Value: "/data/kafka_log",
							},
						},
					},
				},
			},
		},
	},
}

var expectedYMALStr = `pre_build:
  image: caicloud.io/golang:1.8
  environment:
  - GOROOT=/usr/local/go
  - GOPATH=/Users/robin/gocode
  commands:
  - go build .
  outputs:
  - ./outputs/cyclone
build:
  context_dir: .
  dockerfile_name: Dockerfile.server
integration:
  services:
    kafka:
      image: wurstmeister/kafka:0.10.1.0
      environment:
      - KAFKA_ADVERTISED_HOST_NAME=kafkahost
      - KAFKA_ADVERTISED_PORT=9092
      - KAFKA_LOG_DIRS=/data/kafka_log
    mongo:
      image: mongo:3.0.5
      commmands:
      - mongod --smallfiles
  commands:
  - go test
`

func TestConvertPipelineToService(t *testing.T) {
}

func TestConvertBuildStagesToCaicloudYaml(t *testing.T) {
	t.Log("hello test")

	config, err := convertBuildStagesToCaicloudYaml(Pipeline)
	if err != nil {
		t.Errorf("Fail to convert to caicloud yaml as expect error is nil but got : %s", err.Error())
	}

	if config != expectedYMALStr {
		t.Errorf("The converted caicloud yaml is not correct: \n%s\nexpected: \n%s\n", config, expectedYMALStr)
	}
}
