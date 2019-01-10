package workflowtrigger

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/workflow/common"
	"github.com/pkg/errors"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/rand"
)

const (
	// KeyTemplate ...
	KeyTemplate = "%s/%s"

	// ParamTimeZoneOffset is time zone offset to UTC, in minutes
	// -480 for Asia/Shanghai(+8)
	ParamTimeZoneOffset = "timeZoneOffset"
	// ParamSchedule ...
	ParamSchedule = "schedule"
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

// CronTriggerManager ...
type CronTriggerManager struct {
	Client         clientset.Interface
	CronTriggerMap map[string]*CronTrigger
	mutex          sync.Mutex
}

// NewTriggerManager ...
func NewTriggerManager(client clientset.Interface) *CronTriggerManager {
	return &CronTriggerManager{
		Client:         client,
		CronTriggerMap: make(map[string]*CronTrigger),
		mutex:          sync.Mutex{},
	}
}

// AddTrigger ...
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

// DeleteTrigger ...
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

// ToWorkflowTrigger ...
func ToWorkflowTrigger(obj interface{}) (*v1alpha1.WorkflowTrigger, error) {

	wft, ok := obj.(*v1alpha1.WorkflowTrigger)
	if !ok {
		return nil, errors.New(fmt.Sprintf("I want type: WorkflowTrigger, but it is: %T", obj))
	}

	return wft, nil
}

func (t *CronTrigger) getKeyFromTrigger() string {
	return fmt.Sprintf(KeyTemplate, t.Namespace, t.WorkflowTriggerName)
}

// Run ...
func (t *CronTrigger) Run() {

	if t.WorkflowRun.Labels == nil {
		t.WorkflowRun.Labels = make(map[string]string)
	}
	t.WorkflowRun.Labels[common.WorkflowRunLabelName] = t.WorkflowRun.Spec.WorkflowRef.Name

	for {
		t.WorkflowRun.Name = fmt.Sprintf("%s-%s", t.WorkflowTriggerName, rand.String(5))
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

// CreateCron ...
func (m *CronTriggerManager) CreateCron(wft *v1alpha1.WorkflowTrigger) {

	schedule, has := getParamValue(wft.Spec.Parameters, ParamSchedule)
	if !has {
		log.WithField("wft", wft.Name).Warn("Parameter 'schedule' not set in WorkflowTrigger spec")
		return
	}

	timezone, has := getParamValue(wft.Spec.Parameters, ParamTimeZoneOffset)
	if !has {
		timezone = "0"
	}

	minuteOffUTC, err := strconv.Atoi(timezone)
	if err != nil {
		log.Warnf("can not parse timezone(%s) to int", timezone)
		minuteOffUTC = 0
	}

	ct := &CronTrigger{
		Namespace:           wft.Namespace,
		WorkflowTriggerName: wft.Name,
	}

	wfr := &v1alpha1.WorkflowRun{
		Spec: wft.Spec.WorkflowRunSpec,
	}

	ct.WorkflowRun = wfr

	c := cron.NewWithLocation(time.FixedZone("userZone", -1*60*minuteOffUTC))
	err = c.AddJob(schedule, ct)
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

// UpdateCron ...
func (m *CronTriggerManager) UpdateCron(wft *v1alpha1.WorkflowTrigger) {
	m.DeleteCron(wft)
	m.CreateCron(wft)
}

// DeleteCron ...
func (m *CronTriggerManager) DeleteCron(wft *v1alpha1.WorkflowTrigger) {
	wftKey := getKeyFromWorkflowTrigger(wft)
	m.DeleteTrigger(wftKey)
}

func getKeyFromWorkflowTrigger(wft *v1alpha1.WorkflowTrigger) string {
	return fmt.Sprintf(KeyTemplate, wft.Namespace, wft.Name)
}
