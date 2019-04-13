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

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/config"
	"github.com/caicloud/cyclone/pkg/server/handler"
	utilhttp "github.com/caicloud/cyclone/pkg/util/http"
)

// ReceiveWorkflowRunNotification ...
func ReceiveWorkflowRunNotification(ctx context.Context, wfr *v1alpha1.WorkflowRun) (interface{}, error) {
	_, err := sendNotifications(wfr)
	return nil, err
}

// sendNotifications send notifications for workflowruns when:
// * its workflow has notification config
// * finish time after workflow controller starts
// * notification status of workflowrun is nil
// If the returned notification status is nil, it means that there is no need to send notification.
func sendNotifications(wfr *v1alpha1.WorkflowRun) (map[string]v1alpha1.NotificationStatus, error) {
	if wfr.Status.Notifications != nil {
		return nil, nil
	}

	wfRef := wfr.Spec.WorkflowRef
	if wfRef == nil {
		return nil, fmt.Errorf("Workflow reference of workflow run %s/%s is empty", wfr.Namespace, wfr.Name)
	}
	wf, err := handler.K8sClient.CycloneV1alpha1().Workflows(wfRef.Namespace).Get(wfRef.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if len(wf.Spec.Notification.Receivers) == 0 {
		return nil, nil
	}

	// Send notifications with workflowrun.
	bodyBytes, err := json.Marshal(wfr)
	if err != nil {
		log.WithField("wfr", wfr.Name).Error("Failed to marshal workflowrun: ", err)
		return nil, err
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
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Error(err)
				s.Message = err.Error()
			} else {
				s.Message = fmt.Sprintf("Status code: %d, error: %s", resp.StatusCode, body)
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

	return status, nil
}
