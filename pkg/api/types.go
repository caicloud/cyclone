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

package api

import (
	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/cloud"
	"golang.org/x/oauth2"
	"time"
)

// Project represents a group to manage a set of related applications. It maybe a real project, which contains several or many applications.
type Project struct {
	ID             string     `bson:"_id,omitempty" json:"id,omitempty" description:"id of the project"`
	Name           string     `bson:"name,omitempty" json:"name,omitempty" description:"name of the project, should be unique"`
	Description    string     `bson:"description,omitempty" json:"description,omitempty" description:"description of the project"`
	Owner          string     `bson:"owner,omitempty" json:"owner,omitempty" description:"owner of the project"`
	SCM            *SCMConfig `bson:"scm,omitempty" json:"scm,omitempty" description:"scm config of the project"`
	Worker         *Worker    `bson:"worker,omitempty" json:"worker,omitempty" description:"worker config of the project"`
	CreationTime   time.Time  `bson:"creationTime,omitempty" json:"creationTime,omitempty" description:"creation time of the project"`
	LastUpdateTime time.Time  `bson:"lastUpdateTime,omitempty" json:"lastUpdateTime,omitempty" description:"last update time of the project"`
}

// Worker represents the config of worker for the pipelines of the project.
type Worker struct {
	Namespace string `bson:"namespace,omitempty" json:"namespace,omitempty" description:"k8s namespace to create the worker"`
}

// Pipeline represents a set of configs to describe the workflow of CI/CD.
type Pipeline struct {
	ID          string `bson:"_id,omitempty" json:"id,omitempty" description:"id of the pipeline"`
	Name        string `bson:"name,omitempty" json:"name,omitempty" description:"name of the pipeline，unique in one project"`
	Alias       string `bson:"alias,omitempty" json:"alias,omitempty" description:"alias of the pipeline"`
	Description string `bson:"description,omitempty" json:"description,omitempty" description:"description of the pipeline"`
	Owner       string `bson:"owner,omitempty" json:"owner,omitempty" description:"owner of the pipeline"`
	ProjectID   string `bson:"projectID,omitempty" json:"projectID,omitempty" description:"id of the project which the pipeline belongs to"`
	// TODO （robin）Remove the association between the pipeline and the service after pipeline replaces service.
	ServiceID            string            `bson:"serviceID,omitempty" json:"serviceID,omitempty" description:"id of the service which the pipeline is related to"`
	Build                *Build            `bson:"build,omitempty" json:"build,omitempty" description:"build spec of the pipeline"`
	AutoTrigger          *AutoTrigger      `bson:"autoTrigger,omitempty" json:"autoTrigger,omitempty" description:"auto trigger strategy of the pipeline"`
	CreationTime         *time.Time        `bson:"creationTime,omitempty" json:"creationTime,omitempty" description:"creation time of the pipeline"`
	LastUpdateTime       *time.Time        `bson:"lastUpdateTime,omitempty" json:"lastUpdateTime,omitempty" description:"last update time of the pipeline"`
	RecentRecords        []*PipelineRecord `bson:"recentRecords,omitempty" json:"recentRecords,omitempty" description:"recent records of the pipeline"`
	RecentSuccessRecords []*PipelineRecord `bson:"recentSuccessRecords,omitempty" json:"recentSuccessRecords,omitempty" description:"recent success records of the pipeline"`
	RecentFailedRecords  []*PipelineRecord `bson:"recentFailedRecords,omitempty" json:"recentFailedRecords,omitempty" description:"recent failed records of the pipeline"`

	// The deploy id
	DeployID string `bson:"deploy_id,omitempty" json:"deploy_id,omitempty"`
	// Build script path. When this script path (or maybe change to just text string)
	// is specified, we build image using BuildPath.
	// TODO: It's better to change to PreBuildHook and PostBuildHook to let Cyclone
	// control building image; otherwise, user can use arbitrary image name.
	BuildPath string `bson:"build_path,omitempty" json:"build_path,omitempty"`
	// Repository information of the service.
	// TODO: For private repository, we need OAuth.
	Repository *api.ServiceRepository `bson:"repository,omitempty" json:"repository,omitempty"`
	Jconfig    *api.JenkinsConfig     `bson:"jconfig,omitempty" json:"jconfig,omitempty"`
	// Email porfile.
	Profile *api.NotifyProfile `bson:"profile,omitempty" json:"profile,omitempty"`
	// Record last build version time or service create time first
	LastCreateTIme time.Time `bson:"last_createtime,omitempty" json:"last_createtime,omitempty"`
	// Record last build version name
	LastVersionName string `bson:"last_versionname,omitempty" json:"last_versionname,omitempty"`
	// Deploy plans
	DeployPlans []api.DeployPlan `bson:"deploy_plans,omitempty" json:"deploy_plans,omitempty"`
	// Repository information of the service.
	YAMLConfigName string `bson:"yaml_config_name,omitempty" json:"yaml_config_name,omitempty"`

	// for code to deployment
	Hooks      []api.Hook `bson:"hooks,omitempty" json:"hooks,omitempty"`
	PublishNow bool       `bson:"publishNow,omitempty" json:"publishNow,omitempty"`

	Dockerfile string `bson:"dockerfile,omitempty" json:"dockerfile,omitempty"`

	ImageName string `bson:"image_name,omitempty" json:"image_name,omitempty"`
}

// Build represents the build config and stages of CI.
type Build struct {
	BuilderImage *BuilderImage `bson:"builderImage,omitempty" json:"builderImage,omitempty" description:"image information of the builder"`
	Stages       *BuildStages  `bson:"stages,omitempty" json:"stages,omitempty" description:"stages of CI"`
}

// EnvVar represents the environment variables with name and value.
type EnvVar struct {
	Name  string `bson:"name,omitempty" json:"name,omitempty" description:"name of the environment variable"`
	Value string `bson:"value,omitempty" json:"value,omitempty" description:"value of the environment variable"`
}

// BuilderImage represents the image information of the builder.
type BuilderImage struct {
	Image   string    `bson:"image,omitempty" json:"image,omitempty" description:"image name of the builder"`
	EnvVars []*EnvVar `bson:"envVars,omitempty" json:"envVars,omitempty" description:"environment variables of the builder"`
}

// BuildStages represents the build stages of CI.
type BuildStages struct {
	CodeCheckout    *CodeCheckoutStage    `bson:"codeCheckout,omitempty" json:"codeCheckout,omitempty" description:"code checkout stage"`
	UnitTest        *UnitTestStage        `bson:"unitTest,omitempty" json:"unitTest,omitempty" description:"unit test stage"`
	CodeScan        *CodeScanStage        `bson:"codeScan,omitempty" json:"codeScan,omitempty" description:"code scan stage"`
	Package         *PackageStage         `bson:"package,omitempty" json:"package,omitempty" description:"package stage"`
	ImageBuild      *ImageBuildStage      `bson:"imageBuild,omitempty" json:"imageBuild,omitempty" description:"image build stage"`
	IntegrationTest *IntegrationTestStage `bson:"integrationTest,omitempty" json:"integrationTest,omitempty" description:"integration test stage"`
	ImageRelease    *ImageReleaseStage    `bson:"imageRelease,omitempty" json:"imageRelease,omitempty" description:"image release stage"`

	PreBuild   *Prebuild
	ImageBuild *ImageBuild
	Integrate  *Integrate
	Publish    *Publish
	Deploy     *Deploy
}

type Prebuild struct {
	GeneralStage
	Outputs     []string `bson:"outputs,omitempty" json:"outputs,omitempty" description:"list of output path of this stage"`
	Image       string   `bson:"image,omitempty" json:"image,omitempty" description:"image name of the builder"`
	Environment []string `bson:"env,omitempty" json:"env,omitempty" description:"environment variables of the builder"`
}

type ImageBuild struct {
	ImageBuild     *ImageBuildStage `bson:"imageBuild,omitempty" json:"imageBuild,omitempty" description:"image build stage"`
	DockerfileName string
	ContextDir     string
}

type Integrate struct {
	IntegrationTest *IntegrationTestStage `bson:"integrationTest,omitempty" json:"integrationTest,omitempty" description:"integration test stage"`
	Image           string
	Services        *Service
	GeneralStage
}

type Publish struct {
}

type Deploy struct {
}

// GeneralStage represents the basic config shared by all stages.
type GeneralStage struct {
	Command []string `bson:"command,omitempty" json:"command,omitempty" description:"list of commands to run for this stage"`
}

// CodeCheckoutStage represents the config of code checkout stage.
type CodeCheckoutStage struct {
	CodeSources []*CodeSource `bson:"codeSources,omitempty" json:"codeSources,omitempty" description:"list of code sources to be checked out"`
}

// SCMType represents the type of SCM, supports gitlab, github and svn.
type SCMType string

const (
	GitLab SCMType = "Gitlab"
	GitHub         = "Github"
	SVN            = "SVN"
)

// SCMConfig represents the config of SCM.
type SCMConfig struct {
	Type     SCMType `bson:"type,omitempty" json:"type,omitempty" description:"SCM type, support gitlab, github and svn"`
	Server   string  `bson:"server,omitempty" json:"server,omitempty" description:"server of the SCM"`
	Username string  `bson:"username,omitempty" json:"username,omitempty" description:"username of the SCM"`
	Password string  `bson:"password,omitempty" json:"password,omitempty" description:"password of the SCM"`
	Token    string  `bson:"token,omitempty" json:"token,omitempty" description:"token of the SCM"`
}

// CodeSource represents the config of code source, only one type is supported.
type CodeSource struct {
	Type SCMType `bson:"type,omitempty" json:"type,omitempty" description:"type of code source, support gitlab, github and svn"`
	// Whether is the main repo. Only support webhook and tag for main repo.
	Main   bool       `bson:"main,omitempty" json:"main,omitempty" description:"whether is the main repo"`
	GitLab *GitSource `bson:"gitLab,omitempty" json:"gitLab,omitempty" description:"code from gitlab"`
	GitHub *GitSource `bson:"gitHub,omitempty" json:"gitHub,omitempty" description:"code from github"`
}

// GitSource represents the config to get code from git.
type GitSource struct {
	Url      string `bson:"url,omitempty" json:"url,omitempty" description:"url of git repo"`
	Ref      string `bson:"ref,omitempty" json:"ref,omitempty" description:"reference of git repo, support branch, tag"`
	Username string `bson:"username,omitempty" json:"username,omitempty" description:"username of git"`
	Password string `bson:"password,omitempty" json:"password,omitempty" description:"password of git"`
}

// UnitTestStage represents the config of unit test stage.
type UnitTestStage struct {
	GeneralStage
	Outputs []string `bson:"outputs,omitempty" json:"outputs,omitempty" description:"list of output path of this stage"`
}

// CodeScanStage represents the config of code scan stage.
type CodeScanStage struct {
	GeneralStage
	Outputs []string `bson:"outputs,omitempty" json:"outputs,omitempty" description:"list of output path of this stage"`
}

// PackageStage represents the config of package stage.
type PackageStage struct {
	GeneralStage
	Outputs []string `bson:"outputs,omitempty" json:"outputs,omitempty" description:"list of output path of this stage"`
}

// ImageBuildStage represents the config of image build stage.
type ImageBuildStage struct {
	BuildInfos []*ImageBuildInfo `bson:"buildInfos,omitempty" json:"buildInfos,omitempty" description:"list of output path of this stage"`
}

// ImageBuildInfo represents the config to build the image. Only one of Dockerfile and DockerfilePath needs to be set.
// If both of them are set, Dockerfile will be used with high priority.
type ImageBuildInfo struct {
	ContextDir     string `bson:"contextDir,omitempty" json:"contextDir,omitempty" description:"context directory for image build"`
	Dockerfile     string `bson:"dockerfile,omitempty" json:"dockerfile,omitempty" description:"dockerfile content for image build"`
	DockerfilePath string `bson:"dockerfilePath,omitempty" json:"dockerfilePath,omitempty" description:"dockerfile path for image build"`
	ImageName      string `bson:"imageName,omitempty" json:"imageName,omitempty" description:"name of the built image"`
}

// IntegrationTestStage represents the config of integration test stage.
type IntegrationTestStage struct {
	Config   *IntegrationTestConfig `bson:"Config,omitempty" json:"Config,omitempty" description:"integration test config"`
	Services []*Service             `bson:"services,omitempty" json:"services,omitempty" description:"list of dependent services for integration test"`
}

// IntegrationTestConfig represents the config for integration test.
type IntegrationTestConfig struct {
	ImageName string   `bson:"imageName,omitempty" json:"imageName,omitempty" description:"built image name to run the integration test"`
	Command   []string `bson:"command,omitempty" json:"command,omitempty" description:"list of commands to run for integration test"`
	EnvVars   []EnvVar `bson:"envVars,omitempty" json:"envVars,omitempty" description:"environment variables for integration test"`
}

// Service represents the dependent service needed for integration test.
type Service struct {
	Name    string    `bson:"name,omitempty" json:"name,omitempty" description:"name of the service"`
	Image   string    `bson:"image,omitempty" json:"image,omitempty" description:"image name of the service"`
	Command []string  `bson:"command,omitempty" json:"command,omitempty" description:"list of commands to start the service"`
	EnvVars []*EnvVar `bson:"envVars,omitempty" json:"envVars,omitempty" description:"environment variables of the service"`
}

// ImageReleaseStage represents the config of image release stage.
type ImageReleaseStage struct {
	ReleasePolicy []ImageReleasePolicy `bson:"releasePolicy,omitempty" json:"releasePolicy,omitempty" description:"list of policies for image release"`
}

// ImageReleasePolicyType represents the type of image release policy.
type ImageReleasePolicyType string

const (
	// AlwaysRelease always releases the images.
	AlwaysRelease ImageReleasePolicyType = "Always"

	// IntegrationTestSuccessRelease releases the images only when the integration test success.
	IntegrationTestSuccessRelease ImageReleasePolicyType = "IntegrationTestSuccess"
)

// ImageReleasePolicy represents the policy to release image.
type ImageReleasePolicy struct {
	ImageName string                 `bson:"imageName,omitempty" json:"imageName,omitempty" description:"image to be released"`
	Type      ImageReleasePolicyType `bson:"type,omitempty" json:"type,omitempty" description:"type of image release policy"`
}

// AutoTrigger represents the auto trigger strategy of the pipeline.
type AutoTrigger struct {
	SCMTrigger   *SCMTrigger   `bson:"scmTrigger,omitempty" json:"scmTrigger,omitempty" description:"SCM trigger strategy"`
	TimerTrigger *TimerTrigger `bson:"timerTrigger,omitempty" json:"timerTrigger,omitempty" description:"timer trigger strategy"`
}

// SCMTrigger represents the auto trigger strategy from SCM.
type SCMTrigger struct {
	CommitTrigger *CommitTrigger `bson:"commitTrigger,omitempty" json:"commitTrigger,omitempty" description:"commit trigger strategy"`
}

// GeneralTrigger represents the general config for all auto trigger strategies.
type GeneralTrigger struct {
	Stages string `bson:"stages,omitempty" json:"stages,omitempty" description:"stages of the auto triggered running"`
}

// CommitTrigger represents the trigger from SCM commit.
type CommitTrigger struct {
	GeneralTrigger
}

// CommitWithCommentsTrigger represents the trigger from SCM commit with specified comments.
type CommitWithCommentsTrigger struct {
	GeneralTrigger
	Comment []string `bson:"comment,omitempty" json:"comment,omitempty" description:"list of comments in commit"`
}

// TimerTrigger represents the auto trigger strategy from timer.
type TimerTrigger struct {
	CronTrigger []*CronTrigger `bson:"cronTrigger,omitempty" json:"cronTrigger,omitempty" description:"list of cron trigger strategies"`
}

// CronTrigger represents the trigger from cron job.
type CronTrigger struct {
	GeneralTrigger
	Expression string `bson:"expression,omitempty" json:"expression,omitempty" description:"expression of cron job"`
}

// PipelineRecord represents the running record of pipeline.
type PipelineRecord struct {
	ID            string                 `bson:"_id,omitempty" json:"id,omitempty" description:"id of the pipeline record"`
	Name          string                 `bson:"name,omitempty" json:"name,omitempty" description:"name of the pipeline record"`
	PipelineID    string                 `bson:"pipelineID,omitempty" json:"pipelineID,omitempty" description:"id of the related pipeline which the pipeline record belongs to"`
	VersionID     string                 `bson:"versionID,omitempty" json:"versionID,omitempty" description:"id of the version which the pipeline record is related to"`
	Trigger       string                 `bson:"trigger,omitempty" json:"trigger,omitempty" description:"trigger of the pipeline record"`
	PerformParams *PipelinePerformParams `bson:"performParams,omitempty" json:"performParams,omitempty" description:"perform params of the pipeline record"`
	StageStatus   *StageStatus           `bson:"stageStatus,omitempty" json:"stageStatus,omitempty" description:"status of each pipeline stage"`
	Status        Status                 `bson:"status,omitempty" json:"status,omitempty" description:"status of the pipeline record"`
	StartTime     time.Time              `bson:"startTime,omitempty" json:"startTime,omitempty" description:"start time of the pipeline record"`
	EndTime       time.Time              `bson:"endTime,omitempty" json:"endTime,omitempty" description:"end time of the pipeline record"`
	// The version name, e.g. v1.0.1. This is used as docker image tag directly.
	Name string `bson:"name,omitempty" json:"name,omitempty"`
	// Ref reference of git repo, support branch, tag
	Ref string `bson:"ref,omitempty" json:"ref,omitempty"`
	// A short, human-readable description of the version.
	Description string `bson:"description,omitempty" json:"description,omitempty"`
	// Commit of the version (also known as revision, etc).
	Commit string `bson:"commit,omitempty" json:"commit,omitempty"`
	// Release version URL. This is used to find the release hosted on remote machine,
	// e.g. https://github.com/caicloud/cyclone/releases/v1.0.
	URL string `bson:"url,omitempty" json:"url,omitempty"`
	// Version status is the version's status information.
	Status api.VersionStatus `bson:"status,omitempty" json:"status,omitempty"`
	// Yaml deploy status is the current status of the version's deployment information.
	YamlDeployStatus api.VersionDeployStatus `bson:"yaml_deploy_status,omitempty" json:"yaml_deploy_status,omitempty"`
	// Operator is the version's operator.
	Operator api.VersionOperator `bson:"operator,omitempty" json:"operator,omitempty"`
	// Operation is the version's operation to execute.0000-1111 presents deploy,publish,integrate,build
	Operation int `bson:"operation,omitempty" json:"operation,omitempty"`
	// Deploy plans and status
	DeployPlansStatuses []api.DeployPlanStatus `bson:"deploy_plans_statuses,omitempty" json:"deploy_plans_statuses,omitempty"`
	// Flag of deploying with the information in yaml
	YamlDeploy api.YamlDeployFlag `bson:"yaml_deploy,omitempty" json:"yaml_deploy,omitempty"`
	// ProjectVersionID points to the version's projectVersion.
	ProjectVersionID string `bson:"projectversion_id,omitempty" json:"projectversion_id,omitempty"`
	// Final status is the version's final status information, finished or unfinished
	FinalStatus string `bson:"final_status,omitempty" json:"final_status,omitempty"`
	// BuildResource resoure for building image
	BuildResource api.BuildResource `bson:"build_resource,omitempty" json:"build_resource,omitempty"`
	// NewWorkerInfo
	Worker cloud.WorkerInfo `bson:"worker,omitempty" json:"worker,omitempty"`

	Data map[string]interface{} `bson:"data,omitempty" json:"data,omitempty"`
}

// Status can be the status of some pipeline record or some stage
type Status string

const (
	// Pending represents the status that is triggered but still not running.
	Pending Status = "Pending"
	// Running represents the status that is running.
	Running Status = "Running"
	// Success represents the status that finished and succeeded.
	Success Status = "Success"
	// Failed represents the status that finished but failed.
	Failed Status = "Failed"
	// Aborted represents the status that the stage was aborted by some reason, and we can get the reason from the log.
	Aborted Status = "Aborted"
)

// TODO The status of every stage may be different.
// StageStatus represents the collections of status for all stages.
type StageStatus struct {
	CodeCheckout    *GeneralStageStatus `bson:"codeCheckout,omitempty" json:"codeCheckout,omitempty" description:"status of code checkout stage"`
	UnitTest        *GeneralStageStatus `bson:"unitTest,omitempty" json:"unitTest,omitempty" description:"status of unit test stage"`
	CodeScan        *GeneralStageStatus `bson:"codeScan,omitempty" json:"codeScan,omitempty" description:"status of code scan stage"`
	Package         *GeneralStageStatus `bson:"package,omitempty" json:"package,omitempty" description:"status of package stage"`
	ImageBuild      *GeneralStageStatus `bson:"imageBuild,omitempty" json:"imageBuild,omitempty" description:"status of image build stage"`
	IntegrationTest *GeneralStageStatus `bson:"integrationTest,omitempty" json:"integrationTest,omitempty" description:"status of integration test stage"`
	ImageRelease    *GeneralStageStatus `bson:"imageRelease,omitempty" json:"imageRelease,omitempty" description:"status of image release stage"`
}

// GeneralStageStatus represents the information of stage.
type GeneralStageStatus struct {
	Status    Status    `bson:"status,omitempty" json:"status,omitempty" description:"status of the stage"`
	StartTime time.Time `bson:"startTime,omitempty" json:"startTime,omitempty" description:"start time of the stage"`
	EndTime   time.Time `bson:"endTime,omitempty" json:"endTime,omitempty" description:"end time of the stage"`
}

// ListMeta represents metadata that list resources must have.
type ListMeta struct {
	Total int `json:"total" description:"total items count"`
}

// ListResponse represents a collection of some resources.
type ListResponse struct {
	Meta  ListMeta    `json:"metadata" description:"pagination object"`
	Items interface{} `json:"items" description:"list resources"`
}

// QueryParams represents a collection of query param.
type QueryParams struct {
	Start  int                    `json:"start,omitempty" description:"query start index, default is 0"`
	Limit  int                    `json:"limit,omitempty" description:"specify the number of records, default is +Inf to not limit"`
	Filter map[string]interface{} `json:"filter,omitempty" description:"pattern to filter the records, default is nil to not filter"`
}

const (
	// Limit represents the name of the query parameter for pagination limit.
	Limit string = "limit"

	// Start represents the name of the query parameter for pagination start.
	Start string = "start"

	// Filter represents the name of the query parameter for filtering.
	Filter string = "filter"

	// RecentPipelineRecordCount represents the count of recent pipeline records.
	RecentPipelineRecordCount string = "recentCount"

	// RecentSuccessPipelineRecordCount represents the count of recent success pipeline records.
	RecentSuccessPipelineRecordCount string = "recentSuccessCount"

	// RecentFailedPipelineRecordCount represents the count of recent failed pipeline records.
	RecentFailedPipelineRecordCount string = "recentFailedCount"

	// Download represents the flag whether download pipeline record logs.
	Download string = "download"

	// Repo represents the query param for repo name.
	Repo string = "repo"
)

// ErrorResponse represents response of error.
type ErrorResponse struct {
	Message string `json:"message,omitempty"`
	Reason  string `json:"reason,omitempty"`
	Details string `json:"details,omitempty"`
}

// PipelinePerformParams the params to perform the pipeline.
type PipelinePerformParams struct {
	Ref          string   `bson:"ref,omitempty" json:"ref,omitempty" description:"reference of git repo, support branch, tag"`
	Name         string   `bson:"name,omitempty" json:"name,omitempty" description:"name of this running of pipeline"`
	Description  string   `bson:"description,omitempty" json:"description,omitempty" description:"description of this running of pipeline"`
	CreateSCMTag bool     `bson:"createScmTag,omitempty" json:"createScmTag,omitempty" description:"whether create tag in SCM"`
	Stages       []string `bson:"stages,omitempty" json:"stages,omitempty" description:"stages to be executed"`
}

// ScmToken represents a set of token informations of the projcet.
type ScmToken struct {
	ProjectID string       `bson:"projectId,omitempty" json:"projectId,omitempty" description:"id of the project which the token belongs to"`
	ScmType   string       `bson:"scmType,omitempty" json:"scmType,omitempty" description:"the type of scm, it can be github or gitlab"`
	Token     oauth2.Token `bson:"token,omitempty" json:"token,omitempty"`
}

// ListReposResponse represents a collection of repositories.
type ListReposResponse struct {
	Username  string       `json:"username,omitempty"`
	Repos     []Repository `json:"repos,omitempty"`
	AvatarURL string       `json:"avatarUrl,omitempty"`
}

// Repository represents the information of a repository.
type Repository struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

const (
	// GITHUB is the name of github.
	GITHUB string = "github"
	// GITLAB is the name of gitlab.
	GITLAB string = "gitlab"
)
