package workflowTrigger

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/workflow/controller"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
)

const (
	KeyTemplate = "%s-%s"
)

type CronTrigger struct {
	Cron               *cron.Cron
	IsRun              bool
	SuccCount          int
	FailCount          int
	WorkflowTriggerKey string
	WorkflowRun        *v1alpha1.WorkflowRun
}

var (
	CronTriggerMap = make(map[string]*CronTrigger)
)

// [a-z0-9]{5}
// exp: ab12c
func RandName(length int) (string) {

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

func ToWorkflowTrigger(obj interface{}) (*v1alpha1.WorkflowTrigger) {

	wft, ok := obj.(*v1alpha1.WorkflowTrigger)
	if !ok {
		log.Errorf("I want type: WorkflowTrigger, but it is: %T", obj)
		return nil
	} else {
		return wft
	}
}

func (c *CronTrigger) Run() {

	c.WorkflowRun.Name = c.WorkflowTriggerKey + "-" + RandName(5)

	_, err := controller.Client.CycloneV1alpha1().WorkflowRuns("default").Create(c.WorkflowRun)
	if err != nil {
		c.FailCount++
		log.Errorf("can not create WorkflowRun: %s", err)
		return
	} else {
		c.SuccCount++
	}
}

func getParaValue(items []v1alpha1.ParameterItem, key string) (value string) {
	for i := 0; i < len(items); i++ {
		if items[i].Name == key {
			value = items[i].Value
			return
		}
	}
	return
}

func CreateCron(wft *v1alpha1.WorkflowTrigger) {

	ct := &CronTrigger{
		WorkflowTriggerKey: fmt.Sprintf(KeyTemplate, wft.Namespace, wft.Name),
	}

	wfr := &v1alpha1.WorkflowRun{}

	ct.WorkflowRun = wfr

	c := cron.New()
	err := c.AddJob(getParaValue(wft.Spec.Parameters, "schedule"), ct)
	if err != nil {
		log.Errorf("can not create Cron job: %s", err)
		return
	}

	ct.Cron = c
	CronTriggerMap[ct.WorkflowTriggerKey] = ct

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
	ct, ok := CronTriggerMap[wftKey]
	if ok {
		ct.Cron.Stop()
		ct.IsRun = false
		delete(CronTriggerMap, wftKey)
	} else {
		log.Errorf("can not find the cron object related with the WorkflowTrigger: %s", wftKey)
	}
}
