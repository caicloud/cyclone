package gogs

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/caicloud/nirvana/log"
	gogs "github.com/gogs/go-gogs-client"
	"github.com/parnurzeal/gorequest"

	c_v1alpha1 "github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/biz/scm"
	"github.com/caicloud/cyclone/pkg/util/cerr"
)

func init() {
	if err := scm.RegisterProvider(v1alpha1.Gogs, NewGogs); err != nil {
		log.Errorln(err)
	}
}

// TokenName cyclone register a new token on Gogs
const TokenName = "cyclone-auth-token"

// Gogs Gogs instance
type Gogs struct {
	scmCfg *v1alpha1.SCMSource
	client *gogs.Client
	cookie string
}

// NewGogs create new Gogs client, due to Gogs API not satisfy us,
// so some of the data should be catch from HTML, fontend should auth with
// username and password.
func NewGogs(scmCfg *v1alpha1.SCMSource) (provider scm.Provider, err error) {
	if scmCfg.User == "" || scmCfg.Password == "" {
		err = fmt.Errorf("Gogs's username or password is missing")
		return
	}

	// just create a nil gogs client
	var client = gogs.NewClient(scmCfg.Server, "")

	var accessTokens []*gogs.AccessToken
	if accessTokens, err = client.ListAccessTokens(scmCfg.User, scmCfg.Password); err != nil {
		return
	}

	var token string
	for _, t := range accessTokens {
		if t.Name == TokenName {
			token = t.Sha1
		}
	}

	if token == "" {
		var opt = gogs.CreateAccessTokenOption{Name: TokenName}
		var accessToken *gogs.AccessToken
		if accessToken, err = client.CreateAccessToken(scmCfg.User, scmCfg.Password, opt); err != nil {
			return
		}
		if accessToken == nil || accessToken.Sha1 == "" {
			err = fmt.Errorf("Gogs generate token with an error, try again or check your username and password")
			return
		}
		// got a new valid token
		token = accessToken.Sha1
	}

	var cookie string
	if cookie, err = genCookie(scmCfg); err != nil {
		return
	}

	provider = &Gogs{
		scmCfg: scmCfg,
		client: gogs.NewClient(scmCfg.Server, token),
		cookie: cookie,
	}
	return
}

// GetToken get Gogs's token
func (g *Gogs) GetToken() (token string, err error) {
	token = g.scmCfg.Token
	return
}

// ListRepos get all of users' repo and its orgs' repos
func (g *Gogs) ListRepos() (repos []scm.Repository, err error) {
	var repositories []*gogs.Repository
	if repositories, err = g.client.ListMyRepos(); err != nil {
		return
	}

	for _, r := range repositories {
		repos = append(repos, scm.Repository{Name: r.FullName, URL: r.CloneURL})
	}

	var organizations []*gogs.Organization
	if organizations, err = g.client.ListMyOrgs(); err != nil {
		return
	}

	for _, org := range organizations {
		var orgRepositories []*gogs.Repository
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
func (g *Gogs) ListBranches(repoName string) (branches []string, err error) {
	var owner, repo string
	owner, repo = scm.ParseRepo(repoName)
	if len(owner) == 0 || len(repo) == 0 {
		err = fmt.Errorf("invalid repo %s, must in format of '{owner}/{repo}'", repoName)
		return
	}
	var gogsBranches []*gogs.Branch
	if gogsBranches, err = g.client.ListRepoBranches(owner, repo); err != nil {
		return
	}
	for _, b := range gogsBranches {
		branches = append(branches, b.Name)
	}
	return
}

// ListTags list all of repo's tags
func (g *Gogs) ListTags(repoName string) (tags []string, err error) {
	var owner, repo string
	owner, repo = scm.ParseRepo(repoName)
	if len(owner) == 0 || len(repo) == 0 {
		err = fmt.Errorf("invalid repo %s, must in format of '{owner}/{repo}'", repoName)
		return
	}

	var request = gorequest.New()
	var response gorequest.Response
	var errs []error

	var body string
	var url = fmt.Sprintf("%s/%s/%s", strings.TrimSuffix(g.scmCfg.Server, "/"), owner, repo)
	if response, body, errs = request.Get(url).Set("Cookie", g.cookie).End(); len(errs) != 0 {
		err = errs[len(errs)-1]
		return
	}
	if response == nil {
		err = fmt.Errorf("request Gogs server with fatal error: %s", url)
		return
	}
	if response.StatusCode != 200 {
		err = fmt.Errorf("request Gogs server got code: %d, url: %s", response.StatusCode, url)
		return
	}

	var bodyReader = strings.NewReader(body)

	var document *goquery.Document

	if document, err = goquery.NewDocumentFromReader(bodyReader); err != nil {
		return
	}

	document.Find("#tag-list .item").Each(func(_ int, selection *goquery.Selection) {
		tags = append(tags, strings.TrimSpace(selection.Text()))
	})

	return
}

// ListPullRequests list all of the pr for specified repo
func (g *Gogs) ListPullRequests(repo, state string) (pr []scm.PullRequest, err error) {
	err = cerr.ErrorNotImplemented.Error("get pull request")
	return
}

// ListDockerfiles list dockerfile in specified repo
func (g *Gogs) ListDockerfiles(repo string) (dockerfiles []string, err error) {
	err = cerr.ErrorNotImplemented.Error("list dockerfiles")
	return
}

// CreateStatus create status in specified repo
func (g *Gogs) CreateStatus(status c_v1alpha1.StatusPhase, targetURL, repoURL, commitSHA string) (err error) {
	err = cerr.ErrorNotImplemented.Error("create status")
	return
}

// GetPullRequestSHA get pr's sha
func (g *Gogs) GetPullRequestSHA(repoURL string, number int) (prHash string, err error) {
	err = cerr.ErrorNotImplemented.Error("get pull request SHA")
	return
}

// CheckToken check the token is valid or not
func (g *Gogs) CheckToken() (err error) {
	var repos []scm.Repository
	if repos, err = g.ListRepos(); err != nil {
		return
	}

	if len(repos) > 0 {
		if _, err = g.ListTags(repos[0].Name); err != nil {
			return
		}
	}

	return
}

// CreateWebhook crate web hook for specified repo
func (g *Gogs) CreateWebhook(repoName string, webhook *scm.Webhook) (err error) {
	var owner, repo string
	owner, repo = scm.ParseRepo(repoName)
	if len(owner) == 0 || len(repo) == 0 {
		err = fmt.Errorf("invalid repo %s, must in format of '{owner}/{repo}'", repoName)
		return
	}
	var option = gogs.CreateHookOption{
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
func (g *Gogs) DeleteWebhook(repoName string, webhookURL string) (err error) {
	var owner, repo string
	owner, repo = scm.ParseRepo(repoName)
	if len(owner) == 0 || len(repo) == 0 {
		err = fmt.Errorf("invalid repo %s, must in format of '{owner}/{repo}'", repoName)
		return
	}
	var hooks []*gogs.Hook
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
