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
	"time"

	"golang.org/x/oauth2"

	"github.com/caicloud/cyclone/cmd/worker/options"
)

// Project represents a group to manage a set of related applications. It maybe a real project, which contains several or many applications.
type Project struct {
	ID             string        `bson:"_id,omitempty" json:"id,omitempty" description:"id of the project"`
	Name           string        `bson:"name,omitempty" json:"name,omitempty" description:"name of the project, should be unique"`
	Alias          string        `bson:"alias,omitempty" json:"alias,omitempty" description:"alias of the project"`
	Description    string        `bson:"description,omitempty" json:"description,omitempty" description:"description of the project"`
	Owner          string        `bson:"owner,omitempty" json:"owner,omitempty" description:"owner of the project"`
	SCM            *SCMConfig    `bson:"scm,omitempty" json:"scm,omitempty" description:"scm config of the project"`
	Registry       *Registry     `bson:"registry,omitempty" json:"registry,omitempty" description:"registry config for image operations"`
	Worker         *WorkerConfig `bson:"worker,omitempty" json:"worker,omitempty" description:"worker config of the project"`
	CreationTime   time.Time     `bson:"creationTime,omitempty" json:"creationTime,omitempty" description:"creation time of the project"`
	LastUpdateTime time.Time     `bson:"lastUpdateTime,omitempty" json:"lastUpdateTime,omitempty" description:"last update time of the project"`
}

// Registry represents registry config for image operations of the project.
type Registry struct {
	Server     string `bson:"server,omitempty" json:"server,omitempty"`
	Repository string `bson:"repository,omitempty" json:"repository,omitempty"`
	Username   string `bson:"username,omitempty" json:"username,omitempty"`
	Password   string `bson:"password,omitempty" json:"password,omitempty"`
}

// WorkerConfig represents the config of worker for the pipelines of the project.
type WorkerConfig struct {
	Location         *WorkerLocation                    `bson:"location,omitempty" json:"location,omitempty"`
	DependencyCaches map[BuildToolName]*DependencyCache `bson:"dependencyCaches,omitempty" json:"dependencyCaches,omitempty" description:"dependency caches for worker to speed up"`
	Quota            *WorkerQuota                       `bson:"quota,omitempty" json:"quota,omitempty" description:"quota for cyclone worker"`
}

type WorkerQuota struct {
	LimitsCPU      string `bson:"limitsCPU,omitempty" json:"limitsCPU,omitempty"`
	LimitsMemory   string `bson:"limitsMemory,omitempty" json:"limitsMemory,omitempty"`
	RequestsCPU    string `bson:"requestsCPU,omitempty" json:"requestsCPU,omitempty"`
	RequestsMemory string `bson:"requestsMemory,omitempty" json:"requestsMemory,omitempty"`
}

type WorkerLocation struct {
	CloudName string `bson:"cloudName,omitempty" json:"cloudName,omitempty" description:"name of cloud to create the worker"`
	Namespace string `bson:"namespace,omitempty" json:"namespace,omitempty" description:"k8s namespace to create the worker"`
}

// DependencyCache represents the cache volume of dependency for CI.
type DependencyCache struct {
	Name string `bson:"name,omitempty" json:"name,omitempty" description:"name of the dependency cache"`
}

// Pipeline represents a set of configs to describe the workflow of CI/CD.
type Pipeline struct {
	ID                   string            `bson:"_id,omitempty" json:"id,omitempty" description:"id of the pipeline"`
	Name                 string            `bson:"name,omitempty" json:"name,omitempty" description:"name of the pipeline，unique in one project"`
	Alias                string            `bson:"alias,omitempty" json:"alias,omitempty" description:"alias of the pipeline"`
	Description          string            `bson:"description,omitempty" json:"description,omitempty" description:"description of the pipeline"`
	Owner                string            `bson:"owner,omitempty" json:"owner,omitempty" description:"owner of the pipeline"`
	ProjectID            string            `bson:"projectID,omitempty" json:"projectID,omitempty" description:"id of the project which the pipeline belongs to"`
	Build                *Build            `bson:"build,omitempty" json:"build,omitempty" description:"build spec of the pipeline"`
	Notification         *Notification     `bson:"notification,omitempty" json:"notification,omitempty" description:"notification config of the pipeline"`
	AutoTrigger          *AutoTrigger      `bson:"autoTrigger,omitempty" json:"autoTrigger,omitempty" description:"auto trigger strategy of the pipeline"`
	CreationTime         time.Time         `bson:"creationTime,omitempty" json:"creationTime,omitempty" description:"creation time of the pipeline"`
	LastUpdateTime       time.Time         `bson:"lastUpdateTime,omitempty" json:"lastUpdateTime,omitempty" description:"last update time of the pipeline"`
	Annotations          map[string]string `bson:"annotations,omitempty" json:"annotations,omitempty" description:"pipeline annotations"`
	RecentRecords        []PipelineRecord  `bson:"-" json:"recentRecords,omitempty" description:"recent records of the pipeline"`
	RecentSuccessRecords []PipelineRecord  `bson:"-" json:"recentSuccessRecords,omitempty" description:"recent success records of the pipeline"`
	RecentFailedRecords  []PipelineRecord  `bson:"-" json:"recentFailedRecords,omitempty" description:"recent failed records of the pipeline"`
}

// Notification represents the notification config and stages of CI.
type Notification struct {
	Policy    NotificationPolicyType `bson:"policy" json:"policy" description:"notification policy, always,succuss,failure"`
	Receivers []*Receiver            `bson:"receivers" json:"receivers" description:"notification receivers' config"`
}

// Receiver represents the config of notification receiver.
type Receiver struct {
	Type      string   `bson:"type" json:"type" description:"receiver type, email,webhook,slack"`
	Addresses []string `bson:"addresses,omitempty" json:"addresses,omitempty" description:"receiver addresses"`
	Groups    []string `bson:"groups,omitempty" json:"groups,omitempty" description:"receiver groups"`
}

// Build represents the build config and stages of CI.
type Build struct {
	BuildInfo    *BuildInfo    `bson:"buildInfo,omitempty" json:"buildInfo,omitempty" description:"information to build package"`
	BuilderImage *BuilderImage `bson:"builderImage,omitempty" json:"builderImage,omitempty" description:"image information of the builder"`
	Stages       *BuildStages  `bson:"stages,omitempty" json:"stages,omitempty" description:"stages of CI"`
}

// BuildInfo represents the basic build information of the pipeline.
type BuildInfo struct {
	BuildTool       *BuildTool `bson:"buildTool,omitempty" json:"buildTool,omitempty" description:"tool to build package"`
	CacheDependency bool       `bson:"cacheDependency,omitempty" json:"cacheDependency,omitempty" description:"whether use dependency cache to speedup"`
}

type BuildToolName string

const (
	// MavenTool represents the Maven tool.
	MavenBuildTool BuildToolName = "Maven"

	// MavenTool represents the NPM tool.
	NPMBuildTool BuildToolName = "NPM"

	// GradleBuildTool represents the Gradle tool.
	GradleBuildTool BuildToolName = "Gradle"
)

// BuildTool represents the build tool for CI.
type BuildTool struct {
	Name    BuildToolName `bson:"name,omitempty" json:"name,omitempty" description:"name of build tool"`
	Version string        `bson:"version,omitempty" json:"version,omitempty" description:"version of build tool"`
}

// EnvVar represents the environment variables with name and value.
type EnvVar struct {
	Name  string `bson:"name,omitempty" json:"name,omitempty" description:"name of the environment variable"`
	Value string `bson:"value,omitempty" json:"value,omitempty" description:"value of the environment variable"`
}

// BuilderImage represents the image information of the builder.
type BuilderImage struct {
	Image   string   `bson:"image,omitempty" json:"image,omitempty" description:"image name of the builder"`
	EnvVars []EnvVar `bson:"envVars,omitempty" json:"envVars,omitempty" description:"environment variables of the builder"`
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
}

// GeneralStage represents the basic config shared by all stages.
type GeneralStage struct {
	Command []string `bson:"command,omitempty" json:"command,omitempty" description:"list of commands to run for this stage"`
}

// CodeCheckoutStage represents the config of code checkout stage.
type CodeCheckoutStage struct {
	MainRepo *CodeSource `bson:"mainRepo,omitempty" json:"mainRepo,omitempty" description:"main repository of code sources"`
	DepRepos []*DepRepo  `bson:"depRepos,omitempty" json:"depRepos,omitempty" description:"dependent repositories of code sources"`
}

// SCMType represents the type of SCM, supports gitlab, github and svn.
type SCMType string

const (
	Gitlab SCMType = "Gitlab"
	Github         = "Github"
	SVN            = "SVN"
)

// SCMAuthType represents the type of SCM auth, support password and token.
type SCMAuthType string

const (
	Password SCMAuthType = "Password"
	Token                = "Token"
)

// SVN username and password seperator, because SVN username can not contain ":".
const SVNUsernPwdSep string = ":"

// SCMConfig represents the config of SCM.
type SCMConfig struct {
	Type     SCMType     `bson:"type,omitempty" json:"type,omitempty" description:"SCM type, support gitlab, github and svn"`
	AuthType SCMAuthType `bson:"authType,omitempty" json:"authType,omitempty" description:"auth type, support password and token"`
	Server   string      `bson:"server,omitempty" json:"server,omitempty" description:"server of the SCM"`
	Username string      `bson:"username,omitempty" json:"username,omitempty" description:"username of the SCM"`
	Password string      `bson:"password,omitempty" json:"password,omitempty" description:"password of the SCM"`
	Token    string      `bson:"token,omitempty" json:"token,omitempty" description:"token of the SCM"`
}

// CodeSource represents the config of code source, only one type is supported.
type CodeSource struct {
	Type   SCMType    `bson:"type,omitempty" json:"type,omitempty" description:"type of code source, support gitlab, github and svn"`
	Gitlab *GitSource `bson:"gitlab,omitempty" json:"gitlab,omitempty" description:"code from gitlab"`
	Github *GitSource `bson:"github,omitempty" json:"github,omitempty" description:"code from github"`
	SVN    *GitSource `bson:"svn,omitempty" json:"svn,omitempty" description:"code from svn"`
}

// DepRepo represents the dependent repositories' config of code source.
type DepRepo struct {
	CodeSource `bson:",inline"`
	Folder     string `bson:"folder,omitempty" json:"folder,omitempty" description:"folder represents the place where the dependent repository lacated in"`
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
	GeneralStage `bson:",inline"`
	Outputs      []string `bson:"outputs,omitempty" json:"outputs,omitempty" description:"list of output path of this stage"`
}

// CodeScanStage represents the config of code scan stage.
type CodeScanStage struct {
	GeneralStage `bson:",inline"`
	Outputs      []string `bson:"outputs,omitempty" json:"outputs,omitempty" description:"list of output path of this stage"`
}

// PackageStage represents the config of package stage.
type PackageStage struct {
	GeneralStage `bson:",inline"`
	Outputs      []string `bson:"outputs,omitempty" json:"outputs,omitempty" description:"list of output path of this stage"`
}

// ImageBuildStage represents the config of image build stage.
type ImageBuildStage struct {
	BuildInfos []*ImageBuildInfo `bson:"buildInfos,omitempty" json:"buildInfos,omitempty" description:"list of output path of this stage"`
}

// ImageBuildInfo represents the config to build the image. Only one of Dockerfile and DockerfilePath needs to be set.
// If both of them are set, Dockerfile will be used with high priority.
type ImageBuildInfo struct {
	TaskName       string `bson:"taskName,omitempty" json:"taskName,omitempty" description:"task name for image build"`
	ContextDir     string `bson:"contextDir,omitempty" json:"contextDir,omitempty" description:"context directory for image build"`
	Dockerfile     string `bson:"dockerfile,omitempty" json:"dockerfile,omitempty" description:"dockerfile content for image build"`
	DockerfilePath string `bson:"dockerfilePath,omitempty" json:"dockerfilePath,omitempty" description:"dockerfile path for image build"`
	ImageName      string `bson:"imageName,omitempty" json:"imageName,omitempty" description:"name of the built image"`
}

// IntegrationTestStage represents the config of integration test stage.
type IntegrationTestStage struct {
	Config   *IntegrationTestConfig `bson:"Config,omitempty" json:"Config,omitempty" description:"integration test config"`
	Services []Service              `bson:"services,omitempty" json:"services,omitempty" description:"list of dependent services for integration test"`
}

// IntegrationTestConfig represents the config for integration test.
type IntegrationTestConfig struct {
	ImageName string   `bson:"imageName,omitempty" json:"imageName,omitempty" description:"built image name to run the integration test"`
	Command   []string `bson:"command,omitempty" json:"command,omitempty" description:"list of commands to run for integration test"`
	EnvVars   []EnvVar `bson:"envVars,omitempty" json:"envVars,omitempty" description:"environment variables for integration test"`
}

// Service represents the dependent service needed for integration test.
type Service struct {
	Name    string   `bson:"name,omitempty" json:"name,omitempty" description:"name of the service"`
	Image   string   `bson:"image,omitempty" json:"image,omitempty" description:"image name of the service"`
	Command []string `bson:"command,omitempty" json:"command,omitempty" description:"list of commands to start the service"`
	EnvVars []EnvVar `bson:"envVars,omitempty" json:"envVars,omitempty" description:"environment variables of the service"`
}

// ImageReleaseStage represents the config of image release stage.
type ImageReleaseStage struct {
	ReleasePolicies []ImageReleasePolicy `bson:"releasePolicies,omitempty" json:"releasePolicies,omitempty" description:"list of policies for image release"`
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
	PostCommit         *PostCommitTrigger         `bson:"postCommit,omitempty" json:"postCommit,omitempty" description:"post commit trigger strategy"`
	Push               *PushTrigger               `bson:"push,omitempty" json:"push,omitempty" description:"push trigger strategy"`
	TagRelease         *TagReleaseTrigger         `bson:"tagRelease,omitempty" json:"tagRelease,omitempty" description:"commit trigger strategy"`
	PullRequest        *PullRequestTrigger        `bson:"pullRequest,omitempty" json:"pullRequest,omitempty" description:"pull request trigger strategy"`
	PullRequestComment *PullRequestCommentTrigger `bson:"pullRequestComment,omitempty" json:"pullRequestComment,omitempty" description:"pull request comment trigger strategy"`

	Webhook string `bson:"webhook,omitempty" json:"webhook,omitempty" description:"webhook for the SCM trigger"`
}

// GeneralTrigger represents the general config for all auto trigger strategies.
type GeneralTrigger struct {
	Stages []PipelineStageName `bson:"stages,omitempty" json:"stages,omitempty" description:"stages to be executed when automatically triggered by SCM webhook"`
}

// CommitTrigger represents the trigger from SCM commit.
type CommitTrigger struct {
	GeneralTrigger `bson:",inline"`
}

// PullRequestTrigger represents the SCM auto trigger from pull request.
type PullRequestTrigger struct {
	GeneralTrigger `bson:",inline"`
}

// PullRequestCommentTrigger represents the SCM auto trigger from pull request comment.
type PullRequestCommentTrigger struct {
	GeneralTrigger `bson:",inline"`
	Comments       []string `bson:"comments" json:"comments" description:"pull request comments to trigger"`
}

// PushTrigger represents the SCM auto trigger from push into branches.
type PushTrigger struct {
	GeneralTrigger `bson:",inline"`
	Branches       []string `bson:"branches" json:"branches" description:"branches with new commit to trigger"`
}

// PostCommitTrigger represents the SCM auto trigger from SVN post_commit.
type PostCommitTrigger struct {
	GeneralTrigger `bson:",inline"`

	RepoInfo *RepoInfo `bson:"repoInfo" json:"repoInfo" description:"svn repository information"`
}

// RepoInfo contains svn repository information, id and root-url.
type RepoInfo struct {
	// ID represents SVN repository UUID, this ID is retrieved by cyclone-server by
	//
	// 'svn info --show-item repos-uuid --username {user} --password {password} --non-interactive
	// --trust-server-cert-failures unknown-ca,cn-mismatch,expired,not-yet-valid,other
	// --no-auth-cache {remote-svn-address}'
	//
	ID string `bson:"id" json:"id" description:"svn repository UUID"`

	// RootURL represents SVN repository root url, this root is retrieved by cyclone-server by
	//
	// 'svn info --show-item repos-root-url --username {user} --password {password} --non-interactive
	// --trust-server-cert-failures unknown-ca,cn-mismatch,expired,not-yet-valid,other
	// --no-auth-cache {remote-svn-address}'
	//
	RootURL string `bson:"rootURL" json:"rootURL" description:"svn repository root url"`
}

// TagReleaseTrigger represents the SCM auto trigger from tag release.
type TagReleaseTrigger struct {
	GeneralTrigger `bson:",inline"`
}

// CommitWithCommentsTrigger represents the trigger from SCM commit with specified comments.
type CommitWithCommentsTrigger struct {
	GeneralTrigger `bson:",inline"`
	Comment        []string `bson:"comment,omitempty" json:"comment,omitempty" description:"list of comments in commit"`
}

// TimerTrigger represents the auto trigger strategy from timer.
type TimerTrigger struct {
	CronTrigger []*CronTrigger `bson:"cronTrigger,omitempty" json:"cronTrigger,omitempty" description:"list of cron trigger strategies"`
}

// CronTrigger represents the trigger from cron job.
type CronTrigger struct {
	GeneralTrigger `bson:",inline"`
	Expression     string `bson:"expression,omitempty" json:"expression,omitempty" description:"expression of cron job"`
}

// PipelineRecord represents the running record of pipeline.
type PipelineRecord struct {
	ID              string                 `bson:"_id,omitempty" json:"id,omitempty" description:"id of the pipeline record"`
	Name            string                 `bson:"name,omitempty" json:"name,omitempty" description:"name of the pipeline record"`
	PipelineID      string                 `bson:"pipelineID,omitempty" json:"pipelineID,omitempty" description:"id of the related pipeline which the pipeline record belongs to"`
	Trigger         string                 `bson:"trigger,omitempty" json:"trigger,omitempty" description:"trigger of the pipeline record"`
	PerformParams   *PipelinePerformParams `bson:"performParams,omitempty" json:"performParams,omitempty" description:"perform params of the pipeline record"`
	StageStatus     *StageStatus           `bson:"stageStatus,omitempty" json:"stageStatus,omitempty" description:"status of each pipeline stage"`
	Status          Status                 `bson:"status,omitempty" json:"status,omitempty" description:"status of the pipeline record"`
	ErrorMessage    string                 `bson:"errorMessage,omitempty" json:"errorMessage,omitempty" description:"error message for the pipeline failure"`
	PRLastCommitSHA string                 `bson:"prLastCommitSHA,omitempty" json:"prLastCommitSHA,omitempty" description:"last commit sha of PR"`
	StartTime       time.Time              `bson:"startTime,omitempty" json:"startTime,omitempty" description:"start time of the pipeline record"`
	EndTime         time.Time              `bson:"endTime,omitempty" json:"endTime,omitempty" description:"end time of the pipeline record"`
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
	CodeCheckout    *CodeCheckoutStageStatus `bson:"codeCheckout,omitempty" json:"codeCheckout,omitempty" description:"status of code checkout stage"`
	UnitTest        *GeneralStageStatus      `bson:"unitTest,omitempty" json:"unitTest,omitempty" description:"status of unit test stage"`
	CodeScan        *GeneralStageStatus      `bson:"codeScan,omitempty" json:"codeScan,omitempty" description:"status of code scan stage"`
	Package         *GeneralStageStatus      `bson:"package,omitempty" json:"package,omitempty" description:"status of package stage"`
	ImageBuild      *ImageBuildStageStatus   `bson:"imageBuild,omitempty" json:"imageBuild,omitempty" description:"status of image build stage"`
	IntegrationTest *GeneralStageStatus      `bson:"integrationTest,omitempty" json:"integrationTest,omitempty" description:"status of integration test stage"`
	ImageRelease    *ImageReleaseStageStatus `bson:"imageRelease,omitempty" json:"imageRelease,omitempty" description:"status of image release stage"`
}

// GeneralStageStatus represents the information of stage.
type GeneralStageStatus struct {
	Status    Status    `bson:"status,omitempty" json:"status,omitempty" description:"status of the stage"`
	StartTime time.Time `bson:"startTime,omitempty" json:"startTime,omitempty" description:"start time of the stage"`
	EndTime   time.Time `bson:"endTime,omitempty" json:"endTime,omitempty" description:"end time of the stage"`
}

// TaskStatus represents the information of subtasks in one stage.
type TaskStatus struct {
	Name      string    `bson:"name,omitempty" json:"name,omitempty" description:"name of subtask in one stage"`
	Status    Status    `bson:"status,omitempty" json:"status,omitempty" description:"status of the stage"`
	StartTime time.Time `bson:"startTime,omitempty" json:"startTime,omitempty" description:"start time of the stage"`
	EndTime   time.Time `bson:"endTime,omitempty" json:"endTime,omitempty" description:"end time of the stage"`
}

// CodeCheckoutStageStatus includes GeneralStageStatus and pipelineRecord version.
type CodeCheckoutStageStatus struct {
	GeneralStageStatus `bson:",inline"`
	Commits            Commits `bson:"commits,omitempty" json:"commits,omitempty" description:"commits of the pipeline record"`
}

// ImageBuildStageStatus includes GeneralStageStatus and image build infos.
type ImageBuildTaskStatus struct {
	TaskStatus `bson:",inline"`
	Image      string `bson:"image,omitempty" json:"image,omitempty" description:"built image name"`
}

// ImageBuildStageStatus includes GeneralStageStatus and image build infos.
type ImageBuildStageStatus struct {
	GeneralStageStatus `bson:",inline"`
	Tasks              []*ImageBuildTaskStatus `bson:"tasks,omitempty" json:"tasks,omitempty" description:"task status of the stage"`
}

// ImageReleaseStageStatus includes GeneralStageStatus and Images.
type ImageReleaseStageStatus struct {
	GeneralStageStatus `bson:",inline"`
	Tasks              []*ImageReleaseTaskStatus `bson:"tasks,omitempty" json:"tasks,omitempty" description:"task status of the stage"`
}

// ImageReleaseTaskStatus includes GeneralStageStatus and image release infos.
type ImageReleaseTaskStatus struct {
	TaskStatus `bson:",inline"`
	Image      string `bson:"image,omitempty" json:"image,omitempty" description:"released image name"`
}

// ImageReleaseStageStatus includes GeneralStageStatus and Images.
type Commits struct {
	MainRepo *CommitLog   `bson:"mainRepo" json:"mainRepo" description:"main repository's' commit of code sources"`
	DepRepos []*CommitLog `bson:"depRepos,omitempty" json:"depRepos,omitempty" description:"dependent repositories' commit of code sources"`
}

type CommitLog struct {
	RepoName string    `bson:"repoName,omitempty" json:"repoName,omitempty" description:"repo name"`
	ID       string    `bson:"id,omitempty" json:"id,omitempty" description:"commint id"`
	Author   string    `bson:"author,omitempty" json:"author,omitempty" description:"author name"`
	Date     time.Time `bson:"date,omitempty" json:"date,omitempty" description:"author date"`
	Message  string    `bson:"message,omitempty" json:"message,omitempty" description:"commint message"`
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

	// StartTime represents the query param for stats start time.
	StartTime string = "startTime"

	// EndTime represents the query param for stats end time.
	EndTime string = "endTime"
)

// ErrorResponse represents response of error.
type ErrorResponse struct {
	Message string `json:"message,omitempty"`
	Reason  string `json:"reason,omitempty"`
	Details string `json:"details,omitempty"`
}

// PipelineStageName represents the name of stages in pipeline.
type PipelineStageName string

const (
	// CodeCheckoutStageName represents the name of code checkout stage.
	CodeCheckoutStageName PipelineStageName = "codeCheckout"

	// UnitTestStageName represents the name of unit test stage.
	UnitTestStageName PipelineStageName = "unitTest"

	// PackageStageName represents the name of package stage.
	PackageStageName PipelineStageName = "package"

	// ImageBuildStageName represents the name of image build stage.
	ImageBuildStageName PipelineStageName = "imageBuild"

	// IntegrationTestStageName represents the name of integration test stage.
	IntegrationTestStageName PipelineStageName = "integrationTest"

	// ImageReleaseStageName represents the name of image release stage.
	ImageReleaseStageName PipelineStageName = "imageRelease"
)

// PipelinePerformParams the params to perform the pipeline.
type PipelinePerformParams struct {
	Ref             string              `bson:"ref,omitempty" json:"ref,omitempty" description:"reference of git repo, support branch, tag"`
	Name            string              `bson:"name,omitempty" json:"name,omitempty" description:"name of this running of pipeline"`
	Description     string              `bson:"description,omitempty" json:"description,omitempty" description:"description of this running of pipeline"`
	CreateSCMTag    bool                `bson:"createScmTag,omitempty" json:"createScmTag,omitempty" description:"whether create tag in SCM"`
	CacheDependency bool                `bson:"cacheDependency,omitempty" json:"cacheDependency,omitempty" description:"whether use dependency cache to speedup"`
	Stages          []PipelineStageName `bson:"stages,omitempty" json:"stages,omitempty" description:"stages to be executed"`
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

// CloudType represents cloud type, supports Docker and Kubernetes.
type CloudType string

const (
	// CloudTypeDocker represents the Docker cloud type.
	CloudTypeDocker CloudType = "Docker"

	// CloudTypeKubernetes represents the Kubernetes cloud type.
	CloudTypeKubernetes CloudType = "Kubernetes"
)

type CloudDocker struct {
	Host     string `json:"host,omitempty" bson:"host,omitempty"`
	Insecure bool   `json:"insecure,omitempty" bson:"insecure,omitempty"`
	CertPath string `json:"certPath,omitempty" bson:"certPath,omitempty"`
}

type CloudKubernetes struct {
	Host            string           `json:"host,omitempty" bson:"host,omitempty"`
	InCluster       bool             `json:"inCluster,omitempty" bson:"inCluster,omitempty"`
	Namespace       string           `json:"namespace,omitempty" bson:"-"`
	BearerToken     string           `json:"bearerToken,omitempty" bson:"bearerToken,omitempty"`
	Username        string           `json:"username,omitempty" bson:"username"`
	Password        string           `json:"password,omitempty" bson:"password"`
	TLSClientConfig *TLSClientConfig `json:"tlsClientConfig,omitempty" bson:"tlsClientConfig"`
}

// +k8s:deepcopy-gen=true
// TLSClientConfig contains settings to enable transport layer security
type TLSClientConfig struct {
	// Server should be accessed without verifying the TLS certificate. For testing only.
	Insecure bool `json:"insecure,omitempty" bson:"insecure"`

	// Trusted root certificates for server
	CAFile string `json:"caFile,omitempty" bson:"caFile"`

	// CAData holds PEM-encoded bytes (typically read from a root certificates bundle).
	// CAData takes precedence over CAFile
	CAData []byte `json:"caData,omitempty" bson:"caData"`
}

// Cloud represents clouds for workers.
type Cloud struct {
	ID         string           `bson:"_id,omitempty" json:"id,omitempty"`
	Type       CloudType        `bson:"type,omitempty" json:"type,omitempty"`
	Name       string           `json:"name,omitempty" bson:"name,omitempty"`
	Docker     *CloudDocker     `json:"docker,omitempty" bson:"docker,omitempty"`
	Kubernetes *CloudKubernetes `json:"kubernetes,omitempty" bson:"kubernetes,omitempty"`
}

const (
	// GITHUB is the name of github.
	GITHUB string = "github"
	// GITLAB is the name of gitlab.
	GITLAB string = "gitlab"
)

// Event represents the pipeline perform event, which contains all basic information to perform the pipeline,
// and the event status in queue.
type Event struct {
	ID             string          `bson:"_id,omitempty" json:"_id,omitempty"`
	Project        *Project        `bson:"project,omitempty" json:"project,omitempty"`
	Pipeline       *Pipeline       `bson:"pipeline,omitempty" json:"pipeline,omitempty"`
	PipelineRecord *PipelineRecord `bson:"pipelineRecord,omitempty" json:"pipelineRecord,omitempty"`

	// WorkerInfo represents the worker infos used to delete the workers.
	WorkerInfo *WorkerInfo `bson:"workerInfo,omitempty" json:"workerInfo,omitempty"`

	// Retry represents the number of retry when cloud is busy.
	Retry       int         `bson:"retry,omitempty" json:"retry,omitempty"`
	InTime      time.Time   `bson:"inTime,omitempty" json:"inTime,omitempty"`
	OutTime     time.Time   `bson:"outTime,omitempty" json:"outTime,omitempty"`
	QueueStatus QueueStatus `bson:"queueStatus,omitempty" json:"queueStatus,omitempty"`
}

type WorkerInfo struct {
	// CacheVolume represents the volume to cache dependency for worker.
	CacheVolume string

	// MountPath represents the mount path for the cache volume.
	MountPath string

	// quota represents the resource quota for worker.
	Quota options.Quota

	Name string

	StartTime time.Time `json:"startTime,omitempty" bson:"startTime,omitempty"`
	DueTime   time.Time `json:"dueTime,omitempty" bson:"dueTime,omitempty"`
}

// PipelineStatusStats represents statistics of workspace or pipeline.
type PipelineStatusStats struct {
	Overview StatsOverview  `json:"overview"`
	Details  []*StatsDetail `json:"details"`
}

type StatsOverview struct {
	Total int `json:"total"`
	StatsStatus
	SuccessRatio string `json:"successRatio"`
}

type StatsDetail struct {
	Timestamp int64 `json:"timestamp"`
	StatsStatus
}

type StatsStatus struct {
	Success int `json:"success"`
	Failed  int `json:"failed"`
	Aborted int `json:"aborted"`
}

// QueueStatus represents the event status in the queue.
type QueueStatus string

const (
	InQueue  QueueStatus = "in"
	OutQueue QueueStatus = "out"
	Handling QueueStatus = "handling"
)

// pipeline record trigger type.
const (
	TriggerSCM  string = "webhook"
	TriggerCron string = "timer"

	// Fixme, this is a litter tricky.
	// SVNPostCommitRefPrefix is a flag used by svn code checkout;
	// If 'ref' with this prefix, we will checkout code frome a specific revision,
	// otherwise, apped ref to clone url, then do checkout.
	SVNPostCommitRefPrefix           string = "hook-post-commit-"
	TriggerSVNHookPostCommit         string = "hook-post-commit"
	TriggerWebhookPush               string = "webhook-push"
	TriggerWebhookTagRelease         string = "webhook-tag-release"
	TriggerWebhookPullRequest        string = "webhook-pull-request"
	TriggerWebhookPullRequestComment string = "webhook-pull-request-comment"
)

// WorkerInstance represents some infomation of cyclone worker instance, e.g. pod of k8s, container of docker.
type WorkerInstance struct {
	Name           string    `bson:"name,omitempty" json:"name,omitempty" description:"name of the worker pod, should be unique"`
	Status         string    `bson:"status,omitempty" json:"status,omitempty" description:"status of the worker pod"`
	CreationTime   time.Time `bson:"creationTime,omitempty" json:"creationTime,omitempty" description:"creation time of the worker pod"`
	LastUpdateTime time.Time `bson:"lastUpdateTime,omitempty" json:"lastUpdateTime,omitempty" description:"last update time of worker pod"`
	ProjectName    string    `bson:"projectName,omitempty" json:"projectName,omitempty" description:"name of the cyclone project"`
	PipelineName   string    `bson:"pipelineName,omitempty" json:"pipelineName,omitempty" description:"name of the pipeline"`
	RecordID       string    `bson:"recordID,omitempty" json:"recordID,omitempty" description:"id of the pipeline record"`
}

// Template contains some configurations of creating pipeline.
type Template struct {
	Name                 string `yaml:"name,omitempty" json:"name,omitempty" description:"name of the template, should be unique"`
	Type                 string `yaml:"type,omitempty" json:"type,omitempty" description:"type of the template"`
	BuilderImage         string `yaml:"builderImage,omitempty" json:"builderImage,omitempty" description:"image information of the builder"`
	TestCommands         string `yaml:"testCommands,omitempty" json:"testCommands,omitempty" description:"sample commands of unit test stage"`
	PackageCommands      string `yaml:"packageCommands,omitempty" json:"packageCommands,omitempty" description:"sample commands of package stage"`
	CustomizedDockerfile string `yaml:"customizedDockerfile,omitempty" json:"customizedDockerfile,omitempty" description:"sample Dockerfile"`
}

// TemplateType represents the type of the template
type TemplateType struct {
	Type string `yaml:"type,omitempty" json:"type,omitempty" description:"type of the template"`
}

const (
	// JavaScriptRepoType represents the language type JavaScript.
	JavaScriptRepoType string = "JavaScript"

	// JavaRepoType represents the language type Java.
	JavaRepoType string = "Java"

	// MavenRepoType represents the repository type Maven.
	MavenRepoType string = "Maven"

	// GradleRepoType represents the repository type Gradle.
	GradleRepoType string = "Gradle"

	// NodeRepoType represents the repository type NodeJS.
	NodeRepoType string = "NodeJS"
)

// NotificationContent contains some pipeline record infomation.
type NotificationContent struct {
	ProjectName      string    `bson:"projectName,omitempty" json:"projectName,omitempty" `
	PipelineName     string    `bson:"pipelineName,omitempty" json:"pipelineName,omitempty" `
	RecordName       string    `bson:"recordName,omitempty" json:"recordName,omitempty" `
	RecordID         string    `bson:"recordID,omitempty" json:"recordID,omitempty" `
	Trigger          string    `bson:"trigger,omitempty" json:"trigger,omitempty" description:"trigger of the pipeline record"`
	Status           Status    `bson:"status,omitempty" json:"status,omitempty" description:"status of the pipeline record"`
	ErrorMessage     string    `bson:"errorMessage,omitempty" json:"errorMessage,omitempty" description:"error message for the pipeline failure"`
	PipelinRecordURL string    `bson:"pipelinRecordURL,omitempty" json:"pipelinRecordURL,omitempty" description:"URL of the pipeline record"`
	StartTime        time.Time `bson:"startTime,omitempty" json:"startTime,omitempty" description:"start time of the pipeline record"`
	EndTime          time.Time `bson:"endTime,omitempty" json:"endTime,omitempty" description:"end time of the pipeline record"`
}

// NotificationPolicyType represents the type of notification
type NotificationPolicyType string

const (
	// AlwaysNotify represents always send notification regardless of the pipeline record result.
	AlwaysNotify NotificationPolicyType = "Always"

	// SuccessNotify represents send notification when pipeline record execute successfully.
	SuccessNotify NotificationPolicyType = "Success"

	// FailureNotify represents send notification when pipeline record execute failed.
	FailureNotify NotificationPolicyType = "Failure"
)

// TestResult contains some pipeline record test result.
type TestResult struct {
	FileName string `bson:"fileName,omitempty" json:"fileName,omitempty" `
}
