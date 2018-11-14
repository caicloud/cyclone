package regex

import (
	"regexp"
	"strconv"
)

const (
	repoNameRegexp = `^http[s]?://(?:git[\w]+\.com|[\d]+\.[\d]+\.[\d]+\.[\d]+:[\d]+|localhost:[\d]+)/([\S]*)\.git$`

	GitLabMergeRequestRegexp       = `^refs/merge-requests/([\d]+)/head`
	GitLabMergeRequestStrictRegexp = `^refs/merge-requests/([\d]+)/head$`

	GitHubPullRequestRegexp = `^refs/pull/([\d]+)/merge$`
)

// GetGitlabMRID retrieve the ID of GitLab merge-requests,
// if strict is false, we will not require that
// the input ref is end with `head`
//
// input   ref      =  `refs/merge-requests/1/head`
// output  id       =  1
func GetGitlabMRID(ref string, strict bool) (int, bool) {
	p := GitLabMergeRequestRegexp
	if strict {
		p = GitLabMergeRequestStrictRegexp
	}

	r := regexp.MustCompile(p)
	results := r.FindStringSubmatch(ref)
	if len(results) < 2 {
		return 0, false
	}

	id, _ := strconv.Atoi(results[1])
	return id, true
}

// GetGithubPRID retrieve the ID of GitHub pull-requests,
//
// input   ref      =  `refs/pull/1/merge`
// output  id       =  1
func GetGithubPRID(ref string) (int, bool) {
	r := regexp.MustCompile(GitHubPullRequestRegexp)
	results := r.FindStringSubmatch(ref)
	if len(results) < 2 {
		return 0, false
	}

	id, _ := strconv.Atoi(results[1])
	return id, true
}
