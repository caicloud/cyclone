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

import "time"

// Project represents a group to manage a set of related applications. It maybe a real project, which contains several or many applications.
type Project struct {
	ID          string    `bson:"_id,omitempty" json:"id,omitempty" description:"id of the project"`
	Name        string    `bson:"name,omitempty" json:"name,omitempty" description:"name of the project, should be unique"`
	Description string    `bson:"description,omitempty" json:"description,omitempty" description:"description of the project"`
	Owner       string    `bson:"owner,omitempty" json:"owner,omitempty" description:"owner of the project"`
	CreatedTime time.Time `bson:"createdTime,omitempty" json:"createdTime,omitempty" description:"created time of the project"`
	UpdatedTime time.Time `bson:"updatedTime,omitempty" json:"updatedTime,omitempty" description:"updated time of the project"`
}

// Pipeline represents a set of configs to describe the workflow of CI/CD.
type Pipeline struct {
	ID          string `bson:"_id,omitempty" json:"id,omitempty" description:"id of the pipeline"`
	Name        string `bson:"name,omitempty" json:"name,omitempty" description:"name of the pipeline，unique in one project"`
	Description string `bson:"description,omitempty" json:"description,omitempty" description:"description of the pipeline"`
	Owner       string `bson:"owner,omitempty" json:"owner,omitempty" description:"owner of the pipeline"`
	ProjectID   string `bson:"projectId,omitempty" json:"projectId,omitempty" description:"id of the project which the pipeline belongs to"`
	// TODO （robin）Remove the association between the pipeline and the service after pipeline replaces service.
	ServiceID   string       `bson:"serviceId,omitempty" json:"serviceId,omitempty" description:"id of the service which the pipeline is related to"`
	Build       *Build       `bson:"build,omitempty" json:"build,omitempty" description:"build spec of the pipeline"`
	AutoTrigger *AutoTrigger `bson:"autoTrigger,omitempty" json:"autoTrigger,omitempty" description:"auto trigger strategy of the pipeline"`
	CreatedTime time.Time    `bson:"createdTime,omitempty" json:"createdTime,omitempty" description:"created time of the pipeline"`
	UpdatedTime time.Time    `bson:"updatedTime,omitempty" json:"updatedTime,omitempty" description:"updated time of the pipeline"`
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
	CodeSources []*CodeSource `bson:"codeSources,omitempty" json:"codeSources,omitempty" description:"list of code sources to be checked out"`
}

// CodeSourceType represents the type of code source, supports gitlab, github and svn.
type CodeSourceType string

const (
	GitLab CodeSourceType = "gitlab"
	GitHub                = "github"
	SVN                   = "svn"
)

// CodeSource represents the config of code source, only one type is supported.
type CodeSource struct {
	Type CodeSourceType `bson:"type,omitempty" json:"type,omitempty" description:"type of code source, support gitlab, github and svn"`
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
	IntegrationTestSet *IntegrationTestSet `bson:"integrationTestSet,omitempty" json:"integrationTestSet,omitempty" description:"integration test config"`
	Services           []Service           `bson:"services,omitempty" json:"services,omitempty" description:"list of dependent services for integration test"`
}

// IntegrationTestSet represents the config for integration test set.
type IntegrationTestSet struct {
	Image   string   `bson:"image,omitempty" json:"image,omitempty" description:"built image to run the integration test"`
	Command []string `bson:"command,omitempty" json:"command,omitempty" description:"list of commands to run for integration test"`
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
	ReleasePolicy []ImageReleasePolicy `bson:"releasePolicy,omitempty" json:"releasePolicy,omitempty" description:"list of policies for image release"`
}

// ImageReleasePolicyType represents the type of image release policy.
type ImageReleasePolicyType string

const (
	// AlwaysRelease always releases the images.
	AlwaysRelease ImageReleasePolicyType = "Always"

	// NeverRelease never releases the images.
	NeverRelease ImageReleasePolicyType = "Never"

	// IntegrationTestSuccessRelease releases the images only when the integration test success.
	IntegrationTestSuccessRelease ImageReleasePolicyType = "IntegrationTestSuccess"
)

// ImageReleasePolicy represents the policy to release image.
type ImageReleasePolicy struct {
	Image string                 `bson:"image,omitempty" json:"image,omitempty" description:"image to be released"`
	Type  ImageReleasePolicyType `bson:"type,omitempty" json:"type,omitempty" description:"type of image release policy"`
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
	FinalStage string `bson:"finalStage,omitempty" json:"finalStage,omitempty" description:"final stage of the auto triggered running"`
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
	ID          string       `bson:"_id,omitempty" json:"id,omitempty" description:"id of the pipeline record"`
	PipelineID  string       `bson:"pipelineId,omitempty" json:"pipelineId,omitempty" description:"id of the related pipeline which the pipeline record belongs to"`
	VersionID   string       `bson:"versionId,omitempty" json:"versionId,omitempty" description:"id of the version which the pipeline record is related to"`
	Trigger     string       `bson:"trigger,omitempty" json:"trigger,omitempty" description:"trigger of the pipeline record"`
	StageStatus *StageStatus `bson:"stageStatus,omitempty" json:"stageStatus,omitempty" description:"status of each pipeline stage"`
	Status      Status       `bson:"status,omitempty" json:"status,omitempty" description:"status of the pipeline record"`
	StartTime   time.Time    `bson:"startTime,omitempty" json:"startTime,omitempty" description:"start time of the pipeline record"`
	EndTime     time.Time    `bson:"endTime,omitempty" json:"endTime,omitempty" description:"end time of the pipeline record"`
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
	// Failure represents the status that finished but failed.
	Failure Status = "Failed"
	// Abort represents the status that the stage was aborted by some reason, and we can get the reason from the log.
	Abort Status = "Aborted"
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
	Total       int `json:"total" description:"total items count"`
	ItemsLength int `json:"itemsLength" description:"returned items count"`
}

// ListResponse represents a collection of some resources.
type ListResponse struct {
	Meta  ListMeta    `json:"metadata" description:"pagination object"`
	Items interface{} `json:"items" description:"list resources"`
}

// QueryParams represents a collection of query param.
type QueryParams struct {
	Start int `json:"start,omitempty" description:"query start index, default 0"`
	Limit int `json:"limit,omitempty" description:"specify the number of records, default +Inf, return all"`
}

const (
	// Limit represents the name of the query parameter for pagination limit.
	Limit string = "limit"

	// Start represents the name of the query parameter for pagination start.
	Start string = "start"
)

// ErrorResponse represents responce of error.
type ErrorResponse struct {
	Message string `json:"message,omitempty"`
	Reason  string `json:"reason,omitempty"`
	Details string `json:"details,omitempty"`
}
