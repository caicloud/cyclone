/*
Copyright 2016 caicloud authors. All rights reserved.

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

package stage_test

import (
	"flag"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/docker"
	"github.com/caicloud/cyclone/pkg/worker/cycloneserver"
	"github.com/caicloud/cyclone/pkg/worker/scm"
	_ "github.com/caicloud/cyclone/pkg/worker/scm/provider"
	. "github.com/caicloud/cyclone/pkg/worker/stage"
)

const (
	logDir    = "/tmp/logs"
	codeDir   = "/tmp/code"
	githubURL = "https://github.com/caicloud/toy-dockerfile.git"
)

var dockerManager *docker.DockerManager

func TestStage(t *testing.T) {
	// Log to standard error instead of files.
	flag.Set("logtostderr", "true")

	RegisterFailHandler(Fail)
	RunSpecs(t, "Stage Suite")
}

var _ = BeforeSuite(func() {
	dm, err := docker.NewDockerManager("unix:////Users/robin/Library/Containers/com.docker.docker/Data/s60", "", "", "")
	Expect(err).NotTo(HaveOccurred())

	dockerManager = dm
})

var _ = Describe("Stage", func() {
	var (
		stageManager StageManager
		event        *api.Event
	)

	BeforeEach(func() {
		client := cycloneserver.NewFakeClient("http://fack-server.cyclone.com")
		stageManager = NewStageManager(dockerManager, client)
		event = &api.Event{
			ID:       "1",
			Project:  &api.Project{},
			Pipeline: &api.Pipeline{},
			PipelineRecord: &api.PipelineRecord{
				StageStatus: &api.StageStatus{},
			},
		}
		stageManager.SetEvent(event)
	})

	Describe("Code checkout stage", func() {
		BeforeEach(func() {
			// Cleanup the temp folder.
			os.RemoveAll(codeDir)
			os.RemoveAll(logDir)
			os.MkdirAll(logDir, os.ModePerm)
		})

		AfterEach(func() {
			// Cleanup the temp folder.
			os.RemoveAll(codeDir)
			os.RemoveAll(logDir)
		})

		Context("with git code source", func() {
			stage := &api.CodeCheckoutStage{
				CodeSources: []*api.CodeSource{
					&api.CodeSource{
						Type: api.GitHub,
						Main: true,
						GitHub: &api.GitSource{
							Url: githubURL,
						},
					},
				},
			}
			It("t1", func() {
				err := stageManager.ExecCodeCheckout("", stage)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Image build stage", func() {
		BeforeEach(func() {
			// Cleanup the temp folder.
			os.RemoveAll(logDir)
			os.MkdirAll(logDir, os.ModePerm)

			// Prepare the code dir.
			os.RemoveAll(codeDir)
			cs := &api.CodeSource{
				Type: api.GitHub,
				Main: true,
				GitHub: &api.GitSource{
					Url: githubURL,
				},
			}
			scm.CloneRepo("", cs)
		})

		AfterEach(func() {
			// Cleanup the temp folder.
			os.RemoveAll(logDir)
		})

		Context("for one image", func() {
			Context("from", func() {
				stage := &api.ImageBuildStage{
					BuildInfos: []*api.ImageBuildInfo{
						&api.ImageBuildInfo{
							ImageName: "cargo.caicloud.io/caicloud/test:v1",
						},
					},
				}

				It("default Dockerfile path", func() {
					images, err := stageManager.ExecImageBuild(stage)
					Expect(err).NotTo(HaveOccurred())
					Expect(len(images)).Should(Equal(len(stage.BuildInfos)))
					Expect(images[0]).Should(Equal(stage.BuildInfos[0].ImageName))
				})

				It("correct Dockerfile path", func() {
					stage.BuildInfos[0].Dockerfile = "Dockerfile"
					images, err := stageManager.ExecImageBuild(stage)
					Expect(err).NotTo(HaveOccurred())
					Expect(len(images)).Should(Equal(len(stage.BuildInfos)))
					Expect(images[0]).Should(Equal(stage.BuildInfos[0].ImageName))
				})

				It("wrong Dockerfile path", func() {
					stage.BuildInfos[0].Dockerfile = "Dockerfile-fake"
					_, err := stageManager.ExecImageBuild(stage)
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})

	Describe("Integration test stage", func() {
		BeforeEach(func() {
			// Cleanup the temp folder.
			os.RemoveAll(logDir)
			os.MkdirAll(logDir, os.ModePerm)
		})

		AfterEach(func() {
			// Cleanup the temp folder.
			os.RemoveAll(logDir)
		})

		Context("without services", func() {
			Context("for", func() {
				stage := &api.IntegrationTestStage{
					Config: &api.IntegrationTestConfig{
						Command:   []string{"echo hello", "echo integration test"},
						ImageName: "cargo.caicloud.io/caicloud/busybox",
					},
				}

				It("single built image", func() {
					builtImages := []string{"cargo.caicloud.io/caicloud/busybox:latest"}
					err := stageManager.ExecIntegrationTest(builtImages, stage)
					Expect(err).NotTo(HaveOccurred())

					stage.Config.ImageName = "cargo.caicloud.io/caicloud/testabc"
					err = stageManager.ExecIntegrationTest(builtImages, stage)
					Expect(err).Should(HaveOccurred())
				})
			})
		})

		Context("with services", func() {
			Context("for", func() {
				stage := &api.IntegrationTestStage{
					Config: &api.IntegrationTestConfig{
						Command:   []string{"echo hello", "echo integration test"},
						ImageName: "cargo.caicloud.io/caicloud/busybox",
					},
					Services: []api.Service{
						api.Service{
							Name:    "mongo",
							Image:   "mongo:3.0.5",
							Command: []string{"mongod", "--smallfiles"},
						},
					},
				}

				It("single built image", func() {
					builtImages := []string{"cargo.caicloud.io/caicloud/busybox:latest"}
					err := stageManager.ExecIntegrationTest(builtImages, stage)
					Expect(err).NotTo(HaveOccurred())

					stage.Config.ImageName = "cargo.caicloud.io/caicloud/testabc"
					err = stageManager.ExecIntegrationTest(builtImages, stage)
					Expect(err).Should(HaveOccurred())
				})
			})
		})
	})
})
