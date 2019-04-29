package v1alpha1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/meta"
	s_v1alpha1 "github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/biz/scm"
	"github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/server/config"
	"github.com/caicloud/cyclone/pkg/server/handler"
	utilhttp "github.com/caicloud/cyclone/pkg/util/http"
	"github.com/caicloud/cyclone/pkg/workflow/workload/pod"
)

// HandleWorkflowRunNotification handles workflowrun finished notification from workflow engine.
func HandleWorkflowRunNotification(ctx context.Context, wfr *v1alpha1.WorkflowRun) (interface{}, error) {
	err := updatePullRequestStatus(wfr)
	if err != nil {
		log.WithField("wfr", wfr.Name).Error("Failed to update SCM status: ", err)
	}

	err = sendNotifications(wfr)
	if err != nil {
		log.WithField("wfr", wfr.Name).Error("Failed to send notifications: ", err)
	}

	// Always return success as have received the workflowrun notifications.
	// No need to return the errors when handle it.
	return nil, nil
}

// sendNotifications send notifications to subscribe systems, and record the results in status.
func sendNotifications(wfr *v1alpha1.WorkflowRun) error {
	// If already there are notification status, no need to send notifications again.
	if wfr.Status.Notifications != nil {
		return nil
	}

	// Send notifications with workflowrun.
	bodyBytes, err := json.Marshal(wfr)
	if err != nil {
		log.WithField("wfr", wfr.Name).Error("Failed to marshal workflowrun: ", err)
		return err
	}
	body := bytes.NewReader(bodyBytes)

	status := make(map[string]v1alpha1.NotificationStatus)
	for _, endpoint := range config.Config.Notifications {
		req, err := http.NewRequest(http.MethodPost, endpoint.URL, body)
		if err != nil {
			err = fmt.Errorf("Failed to new notification request: %v", err)
			log.WithField("wfr", wfr.Name).Error(err)
			status[endpoint.Name] = v1alpha1.NotificationStatus{
				Result:  v1alpha1.NotificationResultFailed,
				Message: err.Error(),
			}
			continue
		}
		// Set Json content type in Http header.
		req.Header.Set(utilhttp.HeaderContentType, utilhttp.HeaderContentTypeJSON)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			s := v1alpha1.NotificationStatus{
				Result: v1alpha1.NotificationResultFailed,
			}

			log.WithField("wfr", wfr.Name).Errorf("Failed to send notification for %s: %v", endpoint.Name, err)
			if resp != nil {
				defer resp.Body.Close()
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Error(err)
					s.Message = err.Error()
				} else {
					s.Message = fmt.Sprintf("Status code: %d, error: %s", resp.StatusCode, body)
				}
			}

			status[endpoint.Name] = s
			continue
		}

		log.WithField("wfr", wfr.Name).Infof("Status code of notification for %s: %d", endpoint.Name, resp.StatusCode)
		status[endpoint.Name] = v1alpha1.NotificationStatus{
			Result:  v1alpha1.NotificationResultSucceeded,
			Message: fmt.Sprintf("Status code: %d", resp.StatusCode),
		}
	}

	// Update WorkflowRun notification status with retry.
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Get latest WorkflowRun.
		latest, err := handler.K8sClient.CycloneV1alpha1().WorkflowRuns(wfr.Namespace).Get(wfr.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		if latest.Status.Notifications == nil {
			latest.Status.Notifications = status
			_, err = handler.K8sClient.CycloneV1alpha1().WorkflowRuns(wfr.Namespace).Update(latest)
			return err
		}

		return nil
	})

	if err != nil {
		log.WithField("name", wfr.Name).Error("Update workflowrun notification status error: ", err)
	}

	return err
}

func updatePullRequestStatus(wfr *v1alpha1.WorkflowRun) error {
	wfrName := wfr.Name
	if wfr.Annotations == nil {
		return fmt.Errorf("Annotations of workflowrun %s can not be empty", wfrName)
	}

	trigger, ok := wfr.Annotations[meta.AnnotationWorkflowRunTrigger]
	if !ok {
		return fmt.Errorf("Trigger of workflowrun %s can not be empty", wfrName)
	}

	// Skip workflowruns whose trigger are not about pull request.
	if trigger != string(scm.PullRequestCommentEventType) && trigger != string(scm.PullRequestEventType) {
		return nil
	}

	labels := wfr.Labels
	if labels == nil {
		return fmt.Errorf("Labels of workflowrun %s can not be empty", wfrName)
	}

	var tenant, project, wfName string
	tenant = common.NamespaceTenant(wfr.Namespace)
	if project, ok = labels[meta.LabelProjectName]; !ok {
		return fmt.Errorf("Fail to get project name from labels of workflowrun %s", wfrName)
	}
	if wfName, ok = labels[meta.LabelWorkflowName]; !ok {
		return fmt.Errorf("Fail to get workflow name from labels of workflowrun %s", wfrName)
	}

	scm, err := getSCMSourceFromWorkflowRun(wfr)
	if err != nil {
		return err
	}

	event, err := getSCMEventData(wfr.Annotations)
	if err != nil {
		return err
	}

	recordURL, err := generateRecordURL(tenant, project, wfName, wfrName)
	if err != nil {
		return err
	}

	return createSCMStatus(scm, wfr.Status.Overall.Phase, recordURL, event)
}

func getSCMSourceFromWorkflowRun(wfr *v1alpha1.WorkflowRun) (*s_v1alpha1.SCMSource, error) {
	// Parse "GIT_AUTH" from resource parameters. ONLY support one SCM git resource.
	var gitToken string
	found := false
	for _, r := range wfr.Spec.Resources {
		for i, p := range r.Parameters {
			if p.Name == "GIT_AUTH" {
				gitToken = *r.Parameters[i].Value
				found = true
				break
			}
		}

		if found {
			break
		}
	}

	if found {
		secretRef := pod.NewSecretRefValue()
		if err := secretRef.Parse(gitToken); err != nil {
			return nil, err
		}

		in, err := getIntegration(common.NamespaceTenant(secretRef.Namespace), secretRef.Secret)
		if err != nil {
			return nil, err
		}
		if in.Spec.Type != s_v1alpha1.SCM {
			return nil, fmt.Errorf("Type of integration %s is %s, should be %s", in.Name, in.Spec.Type, s_v1alpha1.SCM)
		}

		return in.Spec.SCM, nil
	}

	return nil, fmt.Errorf("Can not get SCM source from workflowrun %s", wfr.Name)
}
