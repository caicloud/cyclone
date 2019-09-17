package workflowrun

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/workflow/common"
)

const regString = "(\\d+h)(our)?|(\\d+m)(in)?|(\\d+s)(econd)?"

var timeParser = regexp.MustCompile(regString)
var matcherReg = regexp.MustCompile(fmt.Sprintf("^(%s)+$", regString))
var timeMap = map[string]time.Duration{
	"h": time.Hour,
	"m": time.Minute,
	"s": time.Second,
}

// ParseTime parses time string like '30min', '2h30m' to time.Time
func ParseTime(t string) (time.Duration, error) {
	if !matcherReg.Match([]byte(strings.ToLower(t))) {
		return 0, fmt.Errorf("%s invalid", t)
	}

	matches := timeParser.FindAllStringSubmatch(strings.ToLower(t), -1)
	if len(matches) == 0 {
		return 0, fmt.Errorf("invalid time string %s", t)
	}
	var result time.Duration
	for _, m := range matches {
		part := m[1] + m[3] + m[5]
		l := len(part)
		n, err := strconv.Atoi(part[0 : l-1])
		if err != nil {
			return 0, err
		}
		result += timeMap[string(part[l-1])] * time.Duration(n)
	}

	return result, nil
}

func newWorkflowRunItem(wfr *v1alpha1.WorkflowRun) *workflowRunItem {
	timeout, _ := ParseTime(wfr.Spec.Timeout)
	return &workflowRunItem{
		name:       wfr.Name,
		namespace:  wfr.Namespace,
		expireTime: time.Now().Add(timeout),
	}
}

// TimeoutProcessor manages timeout of WorkflowRun.
type TimeoutProcessor struct {
	client   clientset.Interface
	recorder record.EventRecorder
	items    map[string]*workflowRunItem
}

// NewTimeoutProcessor creates a timeout manager and run it.
func NewTimeoutProcessor(client clientset.Interface) *TimeoutProcessor {
	manager := &TimeoutProcessor{
		client:   client,
		recorder: common.GetEventRecorder(client, common.EventSourceWfrController),
		items:    make(map[string]*workflowRunItem),
	}
	go manager.Run(time.Second * 5)
	return manager
}

// Add adds a WorkflowRun to the timeout manager.
func (m *TimeoutProcessor) Add(wfr *v1alpha1.WorkflowRun) error {
	_, err := ParseTime(wfr.Spec.Timeout)
	if err != nil {
		return fmt.Errorf("invalid timeout value '%s', error: %v", wfr.Spec.Timeout, err)
	}

	item := newWorkflowRunItem(wfr)
	m.items[item.String()] = item

	return nil
}

// Run will check timeout of managed WorkflowRun and process items that have expired their time.
func (m *TimeoutProcessor) Run(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		m.process()
	}
}

func (m *TimeoutProcessor) process() {
	var expired []*workflowRunItem
	for _, v := range m.items {
		if v.expireTime.Before(time.Now()) {
			expired = append(expired, v)
		}
	}

	for _, i := range expired {
		log.WithField("wfr", i.name).WithField("namespace", i.namespace).Info("Start to process expired WorkflowRun")
		wfr, err := m.client.CycloneV1alpha1().WorkflowRuns(i.namespace).Get(i.name, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				delete(m.items, i.String())
			} else {
				log.WithField("wfr", wfr.Name).Error("Get WorkflowRun error: ", err)
			}
			continue
		}
		m.recorder.Event(wfr, corev1.EventTypeWarning, "Timeout", "WorkflowRun execution timeout")

		clusterClient := common.GetExecutionClusterClient(wfr)
		if clusterClient == nil {
			log.WithField("wfr", wfr.Name).Error("Execution cluster client not found")
			continue
		}

		if wfr.Status.Overall.Phase != v1alpha1.StatusFailed && wfr.Status.Overall.Phase != v1alpha1.StatusSucceeded {
			wfr.Status.Overall = v1alpha1.Status{
				Phase:              v1alpha1.StatusFailed,
				Reason:             "Timeout",
				LastTransitionTime: metav1.Time{Time: time.Now()},
			}

			operator := operator{
				clusterClient: clusterClient,
				client:        m.client,
				wfr:           wfr,
			}
			if err = operator.Update(); err != nil {
				log.WithField("wfr", wfr.Name).Error("Update WorkflowRun status error: ", err)
				continue
			}
		}

		// Kill stage pods.
		stages := wfr.Status.Stages
		for stage, status := range stages {
			if status.Pod == nil {
				continue
			}
			log.WithField("wfr", wfr.Name).
				WithField("pod", status.Pod.Name).
				WithField("stg", stage).
				Info("To delete pod for expired WorkflowRun")
			err = clusterClient.CoreV1().Pods(status.Pod.Namespace).Delete(status.Pod.Name, &metav1.DeleteOptions{})
			if err != nil {
				log.Error("Delete pod error: ", err)
			}
		}

		delete(m.items, i.String())
		m.recorder.Event(wfr, corev1.EventTypeWarning, "Timeout", "Stages stopped due to timeout")
	}
}
