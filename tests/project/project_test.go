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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/log"
	"github.com/caicloud/cyclone/pkg/osutil"

	. "github.com/caicloud/cyclone/tests/common"
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
	//				serviceA         *api.Service
	//		serviceAResponse *api.ServiceCreationResponse
	//		serviceB         *api.Service
	//		serviceBResponse *api.ServiceCreationResponse
	//		serviceC         *api.Service
	//		serviceCResponse *api.ServiceCreationResponse
	//		projectIDDefault string
	)

	BeforeSuite(func() {
		// Wait cyclone to start.
		WaitComponents()

		// Get docker host and cert path.
		endpoint := osutil.GetStringEnv("DOCKER_HOST", DefaultDockerHost)
		certPath := osutil.GetStringEnv("DOCKER_CERT_PATH", "")
		log.Infof("endpoint: %v, certPath:%v", endpoint, certPath)

	})

	Context("project CRUD right information", func() {
		project := &api.Project{
			Name:        "project1",
			Description: "first test project",
			Owner:       "allen",
			SCM: &api.SCMConfig{
				Type:     "SVN",
				AuthType: "Password",
				Server:   "svn://svnbucket.com",
				Username: "zhujian",
				Password: "passw0rd",
			},
		}
		response := &api.Project{}
		errResp := &api.ErrorResponse{}
		listResponse := &ListResponse{}

		It("should create project successfully.", func() {
			code, err := CreateProject(project, response, errResp)
			Expect(err).To(BeNil())
			Expect(code).To(Equal(201))
			//			Expect(response.ErrorMessage).To(Equal(""))
			log.Infof("create project response code :%v", code)
			log.Infof("create project response :%v", response)
			log.Infof("create project error response :%v", errResp)
		})

		It("should create project conflict.", func() {
			code, err := CreateProject(project, response, errResp)
			Expect(err).To(BeNil())
			Expect(code).To(Equal(409))
			//			Expect(response.ErrorMessage).To(Equal(""))
		})

		It("should get project successfully.", func() {
			code, err := GetProject(project.Name, response, errResp)
			Expect(err).To(BeNil())
			Expect(code).To(Equal(200))
			Expect(response.Name).To(Equal(project.Name))
		})

		It("should set project successfully.", func() {
			p1 := *project
			p1.Description = "update describe for project"
			code, err := SetProject(&p1, response, errResp)
			Expect(err).To(BeNil())
			Expect(code).To(Equal(200))
			Expect(response.Description).To(Equal(p1.Description))
		})

		It("should list project successfully.", func() {
			code, err := ListProjects(listResponse, errResp)
			Expect(err).To(BeNil())
			Expect(code).To(Equal(200))
			Expect(listResponse.Metadata.Total).To(Equal(1))
		})

		It("should delete project successfully.", func() {
			code, err := DeleteProject(project.Name, response, errResp)
			Expect(err).To(BeNil())
			Expect(code).To(Equal(204))
			//			Expect(response.ErrorMessage).To(Equal(""))
		})

	})

	//	Context("with right information", func() {
	//		project := &api.Project{
	//			Name:        DefaultProjectName,
	//			Description: "a project for test",
	//		}
	//		projectResponse := &api.ProjectCreationResponse{}

	//		It("should create project successfully.", func() {
	//			err := CreateProject(AliceUID, project, projectResponse)
	//			Expect(err).To(BeNil())
	//			Expect(projectResponse.ErrorMessage).To(Equal(""))
	//			projectIDDefault = projectResponse.ProjectID
	//		})

	//		It("should set project successfully.", func() {
	//			projectSet := &api.Project{}
	//			projectSetResponse := &api.ProjectSetResponse{}
	//			depencyA := api.ServiceDependency{
	//				ServiceID: serviceAResponse.ServiceID,
	//			}
	//			depencyB := api.ServiceDependency{
	//				ServiceID: serviceBResponse.ServiceID,
	//			}
	//			// set mongo-client depend mongo-server
	//			depencyB.Depend = append(depencyB.Depend, depencyA)
	//			projectSet.Services = append(project.Services, depencyB)
	//			err := SetProject(AliceUID, projectResponse.ProjectID, projectSet, projectSetResponse)
	//			Expect(err).To(BeNil())
	//			Expect(projectSetResponse.ErrorMessage).To(Equal(""))
	//		})

	//		It("should be got by HTTP GET method.", func() {
	//			projectGetResponse := &api.ProjectGetResponse{}
	//			err := GetProject(AliceUID, projectResponse.ProjectID, projectGetResponse)
	//			Expect(err).To(BeNil())
	//		})

	//		It("should be list by HTTP GET method.", func() {
	//			projectListResponse := &api.ProjectListResponse{}
	//			err := ListProjects(AliceUID, projectListResponse)
	//			Expect(err).To(BeNil())
	//			Expect(projectListResponse.ErrorMessage).To(Equal(""))
	//			Expect(len(projectListResponse.Projects)).To(Equal(1))
	//		})
	//	})

	//	Context("with the same project name.", func() {
	//		project := &api.Project{
	//			Name:        DefaultProjectName,
	//			Description: "a same name project for test",
	//		}
	//		projectResponse := &api.ProjectCreationResponse{}

	//		It("should create project failed.", func() {
	//			Expect(CreateProject(AliceUID, project, projectResponse)).NotTo(BeNil())
	//		})
	//	})

	//	Context("set project with wrong information.", func() {
	//		project := &api.Project{
	//			Name:        ProjectNameForSet,
	//			Description: "a project for test",
	//		}
	//		projectResponse := &api.ProjectCreationResponse{}

	//		It("should create project successfully.", func() {
	//			err := CreateProject(AliceUID, project, projectResponse)
	//			Expect(err).To(BeNil())
	//			Expect(projectResponse.ErrorMessage).To(Equal(""))
	//		})

	//		It("should set project failure when service not exist.", func() {
	//			projectSet := &api.Project{}
	//			projectSetResponse := &api.ProjectSetResponse{}
	//			depencyT := api.ServiceDependency{
	//				ServiceID: NotExistServiceID,
	//			}
	//			depencyB := api.ServiceDependency{
	//				ServiceID: serviceBResponse.ServiceID,
	//			}
	//			depencyB.Depend = append(depencyB.Depend, depencyT)
	//			projectSet.Services = append(project.Services, depencyB)
	//			err := SetProject(AliceUID, projectResponse.ProjectID, projectSet, projectSetResponse)
	//			Expect(err).NotTo(BeNil())
	//		})

	//		It("should set project failure when service belong to another user.", func() {
	//			projectSet := &api.Project{}
	//			projectSetResponse := &api.ProjectSetResponse{}
	//			depencyC := api.ServiceDependency{
	//				ServiceID: serviceCResponse.ServiceID,
	//			}
	//			depencyB := api.ServiceDependency{
	//				ServiceID: serviceBResponse.ServiceID,
	//			}
	//			depencyB.Depend = append(depencyB.Depend, depencyC)
	//			projectSet.Services = append(project.Services, depencyB)
	//			err := SetProject(AliceUID, projectResponse.ProjectID, projectSet, projectSetResponse)
	//			Expect(err).NotTo(BeNil())
	//		})
	//	})

	//	Context("create project version.", func() {
	//		projectVersionResponse := &api.ProjectVersionCreationResponse{}
	//		It("should create project version successful", func() {
	//			projectVersion := &api.ProjectVersion{
	//				ProjectID:   projectIDDefault,
	//				Policy:      "manual",
	//				Name:        DefaultVersionName,
	//				Description: "a project version for test",
	//			}
	//			err := CreateProjectVersion(AliceUID, projectVersion, projectVersionResponse)
	//			if nil == err {
	//				log.Infof("create project version: %s", projectVersionResponse.ProjectVersionID)
	//			}
	//			Expect(err).To(BeNil())
	//		})

	//		It("shouldn't create version using the same name as the previous creating.", func() {
	//			projectRepeatVersion := &api.ProjectVersion{
	//				ProjectID:   projectIDDefault,
	//				Policy:      "manual",
	//				Name:        DefaultVersionName,
	//				Description: "a project version for test",
	//			}
	//			projectRepeatVersionResponse := &api.ProjectVersionCreationResponse{}
	//			err := CreateProjectVersion(AliceUID, projectRepeatVersion, projectRepeatVersionResponse)
	//			Expect(err).NotTo(BeNil())
	//		})

	//		It("should fail to create the version, because cyclone couldn't find the project by projectID.", func() {
	//			projectRepeatVersion := &api.ProjectVersion{
	//				ProjectID:   "dummy",
	//				Policy:      "manual",
	//				Name:        DefaultVersionName + "2",
	//				Description: "a project version for test",
	//			}
	//			projectRepeatVersionResponse := &api.ProjectVersionCreationResponse{}
	//			err := CreateProjectVersion(AliceUID, projectRepeatVersion, projectRepeatVersionResponse)
	//			Expect(err).NotTo(BeNil())
	//		})

	//		It("should fail to create the version for the project not belong to this user.", func() {
	//			projectRepeatVersion := &api.ProjectVersion{
	//				ProjectID:   projectIDDefault,
	//				Policy:      "manual",
	//				Name:        DefaultVersionName + "3",
	//				Description: "a project version for test",
	//			}
	//			projectRepeatVersionResponse := &api.ProjectVersionCreationResponse{}
	//			err := CreateProjectVersion(BobUID, projectRepeatVersion, projectRepeatVersionResponse)
	//			Expect(err).NotTo(BeNil())
	//		})

	//		It("should be able to get project version via HTTP GET method.", func() {
	//			versionGetResponse := &api.ProjectVersionGetResponse{}
	//			// Wait up to 600 seconds until the project version is successfully created.
	//			err := wait.Poll(20*time.Second, 1800*time.Second, func() (bool, error) {
	//				err := GetProjectVersion(AliceUID, projectVersionResponse.ProjectVersionID, versionGetResponse)
	//				return versionGetResponse.ProjectVersion.Status == api.VersionHealthy, err
	//			})
	//			Expect(err).To(BeNil())
	//			Expect(versionGetResponse.ErrorMessage).To(Equal(""))
	//			Expect(versionGetResponse.ProjectVersion).NotTo(BeNil())
	//			Expect(versionGetResponse.ProjectVersion.Status).To(Equal(api.VersionHealthy))
	//		})

	//		It("should be list by HTTP GET method.", func() {
	//			projectVersionListResponse := &api.ProjectVersionListResponse{}
	//			err := ListProjectVersions(AliceUID, projectIDDefault, projectVersionListResponse)
	//			Expect(err).To(BeNil())
	//			Expect(projectVersionListResponse.ErrorMessage).To(Equal(""))
	//			Expect(len(projectVersionListResponse.ProjectVersions)).To(Equal(1))
	//		})
	//	})

	//	Context("delete project with the different user ID.", func() {
	//		It("shouldn't be deleted by HTTP DELETE method when the project belong to another user.", func() {
	//			projectDelResponse := &api.ProjectDelResponse{}
	//			err := DeleteProject(BobUID, projectIDDefault, projectDelResponse)
	//			Expect(err).NotTo(BeNil())
	//		})

	//		It("should be deleted by HTTP DELETE method.", func() {
	//			projectDelResponse := &api.ProjectDelResponse{}
	//			err := DeleteProject(AliceUID, projectIDDefault, projectDelResponse)
	//			Expect(err).To(BeNil())
	//			Expect(projectDelResponse.ErrorMessage).To(Equal(""))
	//		})
	//	})
})
