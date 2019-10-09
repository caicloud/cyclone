package v1alpha1

import (
	"context"
	"fmt"
	"strings"

	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/service"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/meta"
	api "github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/biz/accelerator"
	"github.com/caicloud/cyclone/pkg/server/biz/hook"
	"github.com/caicloud/cyclone/pkg/server/biz/scm"
	"github.com/caicloud/cyclone/pkg/server/biz/scm/bitbucket"
	"github.com/caicloud/cyclone/pkg/server/biz/scm/github"
	"github.com/caicloud/cyclone/pkg/server/biz/scm/gitlab"
	"github.com/caicloud/cyclone/pkg/server/biz/scm/svn"
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
func HandleWebhook(ctx context.Context, tenant, eventType, integration string) (api.WebhookResponse, error) {
	if eventType != string(v1alpha1.TriggerTypeSCM) {
		err := fmt.Errorf("eventType %s unsupported, support SCM for now", eventType)
		return newWebhookResponse(err.Error()), err
	}
	request := service.HTTPContextFrom(ctx).Request()

	var data *scm.EventData

	if request.Header.Get(github.EventTypeHeader) != "" {
		in, err := getIntegration(common.TenantNamespace(tenant), integration)
		if err != nil {
			return newWebhookResponse(err.Error()), err
		}
		data = github.ParseEvent(in.Spec.SCM, request)
	}

	if request.Header.Get(gitlab.EventTypeHeader) != "" {
		data = gitlab.ParseEvent(request)
	}

	if request.Header.Get(bitbucket.EventTypeHeader) != "" {
		in, err := getIntegration(common.TenantNamespace(tenant), integration)
		if err != nil {
			return newWebhookResponse(err.Error()), err
		}
		data = bitbucket.ParseEvent(in.Spec.SCM, request)
	}

	if request.Header.Get(svn.EventTypeHeader) != "" {
		data = svn.ParseEvent(request)
	}

	if data == nil {
		return newWebhookResponse(ignoredMsg), nil
	}

	wfts, err := hook.ListSCMWfts(tenant, data.Repo, integration)
	if err != nil {
		return newWebhookResponse(err.Error()), err
	}

	sftName := make([]string, 0)
	for _, wft := range wfts.Items {
		log.Infof("Trigger workflow trigger %s", wft.Name)
		sftName = append(sftName, wft.Name)
		if err = createWorkflowRun(tenant, wft, data); err != nil {
			log.Errorf("wft %s create workflow run error:%v", wft.Name, err)
		}
	}
	if len(sftName) > 0 {
		return newWebhookResponse(fmt.Sprintf("%s: %s", succeededMsg, sftName)), nil
	}

	return newWebhookResponse(ignoredMsg), nil
}

func createWorkflowRun(tenant string, wft v1alpha1.WorkflowTrigger, data *scm.EventData) error {
	ns := wft.Namespace
	var err error
	var project string
	if wft.Labels != nil {
		project = wft.Labels[meta.LabelProjectName]
	}
	if project == "" {
		return fmt.Errorf("failed to get project from workflowtrigger labels")
	}

	wfName := wft.Spec.WorkflowRef.Name
	if wfName == "" {
		return fmt.Errorf("workflow reference of workflowtrigger is empty")
	}

	trigger := false
	var tag string
	st := wft.Spec.SCM
	switch data.Type {
	case scm.TagReleaseEventType:
		if st.TagRelease.Enabled {
			trigger = true
			tag = data.Ref
			splitTags := strings.Split(data.Ref, "/")
			if len(splitTags) == 3 {
				tag = splitTags[2]
			}
		}
	case scm.PushEventType:
		trimmedBranch := data.Branch
		if index := strings.LastIndex(data.Branch, "/"); index >= 0 {
			trimmedBranch = trimmedBranch[index+1:]
		}
		for _, branch := range st.Push.Branches {
			if branch == trimmedBranch {
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
	case scm.PostCommitEventType:
		if st.PostCommit.Enabled {
			trigger = true
		}
	}

	if !trigger {
		return nil
	}

	log.Infof("Trigger wft %s with event data: %v", wft.Name, data)

	suffix := rand.String(5)
	name := fmt.Sprintf("%s-%s", wfName, suffix)
	alias := name
	if tag != "" {
		alias = tag
	} else {
		tag = suffix
	}

	// Create workflowrun.
	wfr := &v1alpha1.WorkflowRun{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				meta.AnnotationWorkflowRunTrigger: string(data.Type),
				meta.AnnotationAlias:              alias,
			},
			Labels: map[string]string{
				meta.LabelProjectName:             project,
				meta.LabelWorkflowName:            wfName,
				meta.LabelWorkflowRunAcceleration: wft.Labels[meta.LabelWorkflowRunAcceleration],
			},
		},
		Spec: wft.Spec.WorkflowRunSpec,
	}

	wfr.Annotations, err = setSCMEventData(wfr.Annotations, data)
	if err != nil {
		return err
	}

	// Set "Tag" and "SCM_REVISION" for all resource configs.
	for _, r := range wft.Spec.WorkflowRunSpec.Resources {
		for i, p := range r.Parameters {
			if p.Name == "TAG" {
				r.Parameters[i].Value = &tag
			}

			if p.Name == "SCM_REVISION" && data.Ref != "" {
				r.Parameters[i].Value = &data.Ref
			}
		}
	}

	// Set "Tag" for all stage configs.
	for _, s := range wft.Spec.WorkflowRunSpec.Stages {
		for i, p := range s.Parameters {
			if p.Name == "tag" {
				s.Parameters[i].Value = &tag
			}
		}
	}

	accelerator.NewAccelerator(tenant, project, wfr).Accelerate()
	_, err = handler.K8sClient.CycloneV1alpha1().WorkflowRuns(ns).Create(wfr)
	if err != nil {
		return cerr.ConvertK8sError(err)
	}

	// Init pull-request status to pending
	wfrCopy := wfr.DeepCopy()
	wfrCopy.Status.Overall.Phase = v1alpha1.StatusRunning
	err = updatePullRequestStatus(wfrCopy)
	if err != nil {
		log.Warningf("Init pull request status for %s error: %v", wfr.Name, err)
	}
	return nil
}
