package workflowTrigger

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/robfig/cron"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
)

const (
	KeyTemplate = "%s-%s"
)

type CronTrigger struct {
	Cron                *cron.Cron
	IsRun               bool
	SuccCount           int
	FailCount           int
	Namespace           string
	WorkflowTriggerName string
	WorkflowRun         *v1alpha1.WorkflowRun
}

type CronTriggerManager struct {
	CronTriggerMap map[string]*CronTrigger
}

var (
	CronS  = NewTriggerManager()
	Client clientset.Interface
)

func NewTriggerManager() *CronTriggerManager {
	return &CronTriggerManager{
		CronTriggerMap: make(map[string]*CronTrigger),
	}
}

func (m *CronTriggerManager) AddTrigger(trigger *CronTrigger) {
	wftKey := fmt.Sprintf(KeyTemplate, trigger.Namespace, trigger.WorkflowTriggerName)
	if wft, ok := m.CronTriggerMap[wftKey]; ok {
		// this situation may not happen for k8s will do same name check
		log.Errorf("failed to add cronTrigger, already exists: %+v\n", wft)
	} else {
		m.CronTriggerMap[wftKey] = trigger
	}
}

func (m *CronTriggerManager) DeleteTrigger(wftKey string) {
	if wft, ok := m.CronTriggerMap[wftKey]; ok {
		wft.Cron.Stop()
		wft.IsRun = false
		delete(m.CronTriggerMap, wftKey)
	} else {
		log.Errorf("failed to delete cronTrigger(%s), not exist\n", wftKey)
	}
}

// [a-z0-9]{5}
// exp: ab12c
// Abandoned
func RandName(length int) string {

	var (
		charList []byte
		a        byte
	)

	for a = 'a'; a < 'z'; a++ {
		charList = append(charList, a)
	}
	for a = '0'; a < '9'; a++ {
		charList = append(charList, a)
	}

	var name []byte
	r := rand.New(rand.NewSource(time.Now().Unix()))
	for b := 0; b < length; b++ {
		name = append(name, charList[r.Intn(len(charList))])
	}

	return string(name)
}

func ToWorkflowTrigger(obj interface{}) *v1alpha1.WorkflowTrigger {

	wft, ok := obj.(*v1alpha1.WorkflowTrigger)
	if !ok {
		log.Errorf("I want type: WorkflowTrigger, but it is: %T", obj)
		return nil
	} else {
		return wft
	}
}

func (c *CronTrigger) Run() {

	c.WorkflowRun.Name = c.WorkflowTriggerName + "-" + strings.Replace(uuid.NewV1().String(), "-", "", -1)

	_, err := Client.CycloneV1alpha1().WorkflowRuns(c.Namespace).Create(c.WorkflowRun)
	if err != nil {
		c.FailCount++
		log.Errorf("can not create WorkflowRun: %s", err)
		return
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

func CreateCron(wft *v1alpha1.WorkflowTrigger) {

	schedule, has := getParaValue(wft.Spec.Parameters, "schedule")
	if !has {
		log.Errorf("I need schedule string(like '0 0 */2 * * *') in Parameters of WorkflowTrigger")
		return
	}

	ct := &CronTrigger{
		Namespace:           wft.Namespace,
		WorkflowTriggerName: fmt.Sprintf(KeyTemplate, wft.Namespace, wft.Name),
	}

	wfr := &v1alpha1.WorkflowRun{}

	ct.WorkflowRun = wfr

	c := cron.New()
	err := c.AddJob(schedule, ct)
	if err != nil {
		log.Errorf("can not create Cron job: %s", err)
		return
	}

	ct.Cron = c

	CronS.AddTrigger(ct)

	if wft.Spec.Enabled {
		ct.Cron.Start()
		ct.IsRun = true
	}
}

func UpdateCron(wft *v1alpha1.WorkflowTrigger) {
	DeleteCron(wft)
	CreateCron(wft)
}

func DeleteCron(wft *v1alpha1.WorkflowTrigger) {
	wftKey := fmt.Sprintf(KeyTemplate, wft.Namespace, wft.Name)
	CronS.DeleteTrigger(wftKey)
}
