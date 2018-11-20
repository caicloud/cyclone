package workflowrun

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

type workflowRunItem struct {
	name       string
	namespace  string
	expireTime time.Time
}

func (i *workflowRunItem) String() string {
	return fmt.Sprintf("%s:%s", i.namespace, i.name)
}

func NewWorkflowRunItem(wfr *v1alpha1.WorkflowRun) *workflowRunItem {
	timeout, _ := ParseTime(wfr.Spec.Timeout)
	return &workflowRunItem{
		name:       wfr.Name,
		namespace:  wfr.Namespace,
		expireTime: time.Now().Add(timeout),
	}
}

// TimeoutManager manages timeout of WorkflowRun.
type TimeoutManager struct {
	client   clientset.Interface
	operator *Operator
	items    map[string]*workflowRunItem
}

// NewTimeoutManager creates a timeout manager and run it.
func NewTimeoutManager(client clientset.Interface, operator *Operator) *TimeoutManager {
	manager := &TimeoutManager{
		client:   client,
		operator: operator,
		items:    make(map[string]*workflowRunItem),
	}
	go manager.Run()
	return manager
}

// Add adds a WorkflowRun to the timeout manager.
func (m *TimeoutManager) Add(wfr *v1alpha1.WorkflowRun) error {
	_, err := ParseTime(wfr.Spec.Timeout)
	if err != nil {
		return fmt.Errorf("invalid timeout value '%s', error: %v", wfr.Spec.Timeout, err)
	}

	item := NewWorkflowRunItem(wfr)
	m.items[item.String()] = item

	return nil
}

// Run will check timeout of managed WorkflowRun and process items that have expired their time.
func (m *TimeoutManager) Run() {
	ticker := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-ticker.C:
			m.process()
		}
	}
}

func (m *TimeoutManager) process() {
	var expired []*workflowRunItem
	for _, v := range m.items {
		if v.expireTime.Before(time.Now()) {
			expired = append(expired, v)
		}
	}

	for _, i := range expired {
		log.WithField("workflowrun", i.name).WithField("namespace", i.namespace).Info("Start to process expired WorkflowRun")
		wfr, err := m.client.CycloneV1alpha1().WorkflowRuns(i.namespace).Get(i.name, metav1.GetOptions{})
		if err != nil {
			log.WithField("workflowrun", wfr.Name).Error("Get WorkflowRun error: ", err)
			continue
		}

		if wfr.Status.Overall.Status != v1alpha1.StatusError && wfr.Status.Overall.Status != v1alpha1.StatusCompleted {
			wfr.Status.Overall = v1alpha1.Status{
				Status:             v1alpha1.StatusError,
				Reason:             "Timeout",
				LastTransitionTime: metav1.Time{time.Now()},
			}

			if err = m.operator.UpdateStatus(wfr); err != nil {
				log.WithField("workflowrun", wfr.Name).Error("Update WorkflowRun status error: ", err)
				continue
			}
		}

		// Kill stage pods, notice that already completed WorkflowRun will also have their Pod deleted.
		stages := wfr.Status.Stages
		if stages != nil {
			for stage, status := range stages {
				if status.Pod == nil {
					continue
				}
				log.WithField("workflowrun", wfr.Name).
					WithField("pod", status.Pod.Name).
					WithField("stage", stage).
					Info("To delete pod for expired WorkflowRun")
				err = m.client.CoreV1().Pods(status.Pod.Namespace).Delete(status.Pod.Name, &metav1.DeleteOptions{})
				if err != nil {
					log.Error("Delete pod error: ", err)
				}
			}
		}

		delete(m.items, i.String())
	}
}
