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

package project

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
	ServiceNameA       = "mongo-client"
	TestRepoA          = "https://github.com/caicloud-circle/mongo-client.git"
	ServiceNameB       = "mongo-server"
	TestRepoB          = "https://github.com/caicloud-circle/mongo-server.git"
	ServiceNameC       = "mongo-client"
	TestRepoC          = "https://github.com/caicloud-circle/mongo-client.git"
	DefaultProjectName = "project-mongo"
	DefaultVersionName = "v0.1.0"
	NotExistServiceID  = "tempid"
	ProjectNameForSet  = "project-mongo-set"
)

var _ = Describe("Service", func() {
	var (
		serviceA         *api.Service
		serviceAResponse *api.ServiceCreationResponse
		serviceB         *api.Service
		serviceBResponse *api.ServiceCreationResponse
		serviceC         *api.Service
		serviceCResponse *api.ServiceCreationResponse
		projectIDDefault string
	)

	BeforeSuite(func() {
		// Wait cyclone to start.
		WaitComponents()

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
		var err error
		_, err = docker.NewManager(
			endpoint, certPath, registry)
		if err != nil {
			log.Fatalf("Unable to connect to docker daemon: %v", err)
		}

		// Create a test service A.
		serviceA = &api.Service{
			Name:        ServiceNameA,
			Username:    AliceUser,
			Description: "A mongo service just for test",
			Repository: api.ServiceRepository{
				URL: TestRepoA,
				Vcs: api.Git,
			},
		}
		serviceAResponse = &api.ServiceCreationResponse{}
		err = CreateService(AliceUID, serviceA, serviceAResponse)
		if err != nil {
			log.Errorf("unexpected error while creating service for user %v: %v", AliceUID, err)
		}

		// Wait up to 30 seconds until the service is successfully created.
		err = wait.Poll(2*time.Second, 30*time.Second, func() (bool, error) {
			serviceGetResponse := &api.ServiceGetResponse{}
			err := GetService(AliceUID, serviceAResponse.ServiceID, serviceGetResponse)
			return serviceGetResponse.Service.Repository.Status == api.RepositoryHealthy, err
		})
		if err != nil {
			log.ErrorWithFields("unexpected repository status", log.Fields{"repository": serviceA.Repository,
				"status": serviceA.Repository.Status})
		}

		// Create a test service B.
		serviceB = &api.Service{
			Name:        ServiceNameB,
			Username:    AliceUser,
			Description: "A mongo client just for test",
			Repository: api.ServiceRepository{
				URL: TestRepoB,
				Vcs: api.Git,
			},
		}
		serviceBResponse = &api.ServiceCreationResponse{}
		err = CreateService(AliceUID, serviceB, serviceBResponse)
		if err != nil {
			log.Errorf("unexpected error while creating service for user %v: %v", AliceUID, err)
		}

		// Wait up to 30 seconds until the service is successfully created.
		err = wait.Poll(2*time.Second, 30*time.Second, func() (bool, error) {
			serviceGetResponse := &api.ServiceGetResponse{}
			err := GetService(AliceUID, serviceBResponse.ServiceID, serviceGetResponse)
			return serviceGetResponse.Service.Repository.Status == api.RepositoryHealthy, err
		})
		if err != nil {
			log.ErrorWithFields("unexpected repository status", log.Fields{"repository": serviceB.Repository,
				"status": serviceB.Repository.Status})
		}

		// Create a test service C.
		serviceC = &api.Service{
			Name:        ServiceNameC,
			Username:    BobUser,
			Description: "A mongo client just for test",
			Repository: api.ServiceRepository{
				URL: TestRepoC,
				Vcs: api.Git,
			},
		}
		serviceCResponse = &api.ServiceCreationResponse{}
		err = CreateService(BobUID, serviceC, serviceCResponse)
		if err != nil {
			log.Errorf("unexpected error while creating service for user %v: %v", BobUID, err)
		}

		// Wait up to 30 seconds until the service is successfully created.
		err = wait.Poll(2*time.Second, 30*time.Second, func() (bool, error) {
			serviceGetResponse := &api.ServiceGetResponse{}
			err := GetService(BobUID, serviceCResponse.ServiceID, serviceGetResponse)
			return serviceGetResponse.Service.Repository.Status == api.RepositoryHealthy, err
		})
		if err != nil {
			log.ErrorWithFields("unexpected repository status", log.Fields{"repository": serviceC.Repository,
				"status": serviceC.Repository.Status})
		}

	})

	Context("with right information", func() {
		project := &api.Project{
			Name:        DefaultProjectName,
			Description: "a project for test",
		}
		projectResponse := &api.ProjectCreationResponse{}

		It("should create project successfully.", func() {
			err := CreateProject(AliceUID, project, projectResponse)
			Expect(err).To(BeNil())
			Expect(projectResponse.ErrorMessage).To(Equal(""))
			projectIDDefault = projectResponse.ProjectID
		})

		It("should set project successfully.", func() {
			projectSet := &api.Project{}
			projectSetResponse := &api.ProjectSetResponse{}
			depencyA := api.ServiceDependency{
				ServiceID: serviceAResponse.ServiceID,
			}
			depencyB := api.ServiceDependency{
				ServiceID: serviceBResponse.ServiceID,
			}
			// set mongo-client depend mongo-server
			depencyB.Depend = append(depencyB.Depend, depencyA)
			projectSet.Services = append(project.Services, depencyB)
			err := SetProject(AliceUID, projectResponse.ProjectID, projectSet, projectSetResponse)
			Expect(err).To(BeNil())
			Expect(projectSetResponse.ErrorMessage).To(Equal(""))
		})

		It("should be got by HTTP GET method.", func() {
			projectGetResponse := &api.ProjectGetResponse{}
			err := GetProject(AliceUID, projectResponse.ProjectID, projectGetResponse)
			Expect(err).To(BeNil())
		})

		It("should be list by HTTP GET method.", func() {
			projectListResponse := &api.ProjectListResponse{}
			err := ListProjects(AliceUID, projectListResponse)
			Expect(err).To(BeNil())
			Expect(projectListResponse.ErrorMessage).To(Equal(""))
			Expect(len(projectListResponse.Projects)).To(Equal(1))
		})
	})

	Context("with the same project name.", func() {
		project := &api.Project{
			Name:        DefaultProjectName,
			Description: "a same name project for test",
		}
		projectResponse := &api.ProjectCreationResponse{}

		It("should create project failed.", func() {
			Expect(CreateProject(AliceUID, project, projectResponse)).NotTo(BeNil())
		})
	})

	Context("set project with wrong information.", func() {
		project := &api.Project{
			Name:        ProjectNameForSet,
			Description: "a project for test",
		}
		projectResponse := &api.ProjectCreationResponse{}

		It("should create project successfully.", func() {
			err := CreateProject(AliceUID, project, projectResponse)
			Expect(err).To(BeNil())
			Expect(projectResponse.ErrorMessage).To(Equal(""))
		})

		It("should set project failure when service not exist.", func() {
			projectSet := &api.Project{}
			projectSetResponse := &api.ProjectSetResponse{}
			depencyT := api.ServiceDependency{
				ServiceID: NotExistServiceID,
			}
			depencyB := api.ServiceDependency{
				ServiceID: serviceBResponse.ServiceID,
			}
			depencyB.Depend = append(depencyB.Depend, depencyT)
			projectSet.Services = append(project.Services, depencyB)
			err := SetProject(AliceUID, projectResponse.ProjectID, projectSet, projectSetResponse)
			Expect(err).NotTo(BeNil())
		})

		It("should set project failure when service belong to another user.", func() {
			projectSet := &api.Project{}
			projectSetResponse := &api.ProjectSetResponse{}
			depencyC := api.ServiceDependency{
				ServiceID: serviceCResponse.ServiceID,
			}
			depencyB := api.ServiceDependency{
				ServiceID: serviceBResponse.ServiceID,
			}
			depencyB.Depend = append(depencyB.Depend, depencyC)
			projectSet.Services = append(project.Services, depencyB)
			err := SetProject(AliceUID, projectResponse.ProjectID, projectSet, projectSetResponse)
			Expect(err).NotTo(BeNil())
		})
	})

	Context("create project version.", func() {
		projectVersionResponse := &api.ProjectVersionCreationResponse{}
		It("should create project version successful", func() {
			projectVersion := &api.ProjectVersion{
				ProjectID:   projectIDDefault,
				Policy:      "manual",
				Name:        DefaultVersionName,
				Description: "a project version for test",
			}
			err := CreateProjectVersion(AliceUID, projectVersion, projectVersionResponse)
			if nil == err {
				log.Infof("create project version: %s", projectVersionResponse.ProjectVersionID)
			}
			Expect(err).To(BeNil())
		})

		It("shouldn't create version using the same name as the previous creating.", func() {
			projectRepeatVersion := &api.ProjectVersion{
				ProjectID:   projectIDDefault,
				Policy:      "manual",
				Name:        DefaultVersionName,
				Description: "a project version for test",
			}
			projectRepeatVersionResponse := &api.ProjectVersionCreationResponse{}
			err := CreateProjectVersion(AliceUID, projectRepeatVersion, projectRepeatVersionResponse)
			Expect(err).NotTo(BeNil())
		})

		It("should fail to create the version, because cyclone couldn't find the project by projectID.", func() {
			projectRepeatVersion := &api.ProjectVersion{
				ProjectID:   "dummy",
				Policy:      "manual",
				Name:        DefaultVersionName + "2",
				Description: "a project version for test",
			}
			projectRepeatVersionResponse := &api.ProjectVersionCreationResponse{}
			err := CreateProjectVersion(AliceUID, projectRepeatVersion, projectRepeatVersionResponse)
			Expect(err).NotTo(BeNil())
		})

		It("should fail to create the version for the project not belong to this user.", func() {
			projectRepeatVersion := &api.ProjectVersion{
				ProjectID:   projectIDDefault,
				Policy:      "manual",
				Name:        DefaultVersionName + "3",
				Description: "a project version for test",
			}
			projectRepeatVersionResponse := &api.ProjectVersionCreationResponse{}
			err := CreateProjectVersion(BobUID, projectRepeatVersion, projectRepeatVersionResponse)
			Expect(err).NotTo(BeNil())
		})

		It("should be able to get project version via HTTP GET method.", func() {
			versionGetResponse := &api.ProjectVersionGetResponse{}
			// Wait up to 600 seconds until the project version is successfully created.
			err := wait.Poll(20*time.Second, 1800*time.Second, func() (bool, error) {
				err := GetProjectVersion(AliceUID, projectVersionResponse.ProjectVersionID, versionGetResponse)
				return versionGetResponse.ProjectVersion.Status == api.VersionHealthy, err
			})
			Expect(err).To(BeNil())
			Expect(versionGetResponse.ErrorMessage).To(Equal(""))
			Expect(versionGetResponse.ProjectVersion).NotTo(BeNil())
			Expect(versionGetResponse.ProjectVersion.Status).To(Equal(api.VersionHealthy))
		})

		It("should be list by HTTP GET method.", func() {
			projectVersionListResponse := &api.ProjectVersionListResponse{}
			err := ListProjectVersions(AliceUID, projectIDDefault, projectVersionListResponse)
			Expect(err).To(BeNil())
			Expect(projectVersionListResponse.ErrorMessage).To(Equal(""))
			Expect(len(projectVersionListResponse.ProjectVersions)).To(Equal(1))
		})
	})

	Context("delete project with the different user ID.", func() {
		It("shouldn't be deleted by HTTP DELETE method when the project belong to another user.", func() {
			projectDelResponse := &api.ProjectDelResponse{}
			err := DeleteProject(BobUID, projectIDDefault, projectDelResponse)
			Expect(err).NotTo(BeNil())
		})

		It("should be deleted by HTTP DELETE method.", func() {
			projectDelResponse := &api.ProjectDelResponse{}
			err := DeleteProject(AliceUID, projectIDDefault, projectDelResponse)
			Expect(err).To(BeNil())
			Expect(projectDelResponse.ErrorMessage).To(Equal(""))
		})
	})
})
