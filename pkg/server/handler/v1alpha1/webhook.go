package v1alpha1

import (
	"context"
	"fmt"
	"strings"
	"time"

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
	"github.com/caicloud/cyclone/pkg/util"
	"github.com/caicloud/cyclone/pkg/util/cerr"
)

const (
	succeededMsg = "Successfully triggered"

	ignoredMsg = "Is ignored"
)

type task struct {
	Tenant    string
	Trigger   v1alpha1.WorkflowTrigger
	EventData *scm.EventData
}

var taskQueue = make(chan task, 10)

func init() {
	go webhookWorker(taskQueue)
}

func webhookWorker(tasks <-chan task) {
	for t := range tasks {
		err := createWorkflowRun(t.Tenant, t.Trigger, t.EventData)
		if err != nil {
			log.Errorf("wft %s create workflow run error:%v", t.Trigger.Name, err)
		}
	}
}

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
	// convert the time to UTC timezone
	data.CreatedAt = data.CreatedAt.UTC()

	wfts, err := hook.ListSCMWfts(tenant, data.Repo, integration)
	if err != nil {
		return newWebhookResponse(err.Error()), err
	}

	triggeredWfts := make([]string, 0)
	for _, wft := range wfts.Items {
		log.Infof("Trigger workflow trigger %s", wft.Name)
		taskQueue <- task{
			Tenant:    tenant,
			Trigger:   wft,
			EventData: data,
		}
		triggeredWfts = append(triggeredWfts, wft.Name)
	}
	if len(triggeredWfts) > 0 {
		return newWebhookResponse(fmt.Sprintf("%s: %s", succeededMsg, triggeredWfts)), nil
	}

	return newWebhookResponse(ignoredMsg), nil
}

func sanitizeRef(ref string) string {
	ret := make([]rune, 0, len(ref))
	for _, ch := range ref {
		switch {
		case ch >= 'a' && ch <= 'z':
		case ch >= 'A' && ch <= 'Z':
		case ch >= '0' && ch <= '9':
		default:
			ch = '_'
		}
		ret = append(ret, ch)
	}
	return string(ret)
}

func createWorkflowRun(tenant string, wft v1alpha1.WorkflowTrigger, data *scm.EventData) error {
	ns := wft.Namespace
	cancelPrevious := false
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
			// If tag contains "/", trim it.
			if index := strings.LastIndex(tag, "/"); index >= 0 && len(tag) > index+1 {
				tag = tag[index+1:]
			}
		}
	case scm.PushEventType:
		trimmedBranch := data.Branch
		if index := strings.LastIndex(trimmedBranch, "/"); index >= 0 && len(trimmedBranch) > index+1 {
			trimmedBranch = trimmedBranch[index+1:]
		}
		for _, branch := range st.Push.Branches {
			if branch == trimmedBranch {
				trigger = true
				break
			}
		}
	case scm.PullRequestEventType:
		log.Infof("target branch name: %s", data.Branch)
		if !st.PullRequest.Enabled {
			break
		}

		// Always trigger if Branches are not specified
		if len(st.PullRequest.Branches) == 0 {
			trigger = true
			cancelPrevious = true
			break
		}

		for _, branch := range st.PullRequest.Branches {
			if branch == data.Branch {
				trigger = true
				cancelPrevious = true
				break
			}
		}
	case scm.PullRequestCommentEventType:
		for _, comment := range st.PullRequestComment.Comments {
			if comment == data.Comment {
				trigger = true
				cancelPrevious = true
				break
			}
		}
	case scm.PostCommitEventType:
		if !st.PostCommit.Enabled {
			break
		}

		// For Backward Compatibility, old version workflowTriggers lack of these field
		// and can normally trigger workflow
		if st.PostCommit.RootURL == "" || st.PostCommit.WorkflowURL == "" ||
			data.ChangedFiles == nil || len(data.ChangedFiles) == 0 {
			trigger = true
			break
		}

		for _, file := range data.ChangedFiles {
			fullPath := st.PostCommit.RootURL + "/" + file
			if strings.Contains(fullPath, st.PostCommit.WorkflowURL) {
				trigger = true
				break
			}
		}

	}

	if !trigger {
		return nil
	}

	cycloneClient := handler.K8sClient.CycloneV1alpha1()
	ctx := context.TODO()
	skipCurrent := false
	ref := sanitizeRef(data.Ref)

	var currentWfrs []v1alpha1.WorkflowRun

	if cancelPrevious {
		wfrs, err := cycloneClient.WorkflowRuns(ns).List(metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s,%s=%s,%s=%s",
				meta.LabelProjectName, project,
				meta.LabelWorkflowName, wfName,
				meta.LabelWorkflowRunPRRef, ref),
		})
		if err != nil {
			log.Warningf("Fail to list previous WorkflowRuns: %v. project=%s, workflow=%s, ref=%s, trigger=%s", err,
				project, wfName, ref, data.Type)
		} else if !data.CreatedAt.IsZero() {
			for _, item := range wfrs.Items {
				if len(item.Annotations) == 0 {
					continue
				}
				evtType := scm.EventType(item.Annotations[meta.AnnotationWorkflowRunTrigger])
				if util.IsWorkflowRunTerminated(&item) ||
					(evtType != scm.PullRequestEventType && evtType != scm.PullRequestCommentEventType) {
					continue
				}

				currentWfrs = append(currentWfrs, item)

				updatedAtStr := item.Annotations[meta.AnnotationWorkflowRunPRUpdatedAt]
				if len(updatedAtStr) == 0 {
					continue
				}
				updatedAt, err := time.Parse(time.RFC3339, updatedAtStr)
				if err != nil {
					log.Warningf("Fail to parse pr-updated-at: %s: %v. ns=%s, wfr=%s", updatedAtStr, err,
						item.Namespace, item.Name)
					continue
				}
				if updatedAt.After(data.CreatedAt) {
					skipCurrent = true
					log.Infof("There is already running workflowRun for PR %s. Ignore this event. Existing wfr is %s/%s", data.Ref, item.Namespace, item.Name)
					break
				}
			}
		} else {
			log.Infof("The update time of PR %s/%s is unknown. Turn off canceling previous builds.", data.Repo, data.Ref)
			cancelPrevious = false
		}
	}

	if skipCurrent {
		return nil
	}

	log.Infof("Trigger wft %s with event data: %v", wft.Name, data)

	name := fmt.Sprintf("%s-%s", wfName, rand.String(5))

	// Create workflowrun.
	wfr := &v1alpha1.WorkflowRun{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				meta.AnnotationWorkflowRunPRUpdatedAt: data.CreatedAt.Format(time.RFC3339),
				meta.AnnotationWorkflowRunTrigger:     string(data.Type),
				meta.AnnotationAlias:                  name,
			},
			Labels: map[string]string{
				meta.LabelProjectName:             project,
				meta.LabelWorkflowName:            wfName,
				meta.LabelWorkflowRunAcceleration: wft.Labels[meta.LabelWorkflowRunAcceleration],
			},
		},
		Spec: wft.Spec.WorkflowRunSpec,
	}
	if cancelPrevious {
		wfr.ObjectMeta.Labels[meta.LabelWorkflowRunPRRef] = ref
	}

	wfr.Annotations, err = setSCMEventData(wfr.Annotations, data)
	if err != nil {
		return err
	}

	// Set "Tag" and "SCM_REVISION" for all resource configs.
	for _, r := range wft.Spec.WorkflowRunSpec.Resources {
		for i, p := range r.Parameters {
			if p.Name == "TAG" && tag != "" {
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
			if p.Name == "tag" && tag != "" {
				s.Parameters[i].Value = &tag
			}
		}
	}

	accelerator.NewAccelerator(tenant, project, wfr).Accelerate()
	_, err = cycloneClient.WorkflowRuns(ns).Create(wfr)
	if err != nil {
		return cerr.ConvertK8sError(err)
	}

	go func(wfrCopy *v1alpha1.WorkflowRun) {
		// Init pull-request status to pending
		wfrCopy.Status.Overall.Phase = v1alpha1.StatusRunning
		err = updatePullRequestStatus(wfrCopy)
		if err != nil {
			log.Warningf("Init pull request status for %s error: %v", wfr.Name, err)
		}
	}(wfr.DeepCopy())

	if !cancelPrevious {
		return nil
	}

	log.Infof("Trying to cancel %d previous builds for PR %s. repo=%s", len(currentWfrs), data.Ref, data.Repo)
	for _, item := range currentWfrs {
		_, err := stopWorkflowRun(ctx, &item, "AutoCancelPreviousBuild")
		if err != nil {
			log.Warningf("Fail to stop previous WorkflowRun %s/%s: %v. trigger=%s", err,
				item.Namespace, item.Name, data.Type)
		}
	}

	return nil
}
