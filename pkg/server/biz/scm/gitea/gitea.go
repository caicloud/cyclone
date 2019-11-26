package gitea

import (
	"fmt"

	gitea "code.gitea.io/sdk/gitea"

	c_v1alpha1 "github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/biz/scm"
	"github.com/caicloud/cyclone/pkg/util/cerr"
	"github.com/caicloud/nirvana/log"
)

func init() {
	if err := scm.RegisterProvider(v1alpha1.Gitea, NewGitea); err != nil {
		log.Errorln(err)
	}
}

// TokenName cyclone register a new token on Gitea
const TokenName = "cyclone-auth-token"

// Gitea Gitea instance
type Gitea struct {
	scmCfg *v1alpha1.SCMSource
	client *gitea.Client
}

const checkPassword = ", please try again later or check your username and password"

// NewGitea create new Gitea client
func NewGitea(scmCfg *v1alpha1.SCMSource) (provider scm.Provider, err error) {
	var client *gitea.Client // just create a nil gogs client
	if scmCfg.Token != "" {
		client = gitea.NewClient(scmCfg.Server, scmCfg.Token)
	} else {
		if scmCfg.User == "" || scmCfg.Password == "" {
			err = fmt.Errorf("fail to new Gitea client%s", checkPassword)
			return
		}
		var token string
		if token, err = genTokenByBasicAuth(scmCfg.Server, scmCfg.User, scmCfg.Password); err != nil {
			return
		}
		client = gitea.NewClient(scmCfg.Server, token)
	}

	provider = &Gitea{
		scmCfg: scmCfg,
		client: client,
	}
	return
}

// GetToken get Gitea's token
func (g *Gitea) GetToken() (token string, err error) {
	token, err = genTokenByBasicAuth(g.scmCfg.Server, g.scmCfg.User, g.scmCfg.Password)
	return
}

// genTokenByBasicAuth generate token by username and password
func genTokenByBasicAuth(server, user, password string) (token string, err error) {
	if len(server) == 0 {
		err = fmt.Errorf("Gitea server is missing")
		return
	}
	if len(user) == 0 || len(password) == 0 {
		err = fmt.Errorf("Gitea username or password is missing")
		return
	}
	var client = gitea.NewClient(server, "")
	var accessTokens []*gitea.AccessToken
	if accessTokens, err = client.ListAccessTokens(user, password); err != nil {
		err = fmt.Errorf("Gitea got an error: %v%s", err, checkPassword)
		return
	}

	for _, t := range accessTokens {
		if t.Name == TokenName {
			token = t.Token
		}
	}

	if token == "" {
		var opt = gitea.CreateAccessTokenOption{Name: TokenName}
		var accessToken *gitea.AccessToken
		if accessToken, err = client.CreateAccessToken(user, password, opt); err != nil {
			err = fmt.Errorf("Gitea got an error: %v%s", err, checkPassword)
			return
		}
		if accessToken == nil || accessToken.Token == "" {
			err = fmt.Errorf("Gitea generate token with an error%s", checkPassword)
			return
		}
		token = accessToken.Token // got a new valid token
	}
	return
}

// ListRepos get all of users' repo and its orgs' repos
func (g *Gitea) ListRepos() (repos []scm.Repository, err error) {
	var repositories []*gitea.Repository
	if repositories, err = g.client.ListMyRepos(); err != nil {
		return
	}

	for _, r := range repositories {
		repos = append(repos, scm.Repository{Name: r.FullName, URL: r.CloneURL})
	}

	var organizations []*gitea.Organization
	if organizations, err = g.client.ListMyOrgs(); err != nil {
		return
	}

	for _, org := range organizations {
		var orgRepositories []*gitea.Repository
		if orgRepositories, err = g.client.ListOrgRepos(org.UserName); err != nil {
			return
		}
		for _, r := range orgRepositories {
			repos = append(repos, scm.Repository{Name: r.FullName, URL: r.CloneURL})
		}
	}
	return
}

// ListBranches list all of branch for specified repo
func (g *Gitea) ListBranches(repoName string) (branches []string, err error) {
	var owner, repo string
	owner, repo = scm.ParseRepo(repoName)
	if len(owner) == 0 || len(repo) == 0 {
		err = fmt.Errorf("invalid repo %s, must in format of '{owner}/{repo}'", repoName)
		return
	}
	var giteaBranches []*gitea.Branch
	if giteaBranches, err = g.client.ListRepoBranches(owner, repo); err != nil {
		return
	}
	for _, b := range giteaBranches {
		branches = append(branches, b.Name)
	}
	return
}

// ListTags list all of repo's tags
func (g *Gitea) ListTags(repoName string) (tags []string, err error) {
	var owner, repo string
	owner, repo = scm.ParseRepo(repoName)
	if len(owner) == 0 || len(repo) == 0 {
		err = fmt.Errorf("invalid repo %s, must in format of '{owner}/{repo}'", repoName)
		return
	}
	var giteaTags []*gitea.Tag
	if giteaTags, err = g.client.ListRepoTags(owner, repo); err != nil {
		return
	}
	for _, b := range giteaTags {
		tags = append(tags, b.Name)
	}
	return
}

// ListPullRequests list all of the pr for specified repo
func (g *Gitea) ListPullRequests(repo, state string) (pr []scm.PullRequest, err error) {
	err = cerr.ErrorNotImplemented.Error("get pull request")
	return
}

// ListDockerfiles list dockerfile in specified repo
func (g *Gitea) ListDockerfiles(repo string) (dockerfiles []string, err error) {
	err = cerr.ErrorNotImplemented.Error("list dockerfiles")
	return
}

// CreateStatus create status in specified repo
func (g *Gitea) CreateStatus(status c_v1alpha1.StatusPhase, targetURL, repoURL, commitSHA string) (err error) {
	err = cerr.ErrorNotImplemented.Error("create status")
	return
}

// GetPullRequestSHA get pr's sha
func (g *Gitea) GetPullRequestSHA(repoURL string, number int) (prHash string, err error) {
	err = cerr.ErrorNotImplemented.Error("get pull request SHA")
	return
}

// CheckToken check the token is valid or not
func (g *Gitea) CheckToken() (err error) {
	_, err = g.ListRepos()
	return
}

// CreateWebhook crate web hook for specified repo
func (g *Gitea) CreateWebhook(repoName string, webhook *scm.Webhook) (err error) {
	var owner, repo string
	owner, repo = scm.ParseRepo(repoName)
	if len(owner) == 0 || len(repo) == 0 {
		err = fmt.Errorf("invalid repo %s, must in format of '{owner}/{repo}'", repoName)
		return
	}
	var option = gitea.CreateHookOption{
		Type: "gogs",
		Config: map[string]string{
			"url":          webhook.URL,
			"content_type": "json",
		},
		Events: convertToGogsEvents(webhook.Events),
		Active: true,
	}
	if _, err = g.client.CreateRepoHook(owner, repo, option); err != nil {
		return
	}
	return
}

// Event Gogs event type
type Event string

const (
	// EventPR Gogs event pull request
	EventPR Event = "pull_request"
	// EventPush Gogs event push
	EventPush Event = "push"
	// EventCreate Gogs event create
	EventCreate Event = "create"
)

// convertToGogsEvents converts the defined event types to Gogs event types.
func convertToGogsEvents(events []scm.EventType) (ge []string) {
	for _, e := range events {
		switch e {
		case scm.PullRequestEventType:
			ge = append(ge, string(EventPR))
		case scm.PushEventType:
			ge = append(ge, string(EventPush))
		case scm.TagReleaseEventType:
			ge = append(ge, string(EventCreate))
		default:
			log.Errorf("The event type %s is not supported, will be ignored", e)
		}
	}
	return
}

// DeleteWebhook delete specified repo web hook
func (g *Gitea) DeleteWebhook(repoName string, webhookURL string) (err error) {
	var owner, repo string
	owner, repo = scm.ParseRepo(repoName)
	if len(owner) == 0 || len(repo) == 0 {
		err = fmt.Errorf("invalid repo %s, must in format of '{owner}/{repo}'", repoName)
		return
	}
	var hooks []*gitea.Hook
	if hooks, err = g.client.ListRepoHooks(owner, repo); err != nil {
		return
	}
	var hooksIDDeleting []int64
	for _, h := range hooks {
		if h.Config["url"] == webhookURL {
			hooksIDDeleting = append(hooksIDDeleting, h.ID)
		}
	}
	for _, h := range hooksIDDeleting {
		if err = g.client.DeleteRepoHook(owner, repo, h); err != nil {
			return
		}
	}
	return
}
