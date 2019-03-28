package v1alpha1

import (
	"context"
	"fmt"

	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/service"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	api "github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/biz/scm"
	"github.com/caicloud/cyclone/pkg/server/biz/scm/github"
	"github.com/caicloud/cyclone/pkg/server/biz/scm/gitlab"
	"github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/server/handler"
	"github.com/caicloud/cyclone/pkg/util/cerr"
)

const (
	succeededMsg = "Successfully triggered"

	ignoredMsg = "Is ignored"
)

func newWebhookResponse(msg string) api.WebhookResponse {
	return api.WebhookResponse{
		Message: msg,
	}
}

// HandleWebhook handles webhooks from integrated systems.
func HandleWebhook(ctx context.Context, tenant, integration string) (api.WebhookResponse, error) {
	request := service.HTTPContextFrom(ctx).Request()

	repos, err := getReposFromSecret(tenant, integration)
	if err != nil {
		log.Error(err)
		return newWebhookResponse(err.Error()), err
	}

	if len(repos) == 0 {
		return newWebhookResponse(ignoredMsg), nil
	}

	triggered := false
	if request.Header.Get(github.EventTypeHeader) != "" {
		if data := github.ParseEvent(request); data != nil {
			if wfts, ok := repos[data.Repo]; ok {
				for _, wft := range wfts {
					log.Infof("Trigger workflow trigger %s", wft)
					if err = createWorkflowRun(tenant, wft, data); err != nil {
						log.Error(err)
					}
				}
			}
			triggered = true
		}
	}

	if request.Header.Get(gitlab.EventTypeHeader) != "" {
		if data := gitlab.ParseEvent(request); data != nil {
			if wfts, ok := repos[data.Repo]; ok {
				for _, wft := range wfts {
					log.Infof("Trigger workflow trigger %s", wft)
					if err = createWorkflowRun(tenant, wft, data); err != nil {
						log.Error(err)
					}
				}
			}
			triggered = true
		}
	}

	if triggered {
		return newWebhookResponse(succeededMsg), nil
	}

	return newWebhookResponse(ignoredMsg), nil
}

func createWorkflowRun(tenant, wftName string, data *scm.EventData) error {
	ns := common.TenantNamespace(tenant)
	wft, err := handler.K8sClient.CycloneV1alpha1().WorkflowTriggers(ns).Get(wftName, metav1.GetOptions{})
	if err != nil {
		return cerr.ConvertK8sError(err)
	}

	var project string
	if wft.Labels != nil {
		project = wft.Labels[common.LabelProjectName]
	}
	if project == "" {
		return fmt.Errorf("Failed to get project from workflowtrigger labels")
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

	log.Infof("Trigger wft %s with event data: %v", wftName, data)

	name := fmt.Sprintf("%s-%s", wfName, rand.String(5))
	var alias string
	if tag != "" {
		alias = tag
	} else {
		alias = name
	}

	// Create workflowrun.
	wfr := &v1alpha1.WorkflowRun{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				common.AnnotationTrigger: string(data.Type),
				common.AnnotationAlias:   alias,
			},
			Labels: map[string]string{
				common.LabelProjectName:  project,
				common.LabelWorkflowName: wfName,
			},
		},
		Spec: wft.Spec.WorkflowRunSpec,
	}

	// Set "Tag" and "GIT_REVISION" for all resource configs if they are empty.
	for _, r := range wft.Spec.WorkflowRunSpec.Resources {
		for i, p := range r.Parameters {
			if p.Name == "TAG" && p.Value == "" {
				r.Parameters[i].Value = tag
			}

			if p.Name == "GIT_REVISION" && p.Value == "" {
				r.Parameters[i].Value = data.Ref
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

	return nil
}
