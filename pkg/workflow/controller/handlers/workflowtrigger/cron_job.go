package workflowtrigger

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	ccommon "github.com/caicloud/cyclone/pkg/common"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/meta"
	"github.com/caicloud/cyclone/pkg/server/biz/accelerator"
	"github.com/caicloud/cyclone/pkg/server/common"
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
		t.WorkflowRun.Annotations[meta.AnnotationAlias] = t.WorkflowRun.Name
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

// CreateCron creates a cron trigger from workflow trigger, and add it to cron trigger manager.
func (m *CronTriggerManager) CreateCron(wft *v1alpha1.WorkflowTrigger) {
	ct := &CronTrigger{
		Namespace:           wft.Namespace,
		WorkflowTriggerName: wft.Name,
	}

	wfr := &v1alpha1.WorkflowRun{
		ObjectMeta: meta_v1.ObjectMeta{
			Annotations: map[string]string{
				meta.AnnotationWorkflowRunTrigger: common.CronTimerTrigger,
			},
			Labels: map[string]string{
				meta.LabelWorkflowName: wft.Spec.WorkflowRunSpec.WorkflowRef.Name,
			},
		},
		Spec: wft.Spec.WorkflowRunSpec,
	}

	// If controller instance name is set, add label to the pod created.
	if instance := os.Getenv(ccommon.ControllerInstanceEnvName); len(instance) != 0 {
		wfr.ObjectMeta.Labels[meta.LabelControllerInstance] = instance
	}

	var project, acceleration string
	if wft.Labels != nil {
		if projectName, ok := wft.Labels[meta.LabelProjectName]; ok {
			wfr.Labels[meta.LabelProjectName] = projectName
			project = projectName
		}

		if acc, ok := wft.Labels[meta.LabelWorkflowRunAcceleration]; ok {
			wfr.Labels[meta.LabelWorkflowRunAcceleration] = acc
			acceleration = acc
		}
	}

	if acceleration == meta.LabelValueTrue {
		if project == "" {
			log.Warningf("workflowruns triggered by workflowtrigger %s will not be accelerated as its project is empty", wft.Name)
		} else {
			tenant := common.NamespaceTenant(wft.Namespace)
			accelerator.NewAccelerator(tenant, project, wfr).Accelerate()
		}
	}

	ct.WorkflowRun = wfr

	var schedule cron.Schedule
	var err error
	schedExpr := wft.Spec.Cron.Schedule
	parts := strings.Fields(schedExpr)
	if len(parts) == 5 {
		schedule, err = cron.ParseStandard(schedExpr)
	} else {
		schedule, err = cron.Parse(schedExpr)
	}

	if err != nil {
		log.Errorf("can not parse cron expression %s: %s", schedExpr, err)
		return
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
