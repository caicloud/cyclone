package harbor

import (
	"time"
)

// 此文件中的所有内容是从 v1.4.0 harbor 的源码中 copy 出来，仅做了重命名的修改

// keys of project metadata and severity values
const (
	ProMetaPublic             = "public"
	ProMetaEnableContentTrust = "enable_content_trust"
	ProMetaPreventVul         = "prevent_vul" //prevent vulnerable images from being pulled
	ProMetaSeverity           = "severity"
	ProMetaAutoScan           = "auto_scan"
	SeverityNone              = "negligible"
	SeverityLow               = "low"
	SeverityMedium            = "medium"
	SeverityHigh              = "high"

	//FilterItemKindProject : Kind of filter item is 'project'
	FilterItemKindProject = "project"
	//FilterItemKindRepository : Kind of filter item is 'repository'
	FilterItemKindRepository = "repository"
	//FilterItemKindTag : Kind of filter item is 'tag'
	FilterItemKindTag = "tag"

	//TriggerKindImmediate : Kind of trigger is 'Immediate'
	TriggerKindImmediate = "Immediate"
	//TriggerKindSchedule : Kind of trigger is 'Scheduled'
	TriggerKindSchedule = "Scheduled"
	//TriggerKindManual : Kind of trigger is 'Manual'
	TriggerKindManual = "Manual"

	//TriggerScheduleDaily : type of scheduling is 'Daily'
	TriggerScheduleDaily = "Daily"
	//TriggerScheduleWeekly : type of scheduling is 'Weekly'
	TriggerScheduleWeekly = "Weekly"
)

// =================================================================================================

// HarborProject holds the details of a project.
type HarborProject struct {
	ProjectID    int64             `json:"project_id"`
	OwnerID      int               `json:"owner_id"`
	Name         string            `json:"name"`
	CreationTime time.Time         `json:"creation_time"`
	UpdateTime   time.Time         `json:"update_time"`
	Deleted      int               `json:"deleted"`
	OwnerName    string            `json:"owner_name"`
	Togglable    bool              `json:"togglable"`
	Role         int               `json:"current_user_role_id"`
	RepoCount    int               `json:"repo_count"`
	Metadata     map[string]string `json:"metadata"`
}

// =================================================================================================

type HarborProjectDeletableResp struct {
	Deletable bool   `json:"deletable"`
	Message   string `json:"message"`
}

// =================================================================================================

type HarborRepo struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	ProjectID    int64     `json:"project_id"`
	Description  string    `json:"description"`
	PullCount    int64     `json:"pull_count"`
	StarCount    int64     `json:"star_count"`
	TagsCount    int64     `json:"tags_count"`
	CreationTime time.Time `json:"creation_time"`
	UpdateTime   time.Time `json:"update_time"`
}

// =================================================================================================

type HarborTag struct {
	tagDetail
	Signature    *Target          `json:"signature"`
	ScanOverview *ImgScanOverview `json:"scan_overview,omitempty"`
}

type tagDetail struct {
	Digest        string    `json:"digest"`
	Name          string    `json:"name"`
	Size          int64     `json:"size"`
	Architecture  string    `json:"architecture"`
	OS            string    `json:"os"`
	DockerVersion string    `json:"docker_version"`
	Author        string    `json:"author"`
	Created       time.Time `json:"created"`
	Config        *cfg      `json:"config"`
}

// Target represents the json object of a target of a docker image in notary.
// The struct will be used when repository is know so it won'g contain the name of a repository.
type Target struct {
	Tag    string `json:"tag"`
	Hashes Hashes `json:"hashes"`
}

type Hashes map[string][]byte

//ImgScanOverview mapped to a record of image scan overview.
type ImgScanOverview struct {
	ID              int64               `json:"-"`
	Digest          string              `json:"image_digest"`
	Status          string              `json:"scan_status"`
	JobID           int64               `json:"job_id"`
	Sev             int                 `json:"severity"`
	CompOverviewStr string              `json:"-"`
	CompOverview    *ComponentsOverview `json:"components,omitempty"`
	DetailsKey      string              `json:"details_key"`
	CreationTime    time.Time           `json:"creation_time,omitempty"`
	UpdateTime      time.Time           `json:"update_time,omitempty"`
}

//ComponentsOverview has the total number and a list of components number of different serverity level.
type ComponentsOverview struct {
	Total   int                        `json:"total"`
	Summary []*ComponentsOverviewEntry `json:"summary"`
}

//ComponentsOverviewEntry ...
type ComponentsOverviewEntry struct {
	Sev   int `json:"severity"`
	Count int `json:"count"`
}

type cfg struct {
	Labels map[string]string `json:"labels"`
}

// =================================================================================================

// Severity represents the severity of a image/component in terms of vulnerability.
type Severity int64

// Sevxxx is the list of severity of image after scanning.
const (
	_ Severity = iota
	SevNone
	SevUnknown
	SevLow
	SevMedium
	SevHigh
)

//String is the output function for sererity variable
func (sev Severity) String() string {
	name := []string{"negligible", "unknown", "low", "medium", "high"}
	i := int64(sev)
	switch {
	case i >= 1 && i <= int64(SevHigh):
		return name[i-1]
	default:
		return "unknown"
	}
}

// HarborVulnerability is an item in the vulnerability result returned by vulnerability details API.
type HarborVulnerability struct {
	ID          string   `json:"id"`
	Severity    Severity `json:"severity"`
	Pkg         string   `json:"package"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Link        string   `json:"link"`
	Fixed       string   `json:"fixedVersion,omitempty"`
}

// =================================================================================================

type HarborStatistics struct {
	// PriPC : count of private projects
	PriPC int64 `json:"private_project_count"`
	// PriRC : count of private repositories
	PriRC int64 `json:"private_repo_count"`
	// PubPC : count of public projects
	PubPC int64 `json:"public_project_count"`
	// PubRC : count of public repositories
	PubRC int64 `json:"public_repo_count"`
	// TPC : total count of projects
	TPC int64 `json:"total_project_count"`
	// TRC : total count of repositories
	TRC int64 `json:"total_repo_count"`
}

type HarborVolumes struct {
	Storage HarborStorage `json:"storage"`
}

type HarborStorage struct {
	Total uint64 `json:"total"`
	Free  uint64 `json:"free"`
}

// =================================================================================================

// HarborReplicationPolicy defines the data model used in API level
type HarborReplicationPolicy struct {
	ID                        int64              `json:"id"`
	Name                      string             `json:"name"`
	Description               string             `json:"description"`
	Filters                   []HarborFilter     `json:"filters"`
	ReplicateDeletion         bool               `json:"replicate_deletion"`
	Trigger                   *HarborTrigger     `json:"trigger"`
	Projects                  []*HarborProject   `json:"projects"`
	Targets                   []*HarborRepTarget `json:"targets"`
	CreationTime              time.Time          `json:"creation_time"`
	UpdateTime                time.Time          `json:"update_time"`
	ReplicateExistingImageNow bool               `json:"replicate_existing_image_now"`
	ErrorJobCount             int64              `json:"error_job_count"`
}

// =================================================================================================

// HarborTarget is the model for a replication targe, i.e. destination, which wraps the endpoint URL and username/password of a remote registry.
type HarborRepTarget struct {
	ID           int64     `json:"id"`
	URL          string    `json:"endpoint"`
	Name         string    `json:"name"`
	Username     string    `json:"username"`
	Password     string    `json:"password"`
	Type         int       `json:"type"`
	Insecure     bool      `json:"insecure"`
	CreationTime time.Time `json:"creation_time"`
	UpdateTime   time.Time `json:"update_time"`
}

// =================================================================================================

// HarborFilter is the data model represents the filter defined by user.
type HarborFilter struct {
	Kind    string `json:"kind"`
	Pattern string `json:"pattern"`
}

// =================================================================================================

// HarborTrigger is replication launching approach definition
type HarborTrigger struct {
	Kind          string               `json:"kind"`           // the type of the trigger
	ScheduleParam *HarborScheduleParam `json:"schedule_param"` // optional, only used when kind is 'schedule'
}

// =================================================================================================

// ScheduleParam defines the parameters used by schedule trigger
type HarborScheduleParam struct {
	Type    string `json:"kind"`    //daily or weekly
	Weekday int8   `json:"weekday"` //Optional, only used when type is 'weekly'
	Offtime int64  `json:"offtime"` //The time offset with the UTC 00:00 in seconds
}

// Equal ...
func (s *HarborScheduleParam) Equal(param *HarborScheduleParam) bool {
	if param == nil {
		return false
	}

	return s.Type == param.Type && s.Weekday == param.Weekday && s.Offtime == param.Offtime
}

// =================================================================================================

// HarborRepJob is the model for a replication job, which is the execution unit on job service, currently it is used to transfer/remove
// a repository to/from a remote registry instance.
type HarborRepJob struct {
	ID           int64     `json:"id"`
	Status       string    `json:"status"`
	Repository   string    `json:"repository"`
	PolicyID     int64     `json:"policy_id"`
	Operation    string    `json:"operation"`
	Tags         string    `json:"-"`
	TagList      []string  `json:"tags"`
	CreationTime time.Time `json:"creation_time"`
	UpdateTime   time.Time `json:"update_time"`
}

// =================================================================================================

// HarborAccessLog holds information about logs which are used to record the actions that user take to the resourses.
type HarborAccessLog struct {
	LogID     int       `json:"log_id"`
	Username  string    `json:"username"`
	ProjectID int64     `json:"project_id"`
	RepoName  string    `json:"repo_name"`
	RepoTag   string    `json:"repo_tag"`
	GUID      string    `json:"guid"`
	Operation string    `json:"operation"`
	OpTime    time.Time `json:"op_time"`
}
