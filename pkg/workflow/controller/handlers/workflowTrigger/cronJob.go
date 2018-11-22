package workflowTrigger

import (
	"fmt"
	"strings"
	"sync"

	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/pkg/errors"
	"github.com/robfig/cron"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
)

const (
	KeyTemplate = "%s/%s"
)

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

type CronTriggerManager struct {
	Client         clientset.Interface
	CronTriggerMap map[string]*CronTrigger
	mutex          sync.Mutex
}

func NewTriggerManager(client clientset.Interface) *CronTriggerManager {
	return &CronTriggerManager{
		Client:         client,
		CronTriggerMap: make(map[string]*CronTrigger),
		mutex:          sync.Mutex{},
	}
}

func (trigger *CronTrigger) getKeyFromTrigger() (string) {
	return fmt.Sprintf(KeyTemplate, trigger.Namespace, trigger.WorkflowTriggerName)
}

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

func ToWorkflowTrigger(obj interface{}) (*v1alpha1.WorkflowTrigger, error) {

	wft, ok := obj.(*v1alpha1.WorkflowTrigger)
	if !ok {
		return nil, errors.New(fmt.Sprintf("I want type: WorkflowTrigger, but it is: %T", obj))
	} else {
		return wft, nil
	}
}

func (c *CronTrigger) Run() {

	c.WorkflowRun.Name = c.WorkflowTriggerName + "-" + strings.Replace(uuid.NewV1().String(), "-", "", -1)
	_, err := c.Manage.Client.CycloneV1alpha1().WorkflowRuns(c.Namespace).Create(c.WorkflowRun)
	if err != nil {
		c.FailCount++
		log.Warnf("can not create WorkflowRun: %s", err)
	} else {
		c.SuccCount++
	}
}

func getParaValue(items []v1alpha1.ParameterItem, key string) (string, bool) {
	for _, item := range items {
		if item.Name == key {
			return item.Value, true
		}
	}
	return "", false
}

func (m *CronTriggerManager) CreateCron(wft *v1alpha1.WorkflowTrigger) {

	schedule, has := getParaValue(wft.Spec.Parameters, "schedule")
	if !has {
		log.WithField("wft", wft.Name).Warn("Parameter 'schedule' not set in WorkflowTrigger spec")
		return
	}

	ct := &CronTrigger{
		Namespace:           wft.Namespace,
		WorkflowTriggerName: wft.Name,
	}

	wfr := &v1alpha1.WorkflowRun{
		Spec: wft.Spec.WorkflowRunSpec,
	}

	ct.WorkflowRun = wfr

	c := cron.New()
	err := c.AddJob(schedule, ct)
	if err != nil {
		log.Errorf("can not create Cron job: %s", err)
		return
	}

	ct.Cron = c
	ct.Manage = m
	m.AddTrigger(ct)

	if !wft.Spec.Disabled {
		ct.Cron.Start()
		ct.IsRunning = true
	}
}

func (m *CronTriggerManager) UpdateCron(wft *v1alpha1.WorkflowTrigger) {
	m.DeleteCron(wft)
	m.CreateCron(wft)
}

func (m *CronTriggerManager) DeleteCron(wft *v1alpha1.WorkflowTrigger) {
	wftKey := getKeyFromWorkflowTrigger(wft)
	m.DeleteTrigger(wftKey)
}

func getKeyFromWorkflowTrigger(wft *v1alpha1.WorkflowTrigger) string {
	return fmt.Sprintf(KeyTemplate, wft.Namespace, wft.Name)
}
