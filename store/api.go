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
package store

import (
	oldAPI "github.com/caicloud/cyclone/api"

	"github.com/caicloud/cyclone/cloud"
	"github.com/caicloud/cyclone/pkg/api"

	"time"
)

type DataStore interface {
	ScmToken
	Service
	Version
	Project
	PipeLine
	PipelineRecord
	Resource
	Remote
	Cloud
	Log
	Deploy
	Close()
	Ping() error
}

type PipeLine interface {
	CreatePipeline(pipeline *api.Pipeline) (*api.Pipeline, error)
	FindPipelineByName(projectID string, name string) (*api.Pipeline, error)
	FindPipelineByID(pipelineID string) (*api.Pipeline, error)
	FindPipelineByServiceID(serviceID string) (*api.Pipeline, error)
	FindPipelinesByProjectID(projectID string, queryParams api.QueryParams) ([]api.Pipeline, int, error)
	UpdatePipeline(pipeline *api.Pipeline) error
	DeletePipelineByID(pipelineID string) error
	DeletePipelinesByProjectID(projectID string) error
}

type Project interface {
	CreateProject(project *api.Project) (*api.Project, error)
	FindProjectByName(name string) (*api.Project, error)
	FindProjectByID(projectID string) (*api.Project, error)
	FindProjectByServiceID(serviceID string) (*api.Project, error)
	UpdateProject(project *api.Project) error
	DeleteProjectByID(projectID string) error
	GetProjects(queryParams api.QueryParams) ([]api.Project, int, error)
}

type PipelineRecord interface {
	CreatePipelineRecord(pipelineRecord *api.PipelineRecord) (*api.PipelineRecord, error)
	FindPipelineRecordsByPipelineID(pipelineID string, queryParams api.QueryParams) ([]api.PipelineRecord, int, error)
	FindPipelineRecordByID(pipelineRecordID string) (*api.PipelineRecord, error)
	UpdatePipelineRecord(pipelineRecord *api.PipelineRecord) error
	DeletePipelineRecordByID(pipelineRecordID string) error
	DeletePipelineRecordsByPipelineID(pipelineID string) error
}

type Resource interface {
	NewResourceDocument(resource *oldAPI.Resource) error
	UpdateResourceDocument(resource *oldAPI.Resource) error
	FindResourceByID(userID string) (*oldAPI.Resource, error)
	UpdateResourceStatus(userID string, memory float64, cpu float64) error
}

type Remote interface {
	NewTokenDocument(token *oldAPI.VscToken) error
	FindtokenByUserID(userID, urlvsc string) (*oldAPI.VscToken, error)
	UpdateToken(token *oldAPI.VscToken) error
	RemoveTokeninDB(userID string, urlvsc string) error
}

type Log interface {
	NewVersionLogDocument(versionLog *oldAPI.VersionLog) (string, error)
	FindVersionLogByID(LogID string) (*oldAPI.VersionLog, error)
	FindVersionLogByVersionID(versionID string) (*oldAPI.VersionLog, error)
	UpdateVersionLogDocument(versionLog *oldAPI.VersionLog) error
}

type Deploy interface {
	NewDeployDocument(deploy *oldAPI.Deploy) (string, error)
	FindDeployByID(deployID string) (*oldAPI.Deploy, error)
	UpsertDeployDocument(deploy *oldAPI.Deploy) (string, error)
	FindDeployByUserID(userID string) ([]oldAPI.Deploy, error)
	DeleteDeployByID(deployID string) error
}

type Cloud interface {
	InsertCloud(doc *cloud.Options) error
	UpsertCloud(doc *cloud.Options) error
	FindAllClouds() ([]cloud.Options, error)
	FindCloudByName(name string) (*cloud.Options, error)
	DeleteCloudByName(name string) error
}

type Service interface {
	FindServiceByCondition(userID, servicename string) ([]oldAPI.Service, error)
	NewServiceDocument(service *oldAPI.Service) (string, error)
	UpdateRepositoryStatus(serviceID string, status oldAPI.RepositoryStatus) error
	FindServicesByUserID(userID string) ([]oldAPI.Service, error)
	FindServiceByID(serviceID string) (*oldAPI.Service, error)
	DeleteServiceByID(serviceID string) error
	AddNewVersion(serviceID string, versionID string) error
	AddNewFailVersion(serviceID string, versionID string) error
	UpdateServiceLastInfo(serviceID string, lasttime time.Time, lastname string) error
	UpsertServiceDocument(service *oldAPI.Service) (string, error)
}

type Version interface {
	FindVersionsByCondition(serviceID, versionname string) ([]oldAPI.Version, error)
	NewVersionDocument(version *oldAPI.Version) (string, error)
	UpdateVersionDocument(versionID string, version oldAPI.Version) error
	FindVersionByID(versionID string) (*oldAPI.Version, error)
	FindVersionsByServiceID(serviceID string) ([]oldAPI.Version, error)
	FindVersionsWithPaginationByServiceID(serviceID string, filter map[string]interface{}, start, limit int) ([]oldAPI.Version, int, error)
	FindRecentVersionsByServiceID(serviceID string, filter map[string]interface{}, limit int) ([]oldAPI.Version, int, error)
	FindLatestVersionByServiceID(serviceID string) (*oldAPI.Version, error)
	DeleteVersionByID(versionID string) error
}

type ScmToken interface {
	CreateToken(token *api.ScmToken) (*api.ScmToken, error)
	Findtoken(projectID, scmType string) (*api.ScmToken, error)
	UpdateToken2(token *api.ScmToken) error
	DeleteToken(projectID, scmType string) error
}
