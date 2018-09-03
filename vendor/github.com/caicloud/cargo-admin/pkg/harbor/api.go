package harbor

import (
	"fmt"
	"net/url"
)

const (
	APISearch              = "/api/search"
	APIStatistic           = "/api/statistics"
	APIProjects            = "/api/projects"
	APIProject             = "/api/projects/%d"
	APIProjectDeletable    = "/api/projects/%d/_deletable"
	APIRepositories        = "/api/repositories"
	APIRepository          = "/api/repositories/%s/%s"
	APITags                = "/api/repositories/%s/%s/tags"
	APITag                 = "/api/repositories/%s/%s/tags/%s"
	APIVulnerability       = "/api/repositories/%s/%s/tags/%s/vulnerability/details"
	APILogs                = "/api/logs"
	APIProjectLogs         = "/api/projects/%d/logs"
	APIReplicationPolicies = "/api/policies/replication"
	APIReplicationPolicy   = "/api/policies/replication/%d"
	APIReplications        = "/api/replications"
	APITargets             = "/api/targets"
	APITarget              = "/api/targets/%d"
	APIVolumes             = "/api/systeminfo/volumes"
	APIJobsRep             = "/api/jobs/replication"
)

func ProjectsPath(page, pageSize int, name, public string) string {
	return fmt.Sprintf("%s?page=%d&page_size=%d&name=%s&public=%s", APIProjects, page, pageSize, url.QueryEscape(name), public)
}

func ProjectPath(pid int64) string {
	return fmt.Sprintf(APIProject, pid)
}

func ProjectExist(name string) string {
	return fmt.Sprintf("%s?project_name=%s", APIProjects, url.QueryEscape(name))
}

func ProjectDeletablePath(pid int64) string {
	return fmt.Sprintf(APIProjectDeletable, pid)
}

func RepoPath(project, repo string) string {
	return fmt.Sprintf(APIRepository, project, repo)
}

func TagsPath(project, repo string) string {
	return fmt.Sprintf(APITags, project, repo)
}

func TagPath(project, repo, tag string) string {
	return fmt.Sprintf(APITag, project, repo, tag)
}

func TagVulnerabilityPath(project, repo, tag string) string {
	return fmt.Sprintf(APIVulnerability, project, repo, tag)
}

func LoginUrl(host, user, pwd string) string {
	return fmt.Sprintf("%s/login?principal=%s&password=%s", host, user, pwd)
}

func LogsPath(startTime, endTime int64, op string) string {
	return fmt.Sprintf("%s?begin_timestamp=%d&end_timestamp=%d&operation=%s", APILogs, startTime, endTime, op)
}

func ProjectLogsPath(pid, startTime, endTime int64, op string) string {
	return fmt.Sprintf(APIProjectLogs+"?begin_timestamp=%d&end_timestamp=%d&operation=%s", pid, startTime, endTime, op)
}

func ReposPath(pid int64, query string, page, pageSize int) string {
	return fmt.Sprintf("%s?project_id=%d&q=%s&page=%d&page_size=%d", APIRepositories, pid, url.QueryEscape(query), page, pageSize)
}

func ReplicationPoliciesPath() string {
	return fmt.Sprintf("%s", APIReplicationPolicies)
}

func ProjectReplicationPoliciesPath(pid, page, pageSize int64) string {
	return fmt.Sprintf("%s?project_id=%d&page=%d&page_size=%d", APIReplicationPolicies, pid, page, pageSize)
}

func ReplicationPoliciesPathWithProjectId(pid int64) string {
	return fmt.Sprintf("%s?project_id=%d", APIReplicationPolicies, pid)
}

func ReplicationPolicyPath(rpid int64) string {
	return fmt.Sprintf(APIReplicationPolicy, rpid)
}

func ReplicationsPath() string {
	return fmt.Sprintf("%s", APIReplications)
}

func TargetsPath() string {
	return fmt.Sprintf("%s", APITargets)
}

func TargetPath(tid int64) string {
	return fmt.Sprintf(APITarget, tid)
}

func VolumesPath() string {
	return fmt.Sprintf("%s", APIVolumes)
}

func APIJobsRepPath() string {
	return fmt.Sprintf("%s", APIJobsRep)
}
func APIJobsRepPathWithQuery(policyId int64, repository string, startTime int64, endTime int64) string {
	return fmt.Sprintf("%s?policy_id=%d&repository=%s&start_time=%d&end_time=%d", APIJobsRep, policyId, repository, startTime, endTime)
}
