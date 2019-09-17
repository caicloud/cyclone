package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/caicloud/nirvana/log"

	c_v1alpha1 "github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/biz/scm"
	"github.com/caicloud/cyclone/pkg/util/cerr"
)

// SupportAccessTokenVersion represents the lowest version
// that support personal access token.
const SupportAccessTokenVersion = "5.5.0"

// SupportPrModifiedEvent represents the lowest version
// that support event key named pr:modified.
const SupportPrModifiedEvent = "5.10.0"

// Property represents BitBucket Server property.
type Property struct {
	Version string `json:"version"`
}

// BitbucketServer represents the SCM provider of BitBucket Server.
type BitbucketServer struct {
	scmCfg   *v1alpha1.SCMSource
	v1Client *V1Client
}

// NewBitbucketServer new a SCM provider of BitBucket Server.
func NewBitbucketServer(scmCfg *v1alpha1.SCMSource, v1Client *V1Client) *BitbucketServer {
	return &BitbucketServer{
		scmCfg:   scmCfg,
		v1Client: v1Client,
	}
}

// GetToken gets the token by the username and password of SCM config.
func (b *BitbucketServer) GetToken() (string, error) {
	if b.scmCfg.AuthType == v1alpha1.AuthTypePassword {
		version, err := GetBitbucketVersion(b.v1Client.client, b.v1Client.baseURL)
		if err != nil {
			return "", err
		}
		isSupportToken, err := IsHigherVersion(version, SupportAccessTokenVersion)
		if err != nil {
			return "", err
		}
		if isSupportToken {
			opt := PermissionReq{
				Name:        "continuous-integration/cyclone",
				Permissions: []string{ProjectRead, RepoAdmin},
			}
			token, resp, err := b.v1Client.Authorizations.CreateAccessToken(context.Background(), b.scmCfg.User, opt)
			if err != nil {
				return "", convertBitBucketError(err, resp)
			}
			return token, nil
		}
		return b.scmCfg.Password, nil
	}

	return b.scmCfg.Token, nil
}

// CheckToken checks whether the token has the authority of repo by trying ListRepos with the token.
func (b *BitbucketServer) CheckToken() error {
	repos, err := b.listReposInner(false)
	if err != nil {
		return err
	}
	if len(repos) == 0 {
		return cerr.ErrorExternalAuthorizationFailed.Error(fmt.Errorf("No repositories found"))
	}
	return nil
}

// ListRepos lists the repos by the SCM config.
func (b *BitbucketServer) ListRepos() ([]scm.Repository, error) {
	return b.listReposInner(true)
}

// listReposInner lists the projects by the SCM config,
// list all projects while the parameter 'listAll' is true,
// otherwise, list projects by default 'ListPerPageOpt' number.
func (b *BitbucketServer) listReposInner(listAll bool) ([]scm.Repository, error) {
	opt := ListOpts{}
	var allRepos []Repository
	for {
		repos, resp, err := b.v1Client.Repositories.ListRepositories(context.Background(), "", &opt)
		if err != nil {
			return nil, convertBitBucketError(err, resp)
		}
		allRepos = append(allRepos, repos.Values...)
		if repos.NextPage == nil || !listAll {
			break
		}
		opt.Start = repos.NextPage
	}

	scmRepos := make([]scm.Repository, len(allRepos))
	for i, repo := range allRepos {
		scmRepos[i].Name = fmt.Sprintf("%s/%s", repo.Project.Key, repo.Slug)
		for _, clone := range repo.Links.Clone {
			if clone.Name == "http" || clone.Name == "https" {
				scmRepos[i].URL = clone.Href
				break
			}
		}
	}

	return scmRepos, nil
}

// ListBranches lists the branches for specified repo.
func (b *BitbucketServer) ListBranches(repo string) ([]string, error) {
	var projectKey string
	var err error
	if projectKey, repo, err = parseRepo(b.scmCfg, repo); err != nil {
		log.Error(err)
		return nil, err
	}
	opt := ListOpts{}
	var allBranches []Branch
	for {
		branches, resp, err := b.v1Client.Repositories.ListBranches(context.Background(), projectKey, repo, &opt)
		if err != nil {
			log.Errorf("Fail to list branches for %s as %v", repo, err)
			return nil, convertBitBucketError(err, resp)
		}

		allBranches = append(allBranches, branches.Values...)
		if branches.NextPage == nil {
			break
		}
		opt.Start = branches.NextPage
	}

	branchNames := make([]string, len(allBranches))
	for i, branch := range allBranches {
		branchNames[i] = branch.DisplayID
	}

	return branchNames, nil
}

// ListTags lists the tags for specified repo.
func (b *BitbucketServer) ListTags(repo string) ([]string, error) {
	var projectKey string
	var err error
	if projectKey, repo, err = parseRepo(b.scmCfg, repo); err != nil {
		log.Error(err)
		return nil, err
	}
	var allTags []Tag
	opt := ListOpts{}
	for {
		tags, resp, err := b.v1Client.Repositories.ListTags(context.Background(), projectKey, repo, &opt)
		if err != nil {
			log.Errorf("Fail to list tags for %s as %v", repo, err)
			return nil, convertBitBucketError(err, resp)
		}

		allTags = append(allTags, tags.Values...)
		if tags.NextPage == nil {
			break
		}
		opt.Start = tags.NextPage
	}

	tagNames := make([]string, len(allTags))
	for i, tag := range allTags {
		tagNames[i] = tag.DisplayID
	}

	return tagNames, nil
}

// ListPullRequests lists the pull requests for specified repo.
func (b *BitbucketServer) ListPullRequests(repo, state string) ([]scm.PullRequest, error) {
	//  Bitbucket pr state: OPEN, DECLINED, MERGED, ALL
	var s string
	switch state {
	case "open":
		s = "OPEN"
	case "all":
		s = "ALL"
	case "OPEN", "DECLINED", "MERGED", "ALL":
		s = state
	default:
		return nil, cerr.ErrorUnsupported.Error("Bitbucket pull request state", state)
	}

	var projectKey string
	var err error
	if projectKey, repo, err = parseRepo(b.scmCfg, repo); err != nil {
		log.Error(err)
		return nil, err
	}

	var allPRs []scm.PullRequest
	opt := PullRequestListOpts{
		State: s,
	}
	for {
		prs, resp, err := b.v1Client.Repositories.ListPullRequests(context.Background(), projectKey, repo, &opt)
		if err != nil {
			log.Errorf("Fail to list pull requests for %s as %v", repo, err)
			return nil, convertBitBucketError(err, resp)
		}

		allPRs = append(allPRs, prs.Values...)
		if prs.NextPage == nil {
			break
		}
		opt.Start = prs.NextPage
	}

	return allPRs, nil
}

// ListDockerfiles lists the Dockerfiles for specified repo.
func (b *BitbucketServer) ListDockerfiles(repo string) ([]string, error) {
	var projectKey string
	var err error
	if projectKey, repo, err = parseRepo(b.scmCfg, repo); err != nil {
		log.Error(err)
		return nil, err
	}
	var allFiles []string
	opt := ListOpts{}
	for {
		files, resp, err := b.v1Client.Repositories.ListFiles(context.Background(), projectKey, repo, &opt)
		if err != nil {
			log.Errorf("Fail to list files for %s as %s", repo, err)
			return nil, convertBitBucketError(err, resp)
		}

		allFiles = append(allFiles, files.Values...)
		if files.NextPage == nil {
			break
		}
		opt.Start = files.NextPage
	}

	var dockerfiles []string
	for _, file := range allFiles {
		if scm.IsDockerfile(file) {
			dockerfiles = append(dockerfiles, file)
		}
	}

	return dockerfiles, nil
}

// CreateStatus generate a new status for repository.
func (b *BitbucketServer) CreateStatus(status c_v1alpha1.StatusPhase, targetURL, repoURL, commitSha string) error {
	// BitBucket Server:  SUCCESSFUL, FAILED and INPROGRESS.
	state := ""
	description := ""

	switch status {
	case c_v1alpha1.StatusRunning:
		state = "INPROGRESS"
		description = "Cyclone CI is in progress."
	case c_v1alpha1.StatusSucceeded:
		state = "SUCCESSFUL"
		description = "Cyclone CI passed."
	case c_v1alpha1.StatusFailed:
		state = "FAILED"
		description = "Cyclone CI failed."
	case c_v1alpha1.StatusCancelled:
		state = "FAILED"
		description = "Cyclone CI failed."
	default:
		err := fmt.Errorf("not supported state:%s", status)
		log.Error(err)
		return err
	}

	label := "continuous-integration/cyclone"
	opt := &StatusReq{
		State:       state,
		Key:         label,
		Name:        label,
		URL:         targetURL,
		Description: description,
	}
	resp, err := b.v1Client.Repositories.CreateStatus(context.Background(), commitSha, opt)
	return convertBitBucketError(err, resp)
}

// GetPullRequestSHA gets latest commit SHA of pull request.
func (b *BitbucketServer) GetPullRequestSHA(repoURL string, number int) (string, error) {
	projectKey, name := scm.ParseRepo(repoURL)
	pr, resp, err := b.v1Client.PullRequests.GetPullRequest(context.Background(), projectKey, name, number)
	if err != nil {
		log.Error(err)
		return "", convertBitBucketError(err, resp)
	}

	return pr.FromRef.LatestCommit, nil
}

// CreateWebhook creates webhook for specified repo.
func (b *BitbucketServer) CreateWebhook(repo string, webhook *scm.Webhook) error {
	if webhook == nil || len(webhook.URL) == 0 || len(webhook.Events) == 0 {
		return fmt.Errorf("The webhook %v is not correct", webhook)
	}

	_, err := b.GetWebhook(repo, webhook.URL)
	if err != nil {
		if !cerr.ErrorContentNotFound.Derived(err) {
			return err
		}

		webhookReq := Webhook{
			URL:    webhook.URL,
			Events: make([]string, 0),
			Active: true,
			Name:   "continuous-integration/cyclone",
		}

		for _, e := range webhook.Events {
			switch e {
			case scm.PullRequestEventType:
				version, err := GetBitbucketVersion(b.v1Client.client, b.v1Client.baseURL)
				if err != nil {
					return err
				}
				isSupportPrModified, err := IsHigherVersion(version, SupportPrModifiedEvent)
				if err != nil {
					return err
				}
				if isSupportPrModified {
					webhookReq.Events = append(webhookReq.Events, PrOpened, PrModified)
				} else {
					webhookReq.Events = append(webhookReq.Events, PrOpened)
				}
			case scm.PullRequestCommentEventType:
				webhookReq.Events = append(webhookReq.Events, PrCommentAdded)
			case scm.PushEventType, scm.TagReleaseEventType:
				flag := false
				for _, event := range webhookReq.Events {
					if event == RefsChanged {
						flag = true
						break
					}
				}
				if !flag {
					webhookReq.Events = append(webhookReq.Events, RefsChanged)
				}
			default:
				log.Errorf("The event type %s is not supported, will be ignored", e)
			}
		}

		project, name := scm.ParseRepo(repo)
		_, resp, err := b.v1Client.Repositories.CreateWebhook(context.Background(), project, name, webhookReq)
		if err != nil {
			log.Errorf("Create Webhook error: %v", err)
		}
		return convertBitBucketError(err, resp)
	}

	log.Warningf("Webhook already existed: %+v", webhook)
	return nil
}

// DeleteWebhook deletes webhook from specified repo.
func (b *BitbucketServer) DeleteWebhook(repo string, webhookURL string) error {
	project, name := scm.ParseRepo(repo)

	webhook, err := b.GetWebhook(repo, webhookURL)
	if err != nil {
		return err
	}
	if resp, err := b.v1Client.Repositories.DeleteWebhook(context.Background(), project, name, webhook.ID); err != nil {
		log.Errorf("delete project hook %s for %s/%s error: %v", webhook.ID, project, name, err)
		return convertBitBucketError(err, resp)
	}
	return nil
}

// GetWebhook gets webhook from specified repo.
func (b *BitbucketServer) GetWebhook(repo string, webhookURL string) (*Webhook, error) {
	project, name := scm.ParseRepo(repo)

	opt := ListOpts{}
	var allWebhooks []Webhook
	for {
		hooks, resp, err := b.v1Client.Repositories.ListWebhook(context.Background(), project, name)
		if err != nil {
			return nil, convertBitBucketError(err, resp)
		}

		allWebhooks = append(allWebhooks, hooks.Values...)
		if hooks.NextPage == nil {
			break
		}
		opt.Start = hooks.NextPage
	}

	for _, hook := range allWebhooks {
		if strings.HasPrefix(hook.URL, webhookURL) {
			return &hook, nil
		}
	}
	return nil, cerr.ErrorContentNotFound.Error(fmt.Sprintf("webhook url %s", webhookURL))
}

func parseRepo(scmCfg *v1alpha1.SCMSource, repo string) (string, string, error) {
	projectKey := fmt.Sprintf("~%s", scmCfg.User)
	if strings.Contains(repo, "/") {
		parts := strings.Split(repo, "/")
		if len(parts) != 2 {
			err := fmt.Errorf("invalid repo %s, must in format of '{projectKey}/{repo}'", repo)
			return projectKey, repo, err
		}
		projectKey, repo = parts[0], parts[1]
	}
	return projectKey, repo, nil
}

// GetBitbucketVersion returns the version of the BitBucket Server
func GetBitbucketVersion(client *http.Client, base *url.URL) (string, error) {
	resp, err := client.Get(fmt.Sprintf("%s%s", base.String(), "rest/api/1.0/application-properties"))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var property Property
	if resp.StatusCode/100 == 2 {
		err = json.NewDecoder(resp.Body).Decode(&property)
		if err != nil {
			return "", err
		}
		return property.Version, nil
	}
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	return "", fmt.Errorf("status: %v, Body: %s", resp.Status, string(bodyBytes))
}

// IsHigherVersion compare the version with refVersion.
// If the version is lower than the refVersion, it will return false.
func IsHigherVersion(version string, refVersion string) (bool, error) {
	parts := strings.Split(version, ".")
	refParts := strings.Split(refVersion, ".")
	if len(refParts) != 3 {
		return false, fmt.Errorf("invalid refVersion: %s, it must be x.x.x", refVersion)
	}
	if len(parts) > 3 || len(parts) < 2 {
		return false, fmt.Errorf("invalid version: %s", version)
	}
	if len(parts) == 2 {
		parts = append(parts, "0")
	}
	for i, part := range parts {
		partNumber, err := strconv.Atoi(part)
		if err != nil {
			return false, fmt.Errorf("invalid version: %s, error: %v", version, err)
		}
		refPartNumber, err := strconv.Atoi(refParts[i])
		if err != nil {
			return false, fmt.Errorf("invalid refVersion: %s, error: %v", refVersion, err)
		}
		if refPartNumber > partNumber {
			return false, nil
		} else if refPartNumber < partNumber {
			return true, nil
		}
	}
	return true, nil
}

func convertBitBucketError(err error, resp *http.Response) error {
	if err == nil {
		return nil
	}

	if resp != nil && resp.StatusCode == http.StatusInternalServerError {
		return cerr.ErrorExternalSystemError.Error("BitBucket", err)
	}

	if resp != nil && resp.StatusCode == http.StatusUnauthorized {
		return cerr.ErrorExternalAuthorizationFailed.Error(err)
	}

	if resp != nil && resp.StatusCode == http.StatusForbidden {
		return cerr.ErrorExternalAuthenticationFailed.Error(err)
	}

	return cerr.AutoAnalyse(err)
}
