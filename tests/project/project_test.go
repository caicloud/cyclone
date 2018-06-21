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

var _ = Describe("Project", func() {

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

		BeforeEach(func() {
			response = &api.Project{}
			errResp = &api.ErrorResponse{}
			listResponse = &ListResponse{}
		})

		It("should create project successfully.", func() {
			code, err := CreateProject(project, response, errResp)
			Expect(err).To(BeNil())
			Expect(code).To(Equal(201))
			Expect(errResp.Message).To(Equal(""))
			log.Infof("create project response code :%v", code)
			log.Infof("create project error response :%v", errResp)
		})

		It("should create project failed(conflict).", func() {
			code, err := CreateProject(project, response, errResp)
			Expect(err).To(BeNil())
			Expect(code).To(Equal(409))
			Expect(errResp.Message).To(ContainSubstring("conflict"))
			log.Infof("create project response code :%v", code)
			log.Infof("create project error response :%v", errResp)
		})

		It("should get project successfully.", func() {
			code, err := GetProject(project.Name, response, errResp)
			Expect(err).To(BeNil())
			Expect(code).To(Equal(200))
			Expect(errResp.Message).To(Equal(""))
			Expect(response.Name).To(Equal(project.Name))
		})

		It("should set project successfully.", func() {
			p1 := *project
			p1.Description = "update describe for project"
			code, err := SetProject(&p1, response, errResp)
			Expect(err).To(BeNil())
			Expect(code).To(Equal(200))
			Expect(errResp.Message).To(Equal(""))
			Expect(response.Description).To(Equal(p1.Description))
		})

		It("should list project successfully.", func() {
			code, err := ListProjects(listResponse, errResp)
			Expect(err).To(BeNil())
			Expect(code).To(Equal(200))
			Expect(errResp.Message).To(Equal(""))
			Expect(listResponse.Metadata.Total).To(Equal(1))
		})

		It("should delete project successfully.", func() {
			code, err := DeleteProject(project.Name, response, errResp)
			Expect(err).To(BeNil())
			Expect(errResp.Message).To(Equal(""))
			Expect(code).To(Equal(204))
		})

	})

})
