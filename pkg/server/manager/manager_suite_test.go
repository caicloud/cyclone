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

package manager_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/caicloud/cyclone/pkg/api"
	_ "github.com/caicloud/cyclone/pkg/scm/provider"
	"github.com/caicloud/cyclone/pkg/server/manager"
	"github.com/caicloud/cyclone/store"
	"gopkg.in/mgo.v2"

	"testing"
	"time"
)

const (
	mongoHost = "localhost:27017"
)

func TestManager(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Project Suite")
	//RunSpecs(t, "Pipeline Suite")
	//RunSpecs(t, "PipelineRecord Suite")
}

var (
	dataStore             store.DataStore
	closing               chan struct{}
	mclosed               chan struct{}
	pipelineRecordManager manager.PipelineRecordManager
	pipelineManager       manager.PipelineManager
	projectManager        manager.ProjectManager
)

var _ = Describe("Project Suite", func() {
	BeforeSuite(func() {
		InitEnv()
		dataStore = NewTestDataStore()
		projectManager, pipelineManager, pipelineRecordManager = NewTestManagers(dataStore)
	})

	AfterSuite(func() {
		dataStore.Close()
		closing <- struct{}{}
		<-mclosed
	})

	Context("ProjectManager works correctly when request valid", func() {
		const (
			projectName = "testProject"
			projectID   = "testID"
		)

		type createProjectIn struct {
			project *api.Project
		}
		type createProjectOut struct {
			project *api.Project
			err     bool
		}
		type createProjectCase struct {
			input  createProjectIn
			output createProjectOut
		}

		BeforeEach(func() {
		})

		AfterEach(func() {
			_ = projectManager.DeleteProject(projectName)
		})

		It("Given a projectManager, when creating new project, it should creat project correctly.", func() {

			createProjectCases := make([]createProjectCase, 0)

			for _, i := range createProjectCases {
				response, err := projectManager.CreateProject(i.input.project)
				if i.output.err {
					Expect(err).To(HaveOccurred())
					continue
				}
				Expect(err).NotTo(HaveOccurred())
				Expect(response).To(Equal(i.output))
			}
		})
	})

	Context("ProjectManager retruen error when request invalid", func() {
	})
})

var _ = Describe("PipelineManager Suite", func() {
	Context("PipelineManager works correctly when request valid", func() {

	})

	Context("PipelineManager retruen error when request invalid", func() {

	})
})

var _ = Describe("PipelineRecordManager Suite", func() {
	Context("PipelineRecordManager works correctly when request valid", func() {

	})

	Context("PipelineRecordManager works correctly when request valid", func() {

	})
})

func InitEnv() {

}

func NewTestDataStore() store.DataStore {
	var err error
	var session *mgo.Session

	closing = make(chan struct{})
	mongoGracePeriod := 30 * time.Second
	// init database
	session, mclosed, err = store.Init(mongoHost, mongoGracePeriod, closing)
	Expect(err).NotTo(HaveOccurred())

	session.SetMode(mgo.Eventual, true)

	return store.NewStore()
}

func NewTestManagers(dataStore store.DataStore) (manager.ProjectManager, manager.PipelineManager, manager.PipelineRecordManager) {
	var err error

	pipelineRecordManager, err = manager.NewPipelineRecordManager(dataStore)
	Expect(err).NotTo(HaveOccurred())

	pipelineManager, err = manager.NewPipelineManager(dataStore, pipelineRecordManager)
	Expect(err).NotTo(HaveOccurred())

	projectManager, err = manager.NewProjectManager(dataStore, pipelineManager)
	Expect(err).NotTo(HaveOccurred())

	return projectManager, pipelineManager, pipelineRecordManager
}
