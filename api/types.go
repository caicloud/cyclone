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

package api

import (
	"time"

	"golang.org/x/oauth2"
)

const (
	// APIVersion is the version of API.
	// TODO: Make this package versioned. Right now, this is only used for indetifying
	// the endpoints, i.e. /api/v0.1/; we can't really do version control with it.
	APIVersion string = "v0.1"
)

// HealthCheckResponse is the response type for health check request.
type HealthCheckResponse struct {
	// Return the error message. This is used to provide user-facing errors.
	ErrorMessage string `json:"error_msg,omitempty"`
}

//
// Service related types
//

// Service is the management unit in release system.
type Service struct {
	// Service ID uniquely identifies the service.
	ServiceID string `bson:"_id,omitempty" json:"_id,omitempty"`
	// The user who owns the cluster.
	UserID string `bson:"user_id,omitempty" json:"user_id,omitempty"`
	// The deploy id
	DeployID string `bson:"deploy_id,omitempty" json:"deploy_id,omitempty"`
	// Service name, e.g. OrderSystem.
	Name string `bson:"name,omitempty" json:"name,omitempty"`
	// Username, just for build and push image.
	Username string `bson:"username,omitempty" json:"username,omitempty"`
	// A short, human-readable description of the service.
	Description string `bson:"description,omitempty" json:"description,omitempty"`
	// Build script path. When this script path (or maybe change to just text string)
	// is specified, we build image using BuildPath.
	// TODO: It's better to change to PreBuildHook and PostBuildHook to let Cyclone
	// control building image; otherwise, user can use arbitrary image name.
	BuildPath string `bson:"build_path,omitempty" json:"build_path,omitempty"`
	// A list of success version IDs.
	Versions []string `bson:"versions,omitempty" json:"versions,omitempty"`
	// A list of failed version IDs.
	VersionFails []string `bson:"version_fails,omitempty" json:"version_fails,omitempty"`
	// Repository information of the service.
	// TODO: For private repository, we need OAuth.
	Repository ServiceRepository `bson:"repository,omitempty" json:"repository,omitempty"`
	Jconfig    JenkinsConfig     `bson:"jconfig,omitempty" json:"jconfig,omitempty"`
	// Email porfile.
	Profile NotifyProfile `bson:"profile,omitempty" json:"profile,omitempty"`
	// Record last build version time or service create time first
	LastCreateTIme time.Time `bson:"last_createtime,omitempty" json:"last_createtime,omitempty"`
	// Record last build version name
	LastVersionName string `bson:"last_versionname,omitempty" json:"last_versionname,omitempty"`
	// Deploy plans
	DeployPlans []DeployPlan `bson:"deploy_plans,omitempty" json:"deploy_plans,omitempty"`
	// Repository information of the service.
	YAMLConfigName string `bson:"yaml_config_name,omitempty" json:"yaml_config_name,omitempty"`
}

// DeployPlan is the type for deployment plan.
type DeployPlan struct {
	// Plan name
	PlanName string `bson:"plan_name,omitempty" json:"plan_name,omitempty"`
	// Deploy config
	Config DeployConfig `bson:"config,omitempty" json:"config,omitempty"`
}

// DeployPlanStatus is the type for deployment plan config and status.
type DeployPlanStatus struct {
	// Plan name
	PlanName string `bson:"plan_name,omitempty" json:"plan_name,omitempty"`
	// Deploy config
	Config DeployConfig `bson:"config,omitempty" json:"config,omitempty"`
	// Deploy status
	Status VersionDeployStatus `bson:"status,omitempty" json:"status,omitempty"`
}

// DeployConfig is the type for deplyment config.
type DeployConfig struct {
	// Cluster name
	ClusterName string `bson:"cluster_name,omitempty" json:"cluster_name,omitempty"`
	// Cluster id
	ClusterID string `bson:"cluster_id,omitempty" json:"cluster_id,omitempty"`
	// namespace name
	Namespace string `bson:"namespace,omitempty" json:"namespace,omitempty"`
	// deployment name
	Deployment string `bson:"deployment,omitempty" json:"deployment,omitempty"`
	// container names
	Containers []string `bson:"containers,omitempty" json:"containers,omitempty"`
}

// JenkinsConfig is the type for jenkins config.
type JenkinsConfig struct {
	// Jenkins server address
	Address string `bson:"address,omitempty" json:"address,omitempty"`
	// Jenkins login username
	Username string `bson:"username,omitempty" json:"username,omitempty"`
	// Jenkins login password
	Password string `bson:"password,omitempty" json:"password,omitempty"`
}

// NotifySetting is the setting about whether send the email.
type NotifySetting string

const (
	// SendWhenFailed shows that Cyclone sends the build log email when version creation failed.
	SendWhenFailed NotifySetting = "sendwhenfailed"
	// SendWhenFinished shows that Cyclone sends the build log email when version creation finished.
	SendWhenFinished NotifySetting = "sendwhenfinished"
)

// VscToken is the type for token.
type VscToken struct {
	// The user who owns the cluster.
	UserID string `bson:"userid,omitempty" json:"user_id,omitempty"`

	Vsc string `bson:"vsc,omitempty" json:"vsc,omitempty"`

	Vsctoken oauth2.Token `bson:"vsctoken,omitempty" json:"vsctoken,omitempty"`
}

// VersionControlSystem is information about how service codebase is managed.
type VersionControlSystem string

const (
	// Git is the name of git.
	Git VersionControlSystem = "git"
	// Svn is the name of svn.
	Svn VersionControlSystem = "svn"
	// Mercurial is the name of mercurial.
	Mercurial VersionControlSystem = "mercurial"
	// Fake is the name of fake vcs system, a fake implementation for testing purpose.
	Fake VersionControlSystem = "fake"
)

const (
	// GITHUB is the name of github.
	GITHUB string = "github"
	// GITLAB is the name of gitlab.
	GITLAB string = "gitlab"
)

// RepositoryStatus is a summary of repository status.
type RepositoryStatus string

const (
	// RepositoryAccepted shows that the repositry create request is accepted, and Cyclone is preparing to clone it.
	RepositoryAccepted RepositoryStatus = "accepted"
	// RepositoryHealthy shows that the repository is healthy and ready to be used.
	RepositoryHealthy RepositoryStatus = "healthy"
	// RepositoryMissing shows that the repository is missing, meaning that Cyclone is unable to verity its existence.
	RepositoryMissing RepositoryStatus = "missing"
	// RepositoryUnknownVcs shows that the given vcs is not supported.
	RepositoryUnknownVcs RepositoryStatus = "unknownvcs"
	// RepositoryInternalError shows that the repositry creation has internal errors in Cyclone.
	RepositoryInternalError RepositoryStatus = "internalerror"
)

// WebhookResponse is the response type for webhook request.
type WebhookResponse struct {
	// Return the error message. This is used to provide user-facing errors.
	ErrorMessage string `json:"error_msg,omitempty"`
}

// ServiceRepository is all information about a repository hosting service codebase.
type ServiceRepository struct {
	// URL of the repository, e.g. https://github.com/caicloud/Cyclone.
	URL string `bson:"url,omitempty" json:"url,omitempty"`
	// Version control tool used to host the service repository.
	Vcs VersionControlSystem `bson:"vcs,omitempty" json:"vcs,omitempty"`
	// Subtype of Version control tool
	SubVcs string `bson:"subvcs,omitempty" json:"subvcs,omitempty"`
	// RepositoryStatus is current status of the repository, e.g. repository doesn't exist, etc.
	Status RepositoryStatus `bson:"status,omitempty" json:"status,omitempty"`
	// Username is used for Version Control System to operate
	Username string `bson:"username,omitempty" json:"username,omitempty"`
	// Password is used for Version Control System to operate
	Password string `bson:"password,omitempty" json:"password,omitempty"`
	// Webhook type, such as "github" "bitbuckect"
	Webhook string `bson:"webhook,omitempty" json:"webhook,omitempty"`
}

// ServiceCreationResponse is the response type for service creation request.
type ServiceCreationResponse struct {
	ServiceID string `json:"service_id,omitempty"`
	// Return the error message IFF not successful. This is used to provide user-facing errors.
	ErrorMessage string `json:"error_msg,omitempty"`
}

// ServiceGetResponse is the response type for service get request.
type ServiceGetResponse struct {
	Service Service `json:"service,omitempty"`
	// Return the error message IFF not successful. This is used to provide user-facing errors.
	ErrorMessage string `json:"error_msg,omitempty"`
}

// ServiceListResponse is the response type for service list request.
type ServiceListResponse struct {
	Services []Service `json:"services,omitempty"`
	// Return the error message IFF not successful. This is used to provide user-facing errors.
	ErrorMessage string `json:"error_msg,omitempty"`
}

// ServiceDelResponse is the response type for delete service request.
type ServiceDelResponse struct {
	Result string `json:"result,omitempty"`
	// Return the error message IFF not successful. This is used to provide user-facing errors.
	ErrorMessage string `json:"error_msg,omitempty"`
}

// ServiceSetResponse is the response type for service setting request.
type ServiceSetResponse struct {
	ServiceID string `json:"service_id,omitempty"`
	// Return the error message IFF not successful. This is used to provide user-facing errors.
	ErrorMessage string `json:"error_msg,omitempty"`
}

// RequestTokenResponse is the response type for service list request.
type RequestTokenResponse struct {
	Enterpoint string `json:"enterpoint,omitempty"`
	// Return the error message IFF not successful. This is used to provide user-facing errors.
	ErrorMessage string `json:"error_msg,omitempty"`
}

// AuthCallbackResponse is the response type for auth callback request.
type AuthCallbackResponse struct {
	Result string `json:"result,omitempty"`
	// Return the error message IFF not successful. This is used to provide user-facing errors.
	ErrorMessage string `json:"error_msg,omitempty"`
}

// Repo is the type for a VCS repo.
type Repo struct {
	Name  string `json:"name,omitempty"`
	Owner string `json:"owner,omitempty"`
	URL   string `json:"url,omitempty"`
}

// ListRepoResponse is the response type for repo list request.
type ListRepoResponse struct {
	Username  string `json:"username,omitempty"`
	Repos     []Repo `json:"repos,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
	// Return the error message IFF not successful. This is used to provide user-facing errors.
	ErrorMessage string `json:"error_msg,omitempty"`
}

// LogoutResponse is the response type for logout request.
type LogoutResponse struct {
	Result string `json:"result,omitempty"`
	// Return the error message IFF not successful. This is used to provide user-facing errors.
	ErrorMessage string `json:"error_msg,omitempty"`
}

// Profile of a user.
type Profile struct {
	Mail  string `json:"email"`
	Phone string `json:"cellphone"`
}

// NotifyProfile is the profile of the service.
type NotifyProfile struct {
	// User's profile.
	Profiles []Profile `bson:"profiles,omitempty" json:"profiles,omitempty"`
	// Notify Settings.
	Setting NotifySetting `bson:"setting,omitempty" json:"setting,omitempty"`
}

// GetProfileResponse is the response type for profile get request, sent by paging server.
type GetProfileResponse struct {
	ProfileInfo Profile `json:"profile"`
	Error       error   `json:"error"`
}

//
// Version related types
//

// Version associates with a service - a service can have multiple versions.
type Version struct {
	// VersionID uniquely identifies the version.
	VersionID string `bson:"_id,omitempty" json:"_id,omitempty"`
	// ServiceID points to the version's service.
	ServiceID string `bson:"service_id,omitempty" json:"service_id,omitempty"`
	// Service Name uniquely name the service.
	ServiceName string `bson:"service_name,omitempty" json:"service_name,omitempty"`
	// LogID points to the version's log.
	LogID string `bson:"version_log_id,omitempty" json:"version_log_id,omitempty"`
	// The version name, e.g. v1.0.1. This is used as docker image tag directly.
	Name string `bson:"name,omitempty" json:"name,omitempty"`
	// A short, human-readable description of the version.
	Description string `bson:"description,omitempty" json:"description,omitempty"`
	// A list of dependencies, reference to other versions.
	// TODO: This is only used as a FYI. It's hard to manage composite application.
	DependentVersionIDs []string `bson:"dependent_version_ids,omitempty" json:"dependent_version_ids,omitempty"`
	// Tags is a list of version tags. A version can have multiple tags, including
	// system tags and custom tags. System tags is added by Cyclone and custom tags
	// are added by a user. A version can have multiple tags, e.g. latest version
	// can also be live version. See below for a list of system tags.
	// TODO: Make tag a map instead of version.
	Tags []VersionTag `bson:"tags,omitempty" json:"tags,omitempty"`
	// LiveInfo is only valid when the version is tagged with "live". It contains
	// information about where the version is deployed, when it is deployed, etc.
	// A version can run on multiple cluster/project), so we have a list of LiveInfo.
	LiveInfo []VersionLiveInfo `bson:"live_info,omitempty" json:"live_info,omitempty"`
	// Commit of the version (also known as revision, etc).
	Commit string `bson:"commit,omitempty" json:"commit,omitempty"`
	// Time when the version is created.
	CreateTime time.Time `bson:"create_time,omitempty" json:"create_time,omitempty"`
	// Release version URL. This is used to find the release hosted on remote machine,
	// e.g. https://github.com/caicloud/cyclone/releases/v1.0.
	URL string `bson:"url,omitempty" json:"url,omitempty"`
	// Version status is the version's status information.
	Status VersionStatus `bson:"status,omitempty" json:"status,omitempty"`
	// Yaml deploy status is the current status of the version's deployment information.
	YamlDeployStatus VersionDeployStatus `bson:"yaml_deploy_status,omitempty" json:"yaml_deploy_status,omitempty"`
	// Operation is the version's operation to execute.
	Operation VersionOperation `bson:"operation,omitempty" json:"operation,omitempty"`
	// Operator is the version's operator.
	Operator VersionOperator `bson:"operator,omitempty" json:"operator,omitempty"`
	// Deploy plans and status
	DeployPlansStatuses []DeployPlanStatus `bson:"deploy_plans_statuses,omitempty" json:"deploy_plans_statuses,omitempty"`
	// Flag of deploying with the information in yaml
	YamlDeploy YamlDeployFlag `bson:"yaml_deploy,omitempty" json:"yaml_deploy,omitempty"`
	// Version build error message if any.
	ErrorMessage string `bson:"error_message,omitempty" json:"error_message,omitempty"`
	// ProjectVersionID points to the version's projectVersion.
	ProjectVersionID string `bson:"projectversion_id,omitempty" json:"projectversion_id,omitempty"`
	// Final status is the version's final status information, finished or unfinished
	FinalStatus string `bson:"final_status,omitempty" json:"final_status,omitempty"`
	// SecurtiyCheck for built image
	SecurityCheck bool `bson:"security_check,omitempty" json:"security_check,omitempty"`
	// Securtiy info for built image
	SecurityInfo []Security `bson:"security_info,omitempty" json:"security_info,omitempty"`
	// BuildResource resoure for building image
	BuildResource BuildResource `bson:"build_resource,omitempty" json:"build_resource,omitempty"`
}

// BuildResource is config of resource for building image
type BuildResource struct {
	// The memory config
	Memory float64 `bson:"memory,omitempty" json:"memory,omitempty"`
	// The cpu config
	CPU float64 `bson:"cpu,omitempty" json:"cpu,omitempty"`
}

// Security is the type for security check.
type Security struct {
	Name        string `bson:"name,omitempty" json:"name,omitempty"`
	Description string `bson:"description,omitempty"  json:"description,omitempty"`
	Severity    string `bson:"severity,omitempty" json:"severity,omitempty"`
}

// VersionLog is the version log.
type VersionLog struct {
	// LogID uniquely identifies the version version.
	LogID string `bson:"_id,omitempty" json:"_id,omitempty"`
	// VerisonID points to the log.
	VerisonID string `bson:"version_id,omitempty" json:"version_id,omitempty"`
	// Logs defines the logs when the version is created.
	Logs string `bson:"logs,omitempty" json:"logs,omitempty"`
}

// YamlDeployFlag is the type of version deployment flag.
type YamlDeployFlag string

const (
	// DeployWithYaml shows that the deployment is deployed with yaml.
	DeployWithYaml YamlDeployFlag = "yes"
	// NotDeployWithYaml shows that the deployment is deployed without yaml.
	NotDeployWithYaml YamlDeployFlag = "no"
)

// VersionTag is the type of version tags.
type VersionTag string

// System tags. In the future, we may support more system version tag, e.g. "stage",
// "test". User defined version tag can also be applied to a version, but it's not
// pre-defined.
const (
	LatestVersion VersionTag = "latest"
	LiveVersion   VersionTag = "live"
	NormalVersion VersionTag = "normal"
)

// VersionStatus is the type for version status.
type VersionStatus string

// VersionStatus defines the status of a version, e.g. in process of building image,
// image built, etc.
const (
	VersionFailed  VersionStatus = "failed"
	VersionHealthy VersionStatus = "healthy"
	VersionPending VersionStatus = "pending"
	VersionCancel  VersionStatus = "cancelled"
	VersionRunning VersionStatus = "running"
)

// CIStatus defines the status of a ci
const (
	CISuccess string = "success"
	CIFailure string = "failure"
	CIPending string = "pending"
)

// VersionDeployStatus is the type for version status.
type VersionDeployStatus string

const (
	// DeployNoRun shows that the version's deployment wouldn't be run.
	DeployNoRun VersionDeployStatus = "norun"
	// DeployPending shows that the version's deployment is pending.
	DeployPending VersionDeployStatus = "pending"
	// DeploySuccess shows that the version's deployment is successful.
	DeploySuccess VersionDeployStatus = "success"
	// DeployFailed shows that the version's deployment is failed.
	DeployFailed VersionDeployStatus = "failed"
)

// VersionOperation defines the operations of a version
type VersionOperation string

const (
	// IntegrationOperation is the version operation.
	IntegrationOperation VersionOperation = "integration"
	// PublishOperation is the version operation.
	PublishOperation VersionOperation = "publish"
	// DeployOperation is only used for project deploy，when have pushed image
	// web can not use this value
	DeployOperation VersionOperation = "deploy"
)

// VersionOperator defines the operator of a version
type VersionOperator string

const (
	// WebhookOperator is webhook operator.
	WebhookOperator VersionOperator = "webhook"
	// APIOperator is api operator.
	APIOperator VersionOperator = "api"
)

// AutoCreateTagFlag is the default tag postfix.
const AutoCreateTagFlag = "_Cyclone"

// VersionLiveInfo contains information about how a version is deployed.
type VersionLiveInfo struct {
	// Which cluster and project was the version deployed.
	Cluster string `bson:"cluster,omitempty" json:"cluster,omitempty"`
	Project string `bson:"project,omitempty" json:"project,omitempty"`
	// When was the version deployed on the cluster/project.
	DeployTime time.Time `bson:"deploy_time,omitempty" json:"deploy_time,omitemtpy"`
}

// KeepPoilcyRule is the rule of policy.
// In the future, keep policy can be more descriptive and powerful, e.g. delete this version
// if newer version is live in production for a long time.
type KeepPoilcyRule string

const (
	KeepPoilcyForever   KeepPoilcyRule = "forever"
	KeepPoilcyTimeBound KeepPoilcyRule = "timebound"
)

// VersionKeepPolicy is the type for policy of version.
type VersionKeepPolicy struct {
	Policy     KeepPoilcyRule
	DeleteTime time.Time
}

// DeploymentKind is the type for kind of deployment.
type DeploymentKind string

const (
	// DeploymentKindNew is a kind of deployment, represents "deploy a new version".
	DeploymentKindNew DeploymentKind = "new"
	// DeploymentKindRollingUpgrade is a kind of deployment, represents "rolling upgrade from an old version".
	DeploymentKindRollingUpgrade DeploymentKind = "rollingupgrade"
	// DeploymentKindRollack is a kind of deployment, represents "rollback to an old version".
	DeploymentKindRollack DeploymentKind = "rollback"
)

// VersionDeployment is the type for deployment.
type VersionDeployment struct {
	// The version to deploy.
	VersionID string
	// Kind of the deployment, e.g. deploy a new version, rolling upgrade from an old version.
	DeploymentKind DeploymentKind
	// Information needed to deploy the version.
	VersionLiveInfo VersionLiveInfo
}

// VersionCreationResponse is the response type for version creation request.
type VersionCreationResponse struct {
	VersionID string `json:"version_id,omitempty"`
	// Return the error message IFF not successful. This is used to provide user-facing errors.
	ErrorMessage string `json:"error_msg,omitempty"`
}

// VersionGetResponse is the response type for version get request.
type VersionGetResponse struct {
	Version Version `json:"version,omitempty"`
	// Return the error message IFF not successful. This is used to provide user-facing errors.
	ErrorMessage string `json:"error_msg,omitempty"`
}

// VersionLogCreateResponse is the response type for log create request.
type VersionLogCreateResponse struct {
	// LogID uniquely identifies the version version.
	LogID string `bson:"_id,omitempty" json:"_id,omitempty"`
	// Return the error message IFF not successful. This is used to provide user-facing errors.
	ErrorMessage string `json:"error_msg,omitempty"`
}

// VersionLogGetResponse is the response type for version log get request.
type VersionLogGetResponse struct {
	Logs string `json:"logs,omitempty"`
	// Return the error message IFF not successful. This is used to provide user-facing errors.
	ErrorMessage string `json:"error_msg,omitempty"`
}

// VersionListResponse is the response type for service list request.
type VersionListResponse struct {
	Versions []Version `json:"versions,omitempty"`
	// Return the error message IFF not successful. This is used to provide user-facing errors.
	ErrorMessage string `json:"error_msg,omitempty"`
}

// VersionBuildResponse is the response type for version build request.
type VersionBuildResponse struct {
	// Return the error message IFF not successful. This is used to provide user-facing errors.
	ErrorMessage string `json:"error_msg,omitempty"`
}

// FilterResponse is the response type for filtered request.
type FilterResponse struct {
	// Return the error message IFF not successful. This is used to provide user-facing errors.
	ErrorMessage string `json:"error_msg,omitempty"`
}

// VersionConcelResponse is the response type for version cancel request.
type VersionConcelResponse struct {
	Result string `json:"result,omitempty"`
	// Return the error message IFF not successful. This is used to provide user-facing errors.
	ErrorMessage string `json:"error_msg,omitempty"`
}

// SMTPServerConfig deifnes the config of SMTP Server.
type SMTPServerConfig struct {
	SMTPServer   string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
}

// ResourceCompose that compose the info about the resource
type ResourceCompose struct {
	// The total memory for userid
	MemoryForUser int64 `json:"memoryforuser,omitempty"`
	// The  memory for container
	MemoryForContainer int64 `json:"memoryforcontainer,omitempty"`
	// The total cpu for userid
	CPUForUser int64 `json:"ccpuforuser,omitempty"`
	// The cpu for container
	CPUForContainer int64 `json:"cpuforcontainer,omitempty"`
}

// RegistryCompose that compose the info about the registry
type RegistryCompose struct {
	// Registry's address, ie. cargo.caicloud.io
	RegistryLocation string `json:"registrylocation,omitempty"`
	// RegistryUsername used for operating the images
	RegistryUsername string `json:"registryusername,omitempty"`
	// RegistryPassword used for operating the images
	RegistryPassword string `json:"registrypassword,omitempty"`
}

//
// webhook related types
//

// WebhookGithub contains datas send from github webhook.
type WebhookGithub map[string]interface{}

// Github webhook event type
const (
	GithubWebhookIgnore      string = "ignore"
	GithubWebhookPush        string = "push"
	GithubWebhookPullRequest string = "pull_request"
	GithubWebhookRelease     string = "release"
)

// Github webhook payload flag
const (
	GithubWebhookFlagRelease string = "release"
	GithubWebhookFlagCommit  string = "head_commit"
	GithubWebhookFlagPR      string = "pull_request"
	GithubWebhookFlagAction  string = "action"
	GithubWebhookFlagRef     string = "ref"
	GithubWebhookFlagTags    string = "tags"
)

// Github pull request actions
const (
	PRActionOpened      string = "opened"
	PRActionSynchronize string = "synchronize"
)

// WebhookSVN contains datas send from svn webhook.
type WebhookSVN struct {
	// svn url, ie. svn://192.168.0.2/sample_repo
	URL string `json:"url,omitempty"`
	// event type of webhook
	Event string `json:"event,omitempty"`
	// id of commit
	CommitID string `json:"commit_id,omitempty"`
}

// SVN webhook event type
const (
	SVNWebhookCommit string = "commit"
)

// WebhookGitlab is type for gitlab webhook request.
type WebhookGitlab map[string]interface{}

// Gitlab webhook event type
const (
	GitlabWebhookIgnore      string = "ignore"
	GitlabWebhookPush        string = "push"
	GitlabWebhookPullRequest string = "merge_request"
	GitlabWebhookRelease     string = "tag_push"
)

// Gitlab webhook payload flag
const (
	GitlabWebhookFlagRef  string = "ref"
	GitlabWebhookFlagTags string = "tags"
)

// Daemon records value that is used to store in db
type Daemon struct {
	//VersionID is the cancelled version's ID
	VersionID string `bson:"versionid,omitempty" json:"versionid,omitempty"`
	// DaemonPoolID is the docker daemon pool ID
	DaemonPoolID string `bson:"daemonpool_id,omitempty" json:"daemonpool_id,omitempty"`
	// EnterPoint is the docker daemon's enterpoint
	EnterPoint string `bson:"enterpoint,omitempty" json:"enterpoint,omitempty"`
}

//
// project related types
//

// Project is the management unit in release system.
type Project struct {
	// Project ID uniquely identifies the project.
	ProjectID string `bson:"_id,omitempty" json:"_id,omitempty"`
	// The user who owns the cluster.
	UserID string `bson:"user_id,omitempty" json:"user_id,omitempty"`
	// Project name, e.g. OrderSystem.
	Name string `bson:"name,omitempty" json:"name,omitempty"`
	// A short, human-readable description of the project.
	Description string `bson:"description,omitempty" json:"description,omitempty"`
	// Time when the version is created.
	CreateTime time.Time `bson:"create_time,omitempty" json:"create_time,omitempty"`
	// A list of all project version IDs.
	Versions []string `bson:"versions,omitempty" json:"versions,omitempty"`
	// Dependency of the services.
	Services []ServiceDependency `bson:"services,omitempty" json:"services,omitempty"`
	// Work flow of the services.
	WorkFlow []string `bson:"work_flow,omitempty" json:"work_flow,omitempty"`
}

// ServiceDependency is define the dependency of the services
type ServiceDependency struct {
	// Service ID uniquely identifies the service.
	ServiceID string `bson:"service_id,omitempty" json:"service_id,omitempty"`
	// Service Name uniquely name the service.
	ServiceName string `bson:"service_name,omitempty" json:"service_name,omitempty"`
	// Depend special the depency of this service in project.
	Depend []ServiceDependency `bson:"depend,omitempty" json:"depend,omitempty"`
}

// ProjectCreationResponse is the response type for project set request.
type ProjectCreationResponse struct {
	ProjectID string `json:"project_id,omitempty"`
	// Return the error message IFF not successful. This is used to provide user-facing errors.
	ErrorMessage string `json:"error_msg,omitempty"`
}

// ProjectDelResponse is the response type for project delete request.
type ProjectDelResponse struct {
	Result string `json:"result,omitempty"`
	// Return the error message IFF not successful. This is used to provide user-facing errors.
	ErrorMessage string `json:"error_msg,omitempty"`
}

// ProjectSetResponse is the response type for project setting request.
type ProjectSetResponse struct {
	ProjectID string `json:"project_id,omitempty"`
	// Return the error message IFF not successful. This is used to provide user-facing errors.
	ErrorMessage string `json:"error_msg,omitempty"`
}

// ProjectListResponse is the response type for project list request.
type ProjectListResponse struct {
	Projects []Project `json:"projects,omitempty"`
	// Return the error message IFF not successful. This is used to provide user-facing errors.
	ErrorMessage string `json:"error_msg,omitempty"`
}

// ProjectVersion associates with a project - a project can have multiple versions.
type ProjectVersion struct {
	// VersionID uniquely identifies the version.
	VersionID string `bson:"_id,omitempty" json:"_id,omitempty"`
	// The user who owns the cluster.
	UserID string `bson:"user_id,omitempty" json:"user_id,omitempty"`
	// ServiceID points to the version's service.
	ProjectID string `bson:"project_id,omitempty" json:"project_id,omitempty"`
	// The version name, e.g. v1.0.1. This is used as docker image tag directly.
	Name string `bson:"name,omitempty" json:"name,omitempty"`
	// A short, human-readable description of the version.
	Description string `bson:"description,omitempty" json:"description,omitempty"`
	// Policy of the publish (also known as manual, QA, UAT, etc).
	Policy string `bson:"policy,omitempty" json:"policy,omitempty"`
	// Time when the version is created.
	CreateTime time.Time `bson:"create_time,omitempty" json:"create_time,omitempty"`
	// Version status is the version's status information.
	Status VersionStatus `bson:"status,omitempty" json:"status,omitempty"`
	// Dependency of the services.
	Services []ServiceDependency `bson:"services,omitempty" json:"services,omitempty"`
	// Tasks of this version publishing.
	Tasks []Version `bson:"tasks,omitempty" json:"tasks,omitempty"`
	// Version build error message if any.
	ErrorMessage string `bson:"error_message,omitempty" json:"error_message,omitempty"`
}

// ProjectVersionCreationResponse is the response type for project set request.
type ProjectVersionCreationResponse struct {
	ProjectVersionID string `json:"project_version_id,omitempty"`
	// Return the error message IFF not successful. This is used to provide user-facing errors.
	ErrorMessage string `json:"error_msg,omitempty"`
}

// ProjectVersionListResponse is the response type for project service list request.
type ProjectVersionListResponse struct {
	ProjectVersions []ProjectVersion `json:"project_versions,omitempty"`
	// Return the error message IFF not successful. This is used to provide user-facing errors.
	ErrorMessage string `json:"error_msg,omitempty"`
}

// ProjectGetResponse is the response type for project get request.
type ProjectGetResponse struct {
	Project Project `json:"project,omitempty"`
	// Return the error message IFF not successful. This is used to provide user-facing errors.
	ErrorMessage string `json:"error_msg,omitempty"`
}

// ProjectVersionGetResponse is the response type for project version get request.
type ProjectVersionGetResponse struct {
	ProjectVersion ProjectVersion `json:"projectversion,omitempty"`
	// Return the error message IFF not successful. This is used to provide user-facing errors.
	ErrorMessage string `json:"error_msg,omitempty"`
}

//
// event related types
//

// EventID is the type for event id.
type EventID string

// Operation is the type for operation in the Event.
type Operation string

// Event is an event occurred in the system. It is managed by AsycManager
// which will pass the event to appropriate handler.
type Event struct {
	// ID of the event, uniquely identifies the event.
	EventID EventID `bson:"event_id,omitempty" json:"event_id,omitempty"`
	// Operation the event is about, like cloning repository event, building
	// image event, etc.
	Operation Operation `bson:"operation,omitempty" json:"operation,omitempty"`
	// Almost all Cyclone event deal with user, service, version - the core
	// concept in Cyclone; therefore, event carries these properties. Note
	// sender might not set all of them.
	Service        Service        `bson:"service,omitempty" json:"service,omitempty"`
	Version        Version        `bson:"version,omitempty" json:"version,omitempty"`
	Project        Project        `bson:"project,omitempty" json:"project,omitempty"`
	ProjectVersion ProjectVersion `bson:"project_version,omitempty" json:"project_version,omitempty"`
	// Custom data passed to event operation handler.
	Data       map[string]interface{} `bson:"data,omitempty" json:"data,omitempty"`
	WorkerInfo WorkerInfo             `bson:"worker_info,omitempty" json:"worker_info,omitempty"`
	// The status of the event. External system can retrieve the event and
	// examine the status of the event, including PostHook.
	Status EventStatus `bson:"status,omitempty" json:"status,omitempty"`
	// In case of error, ErrorMessage holds the messge for end user.
	ErrorMessage string `bson:"error_msg,omitempty" json:"error_msg,omitempty"`
}

// EventStatus contains the status of an event.
type EventStatus string

// event status type
const (
	EventStatusPending EventStatus = "pending"
	EventStatusRunning EventStatus = "running"
	EventStatusSuccess EventStatus = "success"
	EventStatusFail    EventStatus = "fail"
	EventStatusCancel  EventStatus = "cancel"
)

// GetEventResponse is the response type for getevent request.
type GetEventResponse struct {
	Event Event `json:"event,omitempty"`
	// Return the error message IFF not successful. This is used to provide user-facing errors.
	ErrorMessage string `json:"error_msg,omitempty"`
}

// SetEvent is the request body for setevent request.
type SetEvent struct {
	Event Event `json:"event,omitempty"`
}

// SetEventResponse is the response type for setevent request.
type SetEventResponse struct {
	// ID of the event, uniquely identifies the event.
	EventID EventID `bson:"event_id,omitempty" json:"event_id,omitempty"`
	// Return the error message IFF not successful. This is used to provide user-facing errors.
	ErrorMessage string `json:"error_msg,omitempty"`
}

// WorkerInfo save the woker info
type WorkerInfo struct {
	// worker docker host
	DockerHost string `json:"docker_host,omitempty"`
	// ID of the container, uniquely identifies the container.
	ContainerID string `json:"container_id,omitempty"`
	// DueTime due time of the event
	DueTime time.Time `json:"due_time,omitempty"`
	// The resource need to used by this event
	UsedResource BuildResource `json:"used_resource,omitempty"`
}

// Resource is the management for user
type Resource struct {
	// The user who owns the cluster.
	UserID string `bson:"user_id,omitempty" json:"user_id,omitempty"`
	// The total memory for user
	TotalResource BuildResource `bson:"total_resource,omitempty" json:"total_resource,omitempty"`
	// PerResource resoure for building image
	PerResource BuildResource `bson:"per_resource,omitempty" json:"per_resource,omitempty"`
	// The left memory for user
	LeftResource BuildResource `bson:"left_resource,omitempty" json:"left_resource,omitempty"`
}

// ResourceSetResponse is the response type for resource setting request.
type ResourceSetResponse struct {
	Result string `json:"result,omitempty"`
	// Return the error message IFF not successful. This is used to provide user-facing errors.
	ErrorMessage string `json:"error_msg,omitempty"`
}

// ResourceGetResponse is the response type for resource get request.
type ResourceGetResponse struct {
	Resource Resource `json:"resource,omitempty"`
	// Return the error message IFF not successful. This is used to provide user-facing errors.
	ErrorMessage string `json:"error_msg,omitempty"`
}

//
// worker node relate types
//

// WorkerNode is the information of worker node.
type WorkerNode struct {
	// ID of the worker node.
	NodeID string `bson:"_id,omitempty" json:"_id,omitempty"`
	// Name of the worker node.
	Name string `bson:"name,omitempty" json:"name,omitempty"`
	// A short, human-readable description of the node.
	Description string `bson:"description,omitempty" json:"description,omitempty"`
	// IP of the worker node.
	IP string `bson:"ip,omitempty" json:"ip,omitempty"`
	// DockerHost of the worker node.
	DockerHost string `bson:"docker_host,omitempty" json:"docker_host,omitempty"`
	// Type of the worker node.
	Type WorkerNodeType `bson:"type,omitempty" json:"type,omitempty"`
	// Total resouce of the worker node
	TotalResource NodeResource `bson:"total_resource,omitempty" json:"total_resource,omitempty"`
	// Left resouce of the worker node
	LeftResource NodeResource `bson:"left_resource,omitempty" json:"left_resource,omitempty"`
}

// WorkerNodeType is the type for node's type, such as "system".
type WorkerNodeType string

// Worker Node Type
const (
	SystemWorkerNode WorkerNodeType = "system"
	UserWorkerNode   WorkerNodeType = "user"
)

// NodeResource is type for resources config, such as CPU and memory.
type NodeResource struct {
	// The memory config
	Memory float64 `bson:"memory,omitempty" json:"memory,omitempty"`
	// The cpu config
	CPU float64 `bson:"cpu,omitempty" json:"cpu,omitempty"`
}

// WorkerNodeCreateResponse is the response type for worker node add request.
type WorkerNodeCreateResponse struct {
	NodeID string `json:"node_id,omitempty"`
	// Return the error message IFF not successful. This is used to provide user-facing errors.
	ErrorMessage string `json:"error_msg,omitempty"`
}

// WorkerNodeGetResponse is the response type for worker node get request.
type WorkerNodeGetResponse struct {
	WorkerNode WorkerNode `json:"worker_node,omitempty"`
	// Return the error message IFF not successful. This is used to provide user-facing errors.
	ErrorMessage string `json:"error_msg,omitempty"`
}

// WorkerNodesListResponse is the response type for worker nodes list request.
type WorkerNodesListResponse struct {
	WorkerNodes []WorkerNode `json:"worker_nodes,omitempty"`
	// Return the error message IFF not successful. This is used to provide user-facing errors.
	ErrorMessage string `json:"error_msg,omitempty"`
}

// WorkerNodeDelResponse is the response type for delete worker node request.
type WorkerNodeDelResponse struct {
	Result string `json:"result,omitempty"`
	// Return the error message IFF not successful. This is used to provide user-facing errors.
	ErrorMessage string `json:"error_msg,omitempty"`
}

// Deploy is the management unit in release system.
type Deploy struct {
	// The user who owns the cluster.
	DeployID string `bson:"_id,omitempty" json:"_id,omitempty"`
	// Service ID uniquely identifies the service.
	ServiceID string `bson:"service_id,omitempty" json:"service_id,omitempty"`
	// Deploy plans
	DeployPlans []DeployPlan `bson:"deploy_plans,omitempty" json:"deploy_plans,omitempty"`
}

// DeployCreationResponse is the response type for deploy creation request
type DeployCreationResponse struct {
	DeployID string `json:"deploy_id,omitempty"`
	// Return the error message IFF not successful. This is used to provide user-facing errors.
	ErrorMessage string `json:"error_msg,omitempty"`
}

// DeployGetResponse is the response type for deploy get request.
type DeployGetResponse struct {
	Deploy Deploy `json:"deploy,omitempty"`
	// Return the error message IFF not successful. This is used to provide user-facing errors.
	ErrorMessage string `json:"error_msg,omitempty"`
}

// DeployCreationResponse is the response type for deploy set request
type DeploySetResponse struct {
	DeployID string `json:"deploy_id,omitempty"`
	// Return the error message IFF not successful. This is used to provide user-facing errors.
	ErrorMessage string `json:"error_msg,omitempty"`
}
