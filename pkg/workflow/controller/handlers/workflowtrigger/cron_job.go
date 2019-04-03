package workflowtrigger

import (
	"fmt"
	"strings"
	"sync"

	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	s_common "github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/workflow/common"
)

const (
	// KeyTemplate ...
	KeyTemplate = "%s/%s"
)

// CronTrigger ...
type CronTrigger struct {
	Cron                *cron.Cron
	IsRunning           bool
	SuccCount           int
	FailCount           int
	Namespace           string
	WorkflowTriggerName string
	WorkflowRun         *v1alpha1.WorkflowRun
	Manage              *CronTriggerManager
}

// CronTriggerManager represents manager for cron triggers.
type CronTriggerManager struct {
	Client         clientset.Interface
	CronTriggerMap map[string]*CronTrigger
	mutex          sync.Mutex
}

// NewTriggerManager returns a cron trigger manager.
func NewTriggerManager(client clientset.Interface) *CronTriggerManager {
	return &CronTriggerManager{
		Client:         client,
		CronTriggerMap: make(map[string]*CronTrigger),
		mutex:          sync.Mutex{},
	}
}

// AddTrigger adds one cron trigger.
func (m *CronTriggerManager) AddTrigger(trigger *CronTrigger) {
	wftKey := trigger.getKeyFromTrigger()
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if wft, ok := m.CronTriggerMap[wftKey]; ok {
		// this situation may not happen for k8s will do same name check
		log.Warnf("failed to add cronTrigger, already exists: %+v\n", wft)
	} else {
		m.CronTriggerMap[wftKey] = trigger
	}
}

// DeleteTrigger deletes one cron trigger.
func (m *CronTriggerManager) DeleteTrigger(wftKey string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if wft, ok := m.CronTriggerMap[wftKey]; ok {
		wft.Cron.Stop()
		wft.IsRunning = false
		delete(m.CronTriggerMap, wftKey)
	} else {
		log.Warnf("failed to delete cronTrigger(%s), not exist\n", wftKey)
	}
}

// ToWorkflowTrigger converts to workflow trigger.
func ToWorkflowTrigger(obj interface{}) (*v1alpha1.WorkflowTrigger, error) {

	wft, ok := obj.(*v1alpha1.WorkflowTrigger)
	if !ok {
		return nil, fmt.Errorf("I want type: WorkflowTrigger, but it is: %T", obj)
	}

	return wft, nil
}

func (t *CronTrigger) getKeyFromTrigger() string {
	return fmt.Sprintf(KeyTemplate, t.Namespace, t.WorkflowTriggerName)
}

// Run triggers the workflows.
func (t *CronTrigger) Run() {
	for {
		t.WorkflowRun.Name = fmt.Sprintf("%s-%s", t.WorkflowTriggerName, rand.String(5))
		t.WorkflowRun.Annotations[s_common.AnnotationAlias] = t.WorkflowRun.Name
		_, err := t.Manage.Client.CycloneV1alpha1().WorkflowRuns(t.Namespace).Create(t.WorkflowRun)
		if err != nil {
			if errors2.IsAlreadyExists(err) {
				continue
			} else {
				t.FailCount++
				log.Warnf("can not create WorkflowRun: %s", err)
				break
			}
		} else {
			t.SuccCount++
			break
		}
	}
}

func getParamValue(items []v1alpha1.ParameterItem, key string) (string, bool) {
	for _, item := range items {
		if item.Name == key {
			return item.Value, true
		}
	}
	return "", false
}

// CreateCron creates a cron trigger from workflow trigger, and add it to cron trigger manager.
func (m *CronTriggerManager) CreateCron(wft *v1alpha1.WorkflowTrigger) {
	ct := &CronTrigger{
		Namespace:           wft.Namespace,
		WorkflowTriggerName: wft.Name,
	}

	wfr := &v1alpha1.WorkflowRun{
		ObjectMeta: meta_v1.ObjectMeta{
			Annotations: map[string]string{
				s_common.AnnotationTrigger: s_common.CronTimerTrigger,
			},
			Labels: map[string]string{
				common.WorkflowNameLabelName: wft.Spec.WorkflowRunSpec.WorkflowRef.Name,
			},
		},
		Spec: wft.Spec.WorkflowRunSpec,
	}

	if wft.Labels != nil {
		if projectName, ok := wft.Labels[s_common.LabelProjectName]; ok {
			wfr.Labels[s_common.LabelProjectName] = projectName
		}
	}

	ct.WorkflowRun = wfr

	schedExpr := wft.Spec.Cron.Schedule
	parts := strings.Fields(schedExpr)
	schedule, err := cron.Parse(schedExpr)
	if err != nil {
		log.Errorf("can not parse cron expression %s: %s", schedExpr, err)
		return
	}

	if len(parts) == 5 {
		schedule, err = cron.ParseStandard(schedExpr)
		if err != nil {
			log.Errorf("can not parse cron expression %s: %s", schedExpr, err)
			return
		}
	}

	c := cron.New()
	c.Schedule(schedule, ct)

	ct.Cron = c
	ct.Manage = m
	m.AddTrigger(ct)

	if !wft.Spec.Disabled {
		ct.Cron.Start()
		ct.IsRunning = true
	}
}

// UpdateCron updates cron trigger based on workflow trigger.
func (m *CronTriggerManager) UpdateCron(wft *v1alpha1.WorkflowTrigger) {
	m.DeleteCron(wft)
	m.CreateCron(wft)
}

// DeleteCron deletes cron trigger from cron trigger manager.
func (m *CronTriggerManager) DeleteCron(wft *v1alpha1.WorkflowTrigger) {
	wftKey := getKeyFromWorkflowTrigger(wft)
	m.DeleteTrigger(wftKey)
}

func getKeyFromWorkflowTrigger(wft *v1alpha1.WorkflowTrigger) string {
	return fmt.Sprintf(KeyTemplate, wft.Namespace, wft.Name)
}
