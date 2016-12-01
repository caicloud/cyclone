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

package service

import (
	"time"

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/docker"
	"github.com/caicloud/cyclone/pkg/log"
	"github.com/caicloud/cyclone/pkg/osutil"
	"github.com/caicloud/cyclone/pkg/wait"

	. "github.com/caicloud/cyclone/tests/common"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	DefaultServiceName = "test-basic-rest-service"
	DefaultTestRepo    = "https://github.com/caicloud/toy-dockerfile"
)

var _ = Describe("Service", func() {

	// Set up the serviceCM, versionCM and dockerManager.
	BeforeSuite(func() {
		var err error

		log.Info("Wait")
		// Wait cyclone to start.
		WaitComponents()
		err = RegisterResource()
		if err != nil {
			log.Fatalf("Unable to register resources: %v", err)
		}

		// Get docker host and cert path.
		endpoint := osutil.GetStringEnv("DOCKER_HOST", DefaultDockerHost)
		certPath := osutil.GetStringEnv("DOCKER_CERT_PATH", "")

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
		_, err = docker.NewManager(
			endpoint, certPath, registry)
		if err != nil {
			log.Fatalf("Unable to connect to docker daemon: %v", err)
		}
	})

	It("should be a available.", func() {
		Expect(IsAvailable()).To(Equal(true))
	})

	Context("with right information", func() {
		service := &api.Service{
			Name:        DefaultServiceName,
			Username:    ListUser,
			Description: "A service just for test",
			Repository: api.ServiceRepository{
				URL: DefaultTestRepo,
				Vcs: api.Git,
			},
		}
		serviceResponse := &api.ServiceCreationResponse{}

		It("should create service successfully.", func() {
			Expect(CreateService(ListUID, service, serviceResponse)).To(BeNil())
			Expect(serviceResponse.ErrorMessage).To(Equal(""))
		})

		It("should be got by HTTP GET method.", func() {
			// Wait up to 120 seconds until the service is successfully created.
			err := wait.Poll(2*time.Second, 120*time.Second, func() (bool, error) {
				serviceGetResponse := &api.ServiceGetResponse{}
				err := GetService(ListUID, serviceResponse.ServiceID, serviceGetResponse)
				return serviceGetResponse.Service.Repository.Status == api.RepositoryHealthy, err
			})
			Expect(err).To(BeNil())
		})

		It("should be listed by HTTP GET method.", func() {
			// List services to check list endpoint works correctly.
			serviceListResponse := &api.ServiceListResponse{}
			Expect(ListServices(ListUID, serviceListResponse)).To(BeNil())
			Expect(serviceResponse.ErrorMessage).To(Equal(""))
			Expect(len(serviceListResponse.Services)).To(Equal(1))
			Expect(serviceListResponse.Services[0].Name).To(Equal(DefaultServiceName))
			Expect(serviceListResponse.Services[0].Repository.Status).To(Equal(api.RepositoryHealthy))
		})
	})

	Context("with the same service name.", func() {
		service := &api.Service{
			Name:        DefaultServiceName,
			Username:    AliceUser,
			Description: "A service just for test",
			Repository: api.ServiceRepository{
				URL: DefaultTestRepo,
				Vcs: api.Git,
			},
		}
		serviceResponse := &api.ServiceCreationResponse{}
		It("should create service failed.", func() {
			Expect(CreateService(ListUID, service, serviceResponse)).NotTo(BeNil())
		})
	})

	Context("with wrong repository url", func() {
		service := &api.Service{
			Name:        "cyclone",
			Username:    AliceUser,
			Description: "A service just for test",
			Repository: api.ServiceRepository{
				URL: "123456",
				Vcs: api.Git,
			},
		}
		serviceResponse := &api.ServiceCreationResponse{}
		It("should create service successfully but with repository status set to 'missing'.", func() {
			Expect(CreateService(AliceUID, service, serviceResponse)).To(BeNil())
			// Wait up to 120 seconds until the service is successfully created.
			err := wait.Poll(2*time.Second, 120*time.Second, func() (bool, error) {
				serviceGetResponse := &api.ServiceGetResponse{}
				err := GetService(AliceUID, serviceResponse.ServiceID, serviceGetResponse)
				return serviceGetResponse.Service.Repository.Status == api.RepositoryMissing, err
			})
			Expect(err).To(BeNil())
		})
	})

	Context("with wrong vcs tool", func() {
		service := &api.Service{
			Name:        "clever",
			Username:    AliceUser,
			Description: "A service just for test",
			Repository: api.ServiceRepository{
				URL: DefaultTestRepo,
				Vcs: "gaocegege",
			},
		}
		serviceResponse := &api.ServiceCreationResponse{}

		It("should create service successfully with repo status set to 'unknownvcs'.", func() {
			Expect(CreateService(AliceUID, service, serviceResponse)).To(BeNil())
			// Wait up to 200 seconds until the service is successfully created.
			err := wait.Poll(3*time.Second, 200*time.Second, func() (bool, error) {
				serviceGetResponse := &api.ServiceGetResponse{}
				err := GetService(AliceUID, serviceResponse.ServiceID, serviceGetResponse)
				return serviceGetResponse.Service.Repository.Status == api.RepositoryUnknownVcs, err
			})
			Expect(err).To(BeNil())
		})
	})

	Context("with another username", func() {
		service := &api.Service{
			Name:        DefaultServiceName,
			Username:    AliceUser,
			Description: "A service just for test",
			Repository: api.ServiceRepository{
				URL: DefaultTestRepo,
				Vcs: api.Git,
			},
		}
		serviceResponse := &api.ServiceCreationResponse{}
		// FIXME: hang until the auth refactored. Right now, we are unable to check
		// if the service.Name matches user.ID.
		It("should create service successfully, but you should know that this is a security hole.", func() {
			Expect(CreateService(BobUID, service, serviceResponse)).To(BeNil())
		})
	})

	Context("with different user ID", func() {
		service := &api.Service{
			Name:        "cubernetes",
			Username:    AliceUser,
			Description: "A service just for test",
			Repository: api.ServiceRepository{
				URL: DefaultTestRepo,
				Vcs: api.Git,
			},
		}
		serviceResponse := &api.ServiceCreationResponse{}

		It("should create service successfully for correct user.", func() {
			Expect(CreateService(AliceUID, service, serviceResponse)).To(BeNil())
			Expect(serviceResponse.ErrorMessage).To(Equal(""))
			// Wait up to 200 seconds until the service is successfully created.
			err := wait.Poll(3*time.Second, 200*time.Second, func() (bool, error) {
				serviceGetResponse := &api.ServiceGetResponse{}
				err := GetService(AliceUID, serviceResponse.ServiceID, serviceGetResponse)
				return serviceGetResponse.Service.Repository.Status == api.RepositoryHealthy, err
			})
			Expect(err).To(BeNil())
		})

		It("should not be got by HTTP GET method due to different user ID.", func() {
			serviceGetResponse := &api.ServiceGetResponse{}
			err := GetService(BobUID, serviceResponse.ServiceID, serviceGetResponse)
			Expect(err).NotTo(BeNil())
		})

		It("should not be deleted by HTTP DELETE method due to different user ID.", func() {
			serviceDelResponse := &api.ServiceDelResponse{}
			err := DeleteService(BobUID, serviceResponse.ServiceID, serviceDelResponse)
			Expect(err).NotTo(BeNil())
		})

		It("should be deleted by HTTP DELETE method for correct user.", func() {
			serviceDelResponse := &api.ServiceDelResponse{}
			err := DeleteService(AliceUID, serviceResponse.ServiceID, serviceDelResponse)
			Expect(err).To(BeNil())

			serviceGetResponse := &api.ServiceGetResponse{}
			err = GetService(AliceUID, serviceResponse.ServiceID, serviceGetResponse)
			Expect(err).NotTo(BeNil())
		})
	})
})
