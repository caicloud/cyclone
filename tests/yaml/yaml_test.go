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

package yaml

import (
	"time"

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/docker"
	"github.com/caicloud/cyclone/pkg/log"
	"github.com/caicloud/cyclone/pkg/osutil"
	"github.com/caicloud/cyclone/pkg/wait"
	"github.com/caicloud/cyclone/utils"

	. "github.com/caicloud/cyclone/tests/common"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	DefaultServiceName            = "test-basic-rest-version"
	DefaultVersionName            = "v0.1.0"
	DefaultIntegrationVersionName = "v0.1.0-integration"
	// ServiceImage is the image used to test yaml function.
	ServiceImage                 = "minimal-long-running-task:latest"
	DefaultVCS                   = api.Git
	DefaultServiceTestDeployName = "test-deploy-service"
	DefaultDeployVersionName     = "v0.1.0-deploy"
	DefaultTestDeployRepo        = "https://github.com/gaocegege/circle-deploy"
	DefaultTestIntegrationRepo   = "https://github.com/superxi911/circle-integration"
)

// TODO: Test whether the repo has been bound into the container in the integration step.
var _ = Describe("Yaml", func() {
	var (
		dockerManager         *docker.Manager
		service               *api.Service
		serviceResponse       *api.ServiceCreationResponse
		serviceTestDeploy     *api.Service
		serviceTestDeployResp *api.ServiceCreationResponse
	)

	// Set up the serviceCM, versionCM and dockerManager and create a service.
	BeforeSuite(func() {
		// Wait cyclone to start.
		WaitComponents()

		// Get docker host and cert path.
		endpoint := osutil.GetStringEnv("DOCKER_HOST", DefaultDockerHost)
		certPath := osutil.GetStringEnv("DOCKER_CERT_PATH", "")
		jenkins := osutil.GetStringEnv("JENKINS", "no")

		// Get the username and password to access the docker registry.
		registryLocation := osutil.GetStringEnv("REGISTRY_LOCATION", DefaultRegistryAddress)
		registryUsername := osutil.GetStringEnv("REGISTRY_USERNAME", AdminUser)
		registryPassword := osutil.GetStringEnv("REGISTRY_PASSWORD", AdminPassword)
		registry := api.RegistryCompose{
			RegistryLocation: registryLocation,
			RegistryUsername: registryUsername,
			RegistryPassword: registryPassword,
		}

		// Create docker manager.
		var err error
		dockerManager, err = docker.NewManager(
			endpoint, certPath, registry)
		if err != nil {
			log.Fatalf("Unable to connect to docker daemon: %v", err)
		}

		// If e2e is running on jenkins, make some differences
		if jenkins == "yes" || jenkins == "YES" {
			log.Infof("Pusing %s to %s.", ServiceImage, dockerManager.Registry)
			if err = PushImageToLocalRegistry(dockerManager, ServiceImage); err != nil {
				log.Fatalf("Unable to push %s to %s: %v", ServiceImage, dockerManager.Registry, err)
			}
		}

		// Create a test service.
		service = &api.Service{
			Name:        DefaultServiceName,
			Username:    AliceUser,
			Description: "A service just for test",
			Repository: api.ServiceRepository{
				URL: DefaultTestIntegrationRepo,
				Vcs: DefaultVCS,
			},
		}
		serviceResponse = &api.ServiceCreationResponse{}
		err = CreateService(AliceUID, service, serviceResponse)
		if err != nil {
			log.Errorf("unexpected error while creating service for user %v: %v", AliceUID, err)
		}
		// Wait up to 120 seconds until the service is successfully created.
		err = wait.Poll(3*time.Second, 120*time.Second, func() (bool, error) {
			serviceGetResponse := &api.ServiceGetResponse{}
			err := GetService(AliceUID, serviceResponse.ServiceID, serviceGetResponse)
			return serviceGetResponse.Service.Repository.Status == api.RepositoryHealthy, err
		})
		if err != nil {
			log.ErrorWithFields("unexpected repository status", log.Fields{"repository": service.Repository, "status": service.Repository.Status})
		}

		// Create one more service to test deploy.
		serviceTestDeploy = &api.Service{
			Name:        DefaultServiceTestDeployName,
			Username:    utils.DeployUser,
			Description: "A service for test deploy version",
			Repository: api.ServiceRepository{
				URL: DefaultTestDeployRepo,
				Vcs: DefaultVCS,
			},
		}
		serviceTestDeployResp = &api.ServiceCreationResponse{}
		err = CreateService(utils.DeployUID, serviceTestDeploy, serviceTestDeployResp)
		if err != nil {
			log.Errorf("unexpected error while creating service for user %v: %v", AliceUID, err)
		}

		// Wait up to 120 seconds until the service is successfully created.
		err = wait.Poll(3*time.Second, 120*time.Second, func() (bool, error) {
			serviceGetResponse := &api.ServiceGetResponse{}
			err := GetService(utils.DeployUID, serviceTestDeployResp.ServiceID, serviceGetResponse)
			return serviceGetResponse.Service.Repository.Status == api.RepositoryHealthy, err
		})
		Expect(err).To(BeNil())
	})

	Context("with right version information", func() {
		Context("with the default operation", func() {
			var (
				versionResponse *api.VersionCreationResponse
			)
			It("should create version successfully with right information.", func() {
				// The version will use a image named "localhost:5000/alice/test-basic-rest-version:v0.1.0"
				// in the integration step. And the image will be pushed to local registry in the version test,
				// So the integration test MUST be executed after the version test.
				version := &api.Version{
					Name:        DefaultVersionName,
					Description: "A version just for integration test",
					ServiceID:   serviceResponse.ServiceID,
					Operation:   api.IntegrationOperation + api.PublishOperation,
				}
				// Create the version.
				versionResponse = &api.VersionCreationResponse{}
				Expect(CreateVersion(AliceUID, version, versionResponse)).To(BeNil())
				Expect(versionResponse.ErrorMessage).To(Equal(""))
			})

			It("should be able to get version via HTTP GET method.", func() {
				versionGetResponse := &api.VersionGetResponse{}
				// Wait up to 60 seconds until docker image has been pushed.
				err := wait.Poll(2*time.Second, 300*time.Second, func() (bool, error) {
					err := GetVersion(AliceUID, versionResponse.VersionID, versionGetResponse)
					return versionGetResponse.Version.Status != api.VersionPending && versionGetResponse.Version.Status != api.VersionRunning, err
				})
				Expect(err).To(BeNil())
				Expect(versionGetResponse.ErrorMessage).To(Equal(""))
				Expect(versionGetResponse.Version).NotTo(BeNil())
				Expect(versionGetResponse.Version.Status).To(Equal(api.VersionHealthy))
			})
		})

		Context("with the integration operation", func() {
			var (
				versionResponse *api.VersionCreationResponse
			)

			It("should create version successfully.", func() {
				version := &api.Version{
					Name:        DefaultIntegrationVersionName,
					Description: "A version just for test",
					ServiceID:   serviceResponse.ServiceID,
					Operation:   api.IntegrationOperation,
				}

				imageName := dockerManager.GetImageNameWithTag(AliceUser, DefaultServiceName, DefaultVersionName)
				// Forcibly remove image.
				dockerManager.RemoveImage(imageName)
				// Create the version.
				versionResponse = &api.VersionCreationResponse{}

				Expect(CreateVersion(AliceUID, version, versionResponse)).To(BeNil())
				Expect(versionResponse.ErrorMessage).To(Equal(""))
			})

			It("should be able to get version via HTTP GET method.", func() {
				versionGetResponse := &api.VersionGetResponse{}
				// Wait up to 30 seconds until docker image has been pushed.
				err := wait.Poll(2*time.Second, 300*time.Second, func() (bool, error) {
					err := GetVersion(AliceUID, versionResponse.VersionID, versionGetResponse)
					return versionGetResponse.Version.Status != api.VersionPending && versionGetResponse.Version.Status != api.VersionRunning, err
				})
				Expect(err).To(BeNil())
				Expect(versionGetResponse.ErrorMessage).To(Equal(""))
				Expect(versionGetResponse.Version).NotTo(BeNil())
				Expect(versionGetResponse.Version.Status).To(Equal(api.VersionHealthy))
			})

			It("should NOT pull newly created docker image successfully.", func() {
				imageName := dockerManager.GetImageNameWithTag(AliceUser, DefaultServiceName, DefaultIntegrationVersionName)
				Expect(dockerManager.PullImage(imageName)).NotTo(BeNil())
			})
		})

		Context("with the deploy operation", func() {

			var (
				versionResponse *api.VersionCreationResponse
			)

			It("should create version successfully.", func() {
				version := &api.Version{
					Name:        DefaultDeployVersionName,
					Description: "A version just for test deploy",
					ServiceID:   serviceTestDeployResp.ServiceID,
					Operation:   api.PublishOperation + api.DeployOperation,
					YamlDeploy:  api.DeployWithYaml,
				}

				// Create the version.
				versionResponse = &api.VersionCreationResponse{}

				Expect(CreateVersion(utils.DeployUID, version, versionResponse)).To(BeNil())
				Expect(versionResponse.ErrorMessage).To(Equal(""))
			})

			It("should be able to get version via HTTP GET method. ", func() {
				versionGetResponse := &api.VersionGetResponse{}
				// Wait up to 30 seconds until docker image has been pushed.
				err := wait.Poll(2*time.Second, 200*time.Second, func() (bool, error) {
					err := GetVersion(utils.DeployUID, versionResponse.VersionID, versionGetResponse)
					done := versionGetResponse.Version.YamlDeployStatus != api.DeployNoRun &&
						versionGetResponse.Version.Status != api.VersionPending && versionGetResponse.Version.Status != api.VersionRunning
					return done, err
				})
				Expect(err).To(BeNil())
				Expect(versionGetResponse.Version).NotTo(BeNil())
				Expect(versionGetResponse.ErrorMessage).To(Equal(""))
				Expect(versionGetResponse.Version.Status).To(Equal(api.VersionHealthy))
				// Mock to DeploySuccess.
				// Expect(versionGetResponse.Version.YamlDeployStatus).To(Equal(api.DeployPending))

				// wait check goroutine to write DB
				err = wait.Poll(2*time.Second, 300*time.Second, func() (bool, error) {
					err := GetVersion(utils.DeployUID, versionResponse.VersionID, versionGetResponse)
					done := versionGetResponse.Version.YamlDeployStatus == api.DeploySuccess
					return done, err

				})
				Expect(err).To(BeNil())
			})
		})
	})
})
