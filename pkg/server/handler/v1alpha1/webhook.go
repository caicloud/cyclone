package v1alpha1

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/service"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/biz/scm"
	"github.com/caicloud/cyclone/pkg/server/biz/scm/github"
	"github.com/caicloud/cyclone/pkg/server/biz/scm/gitlab"
	"github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/server/handler"
	cerr "github.com/caicloud/cyclone/pkg/util/cerr"
)

type webhookResponse struct {
	Message string `json:"message,omitempty"`
}

const (
	succeededMsg = "Successfully triggered"

	IgnoredMsg = "Is ignored"
)

const (
	// branchRefTemplate represents reference template for branches.
	branchRefTemplate = "refs/heads/%s"

	// tagRefTemplate represents reference template for tags.
	tagRefTemplate = "refs/tags/%s"

	// githubPullRefTemplate represents reference template for Github pull request.
	githubPullRefTemplate = "refs/pull/%d/merge"

	// gitlabMergeRefTemplate represents reference template for Gitlab merge request and merge target branch
	gitlabMergeRefTemplate = "refs/merge-requests/%d/head:%s"

	// gitlabEventTypeHeader represents the Gitlab header key used to pass the event type.
	gitlabEventTypeHeader = "X-Gitlab-Event"

	// githubEventTypeHeader represents the Github header key used to pass the event type.
	githubEventTypeHeader = "X-Github-Event"
)

// githubRepoNameRegexp represents the regexp of github status url.
var githubStatusURLRegexp *regexp.Regexp

func init() {
	var statusURLRegexp = `^https://api.github.com/repos/[\S]+/[\S]+/statuses/([\w]+)$`
	githubStatusURLRegexp = regexp.MustCompile(statusURLRegexp)
}

// HandleWebhook handles webhooks from integrated systems.
func HandleWebhook(ctx context.Context, tenant, integration string) (webhookResponse, error) {
	response := webhookResponse{}
	request := service.HTTPContextFrom(ctx).Request()

	repos, err := getReposFromSecret(tenant, integration)
	if err != nil {
		log.Error(err)
		return response, nil
	}

	if len(repos) == 0 {
		return response, nil
	}

	if request.Header.Get(githubEventTypeHeader) != "" {
		data, err := github.ParseEvent(request)
		if err != nil {
			return response, err
		}

		if wfts, ok := repos[data.Repo]; ok {
			for _, wft := range wfts {
				// handleGithubEvent(wft, params)
				log.Infof("Trigger workflow trigger %s\n", wft)
				if err = createWorkflowRun(tenant, wft, data); err != nil {
					log.Error(err)
					return response, nil
				}
			}
		}
	}

	if request.Header.Get(gitlabEventTypeHeader) != "" {
		data, err := gitlab.ParseEvent(request)
		if err != nil {
			return response, err
		}

		if wfts, ok := repos[data.Repo]; ok {
			for _, wft := range wfts {
				// handleGithubEvent(wft, params)
				log.Infof("Trigger workflow trigger %s\n", wft)
				if err = createWorkflowRun(tenant, wft, data); err != nil {
					log.Error(err)
					return response, nil
				}
			}
		}
	}

	return response, nil
}

// return repo, params, error
// func parseGithubEvent(request *http.Request) (string, error) {
// 	payload, err := ioutil.ReadAll(request.Body)
// 	if err != nil {
// 		return "", fmt.Errorf("Fail to read the request body")
// 	}

// 	event, err := github.ParseWebHook(github.WebHookType(request), payload)
// 	if err != nil {
// 		return "", err
// 	}

// 	switch event := event.(type) {
// 	case *github.ReleaseEvent:
// 		log.Infof("release tag: %s\n", *event.Release.TagName)
// 		return *event.Repo.FullName, nil
// 	case *github.PullRequestEvent:
// 		log.Infof("pull request: %s\n", *event.PullRequest.Number)
// 	}

// 	switch request.Header.Get(githubEventTypeHeader) {
// 	case "release":
// 		return "release", nil
// 	}

// 	// log.Infof("Github webhook event: %v", event)
// 	return "", nil
// }

func handleGithubEvent(wfts v1alpha1.WorkflowTrigger, event string) (webhookResponse, error) {
	// Trigger workflows for these workflowtriggers according to event data.
	log.Infoln("Trigger workflow")
	return webhookResponse{}, nil
}

func createWorkflowRun(tenant, wftName string, data *scm.EventData) error {
	log.Infof("data: %v\n", data)
	ns := common.TenantNamespace(tenant)
	wft, err := handler.K8sClient.CycloneV1alpha1().WorkflowTriggers(ns).Get(wftName, metav1.GetOptions{})
	if err != nil {
		return cerr.ConvertK8sError(err)
	}

	wfName := wft.Spec.WorkflowRef.Name
	if wfName == "" {
		return fmt.Errorf("Workflow reference of workflowtrigger is empty")
	}

	trigger := false
	var tag string
	st := wft.Spec.SCM
	switch data.Type {
	case scm.TagReleaseEventType:
		if st.TagRelease.Enabled {
			trigger = true
			tag = data.Ref
		}
	case scm.PushEventType:
		for _, branch := range st.Push.Branches {
			if branch == data.Branch {
				trigger = true
				break
			}
		}
	case scm.PullRequestEventType:
		if st.PullRequest.Enabled {
			trigger = true
		}
	case scm.PullRequestCommentEventType:
		for _, comment := range st.PullRequestComment.Comments {
			if comment == data.Comment {
				trigger = true
			}
		}
	}

	if !trigger {
		return nil
	}
	// wf, err := handler.K8sClient.CycloneV1alpha1().Workflows(ns).Get(wfName, metav1.GetOptions{})
	// if err != nil {
	// 	return cerr.ConvertK8sError(err)
	// }

	log.Infof("Trigger wft %s with event data: %v", wftName, data)

	if tag == "" {
		tag = fmt.Sprintf("%s-%s", wfName, rand.String(5))
	}

	// Create workflowrun.
	wfr := &v1alpha1.WorkflowRun{
		ObjectMeta: metav1.ObjectMeta{
			Name: tag,
		},
		Spec: wft.Spec.WorkflowRunSpec,
	}

	// Set "Tag" for all resource configs.
	for _, r := range wft.Spec.WorkflowRunSpec.Resources {
		for i, p := range r.Parameters {
			if p.Name == "TAG" && p.Value == "" {
				r.Parameters[i].Value = tag
			}
		}
	}

	// Set "Tag" for all stage configs.
	for _, s := range wft.Spec.WorkflowRunSpec.Stages {
		for i, p := range s.Parameters {
			if p.Name == "tag" && p.Value == "" {
				s.Parameters[i].Value = tag
			}
		}
	}

	_, err = handler.K8sClient.CycloneV1alpha1().WorkflowRuns(ns).Create(wfr)
	if err != nil {
		return cerr.ConvertK8sError(err)
	}

	log.Infoln("Create workflowrun by workflow trigger")
	return nil
}

// getHttpRequest gets request from context.
func getHttpRequest(ctx context.Context) *http.Request {
	return service.HTTPContextFrom(ctx).Request()
}
