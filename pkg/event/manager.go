/*
Copyright 2016 caicloud authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"text/template"
	"time"

	"github.com/caicloud/nirvana/log"
	"github.com/zoumo/logdog"
	"gopkg.in/mgo.v2"

	"github.com/caicloud/cyclone/cmd/worker/options"
	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/cloud"
	"github.com/caicloud/cyclone/pkg/scm"
	"github.com/caicloud/cyclone/pkg/store"
	regexutil "github.com/caicloud/cyclone/pkg/util/regex"
)

// maxRetry represents the max number of retry for event when cloud is busy.
// TODO(robin) Make this configurable.
const maxRetry = 60

var (
	defaultWorkerOptions *options.WorkerOptions
	notificationURL      string
	recordWebURLTemplate string
)

// Init init event manager
// Step1: new event manager
// Step2: create a goroutine to watch events
func Init(opts *options.WorkerOptions, notifyURL, recordURLTemplate string) {
	ds := store.NewStore()
	em := NewEventManager(ds)

	defaultWorkerOptions = opts
	notificationURL = notifyURL
	recordWebURLTemplate = recordURLTemplate
	go em.WatchEvent()
}

// EventManager represents the manager of events.
type EventManager interface {
	HandleEvent(event *api.Event) error
	WatchEvent()
}

type eventManager struct {
	ds *store.DataStore
}

// NewEventManager creates the event manager.
func NewEventManager(ds *store.DataStore) EventManager {
	return &eventManager{
		ds: ds,
	}
}

// HandleEvent handles the event.
func (em *eventManager) HandleEvent(event *api.Event) error {
	eventID := event.ID
	pipelineRecord := event.PipelineRecord
	err := createWorkerForEvent(event)
	if err != nil {
		if cloud.IsAllCloudsBusyErr(err) && event.Retry < maxRetry {
			log.Infof("All system worker are busy, wait for 10 seconds, event id:%v retry times:%v",
				event.ID, event.Retry)
			time.Sleep(time.Second * 10)
			em.ds.ResetEvent(event)
			return nil
		}

		pipelineRecord.Status = api.Failed
		pipelineRecord.ErrorMessage = err.Error()
		log.Errorf("handle event %s as err: %s", eventID, err.Error())
		postHookEvent(event)

		if err = em.ds.DeleteEvent(eventID); err != nil {
			log.Errorf("fail to delete event %s as err: %s", eventID, err.Error())
		}
	} else {
		event.QueueStatus = api.Handling
		pipelineRecord.Status = api.Running
		pipelineRecord.StartTime = time.Now()
		if err = em.ds.UpdateEvent(event); err != nil {
			log.Errorf("fail to update event %s as err: %s", eventID, err.Error())
			log.Infof("terminate the worker for event %s", eventID)
			// terminate the worker pod while updated event failed(e.g. event Not Found, deleted by deleting records).
			terminateEventWorker(event)
			return err
		}

		// start to run pipeline. if trigger is webhook pr, update SCM statuses.
		errScm := sendScmStatuses(event)
		if errScm != nil {
			log.Warningf("fail to send scm statuses:%v", errScm)
		}
	}

	if err := em.ds.UpdatePipelineRecord(pipelineRecord); err != nil {
		log.Errorf("fail to update the pipeline record %s: %v", event.ID, err)
		return err
	}

	return nil
}

// WatchEvent watches the event queue, and handles the events.
func (em *eventManager) WatchEvent() {
	for {
		time.Sleep(1 * time.Second)

		event, err := em.ds.NextEvent()
		if err != nil {
			if err != mgo.ErrNotFound {
				log.Errorf("fail to get the event as %s", err.Error())
			}
			continue
		}

		go em.HandleEvent(event)
	}
}

func createWorkerForEvent(event *api.Event) error {
	log.Infof("Create worker for event %s", event.ID)
	workerCfg := event.Project.Worker
	buildInfo := event.Pipeline.Build.BuildInfo
	performParams := event.PipelineRecord.PerformParams
	workerInfo := &api.WorkerInfo{
		Name: "cyclone-worker-" + event.ID,
	}

	if workerCfg != nil {
		// Only use cache when:
		// * the project has enable caches
		// * the pipeline has enable cache and set the build tool
		// * the perform params has enable cache
		// * the project has enable the cache for the build tool of the pipeline
		if buildInfo != nil && buildInfo.CacheDependency && buildInfo.BuildTool != nil && performParams.CacheDependency {
			tool := buildInfo.BuildTool.Name
			if cache, ok := workerCfg.DependencyCaches[tool]; ok {
				switch tool {
				case api.MavenBuildTool:
					workerInfo.MountPath = "/root/.m2"
				case api.NPMBuildTool:
					workerInfo.MountPath = "/root/.npm"
				case api.GradleBuildTool:
					workerInfo.MountPath = "/root/.gradle"
				default:
					// Just log error and let the pipeline to run in non-cache mode.
					return fmt.Errorf("Build tool %s is not supported, only supports: %s, %s, %s",
						tool, api.MavenBuildTool, api.NPMBuildTool, api.GradleBuildTool)
				}

				workerInfo.CacheVolume = cache.Name
			}
		}
	}

	cp, err := getCloudProvider(workerCfg)
	if err != nil {
		log.Error(err)
		return err
	}

	workerInfo.Quota = getWorkerQuota(defaultWorkerOptions.Quota, workerCfg.Quota)

	opts := defaultWorkerOptions
	opts.EventID = event.ID
	opts.ProjectName = event.Project.Name
	opts.PipelineName = event.Pipeline.Name
	workerInfo, err = cp.Provision(workerInfo, opts)
	if err != nil {
		return err
	}

	event.WorkerInfo = workerInfo

	// err = worker.Do()
	// if err != nil {
	// 	return err
	// }

	// update worker info to event
	// event.Worker = worker.GetWorkerInfo()

	// log.Infof("Worker info: %v", event.Worker)

	// trigger after get an valid worker
	// triggerHooks(event, PostStartPhase)

	// No need to call UpdateEvent() here, as the callers will update the event status after call createWorkerForEvent().
	// UpdateEvent(event)
	go CheckWorkerTimeout(event)

	return nil
}

func getWorkerQuota(defaultQuota options.Quota, workerQuota *api.WorkerQuota) options.Quota {
	quota := defaultQuota.DeepCopy()

	if workerQuota != nil {
		if workerQuota.LimitsCPU != "" {
			quota[options.ResourceLimitsCPU].Set(workerQuota.LimitsCPU)
		}

		if workerQuota.LimitsMemory != "" {
			quota[options.ResourceLimitsMemory].Set(workerQuota.LimitsMemory)
		}

		if workerQuota.RequestsCPU != "" {
			quota[options.ResourceRequestsCPU].Set(workerQuota.RequestsCPU)
		}

		if workerQuota.RequestsMemory != "" {
			quota[options.ResourceRequestsMemory].Set(workerQuota.RequestsMemory)
		}
	}

	return quota
}

func postHookEvent(event *api.Event) {
	log.Info("posthook of event")
	log.Infof("worker： %s", event.Project.Worker)
	terminateEventWorker(event)

	errScm := sendScmStatuses(event)
	if errScm != nil {
		log.Warningf("fail to send scm statuses:%v", errScm)
	}

	pipeline := event.Pipeline
	record := event.PipelineRecord

	// post hooks
	if pipeline != nil && pipeline.Notification != nil {

		policy := pipeline.Notification.Policy

		sendFlag := false
		if policy == api.AlwaysNotify {
			sendFlag = true
		} else {
			if policy == api.SuccessNotify && record.Status == api.Success {
				sendFlag = true
			}
			if policy == api.FailureNotify && record.Status != api.Success {
				sendFlag = true
			}
		}

		if sendFlag {
			log.Infof("start to send notification for %v/%v/%v", event.Project.Name, pipeline.Name, record.Name)

			webURL, err := getPipelineRecordWebURL(recordWebURLTemplate, event)
			if err != nil {
				log.Errorf("Fail to get web url for %v/%v/%v as %s",
					event.Project.Name, pipeline.Name, record.Name, err.Error())
				return
			}

			content := &api.NotificationContent{
				ProjectName:      event.Project.Name,
				PipelineName:     pipeline.Name,
				RecordName:       record.Name,
				RecordID:         record.ID,
				Trigger:          record.Trigger,
				Status:           record.Status,
				ErrorMessage:     record.ErrorMessage,
				PipelinRecordURL: webURL,
				StartTime:        record.StartTime,
				EndTime:          record.EndTime,
			}

			err = sendNotification(content)
			if err != nil {
				log.Errorf("Fail to send notification for %v/%v/%v as %s",
					event.Project.Name, pipeline.Name, record.Name, err.Error())
			}
		}

	}
}

func sendNotification(content *api.NotificationContent) error {
	bodyBytes, err := json.Marshal(content)
	if err != nil {
		return err
	}
	reqBody := bytes.NewReader(bodyBytes)

	req, err := http.NewRequest(http.MethodPost, notificationURL, reqBody)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode/100 == 2 {
		return nil
	}

	err = fmt.Errorf("Fail to send notification as %s, response code:%v", body, resp.StatusCode)
	return err
}

func sendScmStatuses(event *api.Event) error {
	// If this event is not related to GitHub PR or GitLab MR, will return.
	_, isGitlabMR := regexutil.GetGitlabMRID(event.PipelineRecord.PerformParams.Ref, false)
	_, isGithubPR := regexutil.GetGithubPRID(event.PipelineRecord.PerformParams.Ref)
	if !isGitlabMR && !isGithubPR {
		return nil
	}

	p, err := scm.GetSCMProvider(event.Project.SCM)
	if err != nil {
		return err
	}

	targetURL, err := getPipelineRecordWebURL(recordWebURLTemplate, event)
	if err != nil {
		return err
	}

	gitSource, err := api.GetGitSource(event.Pipeline.Build.Stages.CodeCheckout.MainRepo)
	if err != nil {
		return err
	}

	err = p.CreateStatus(event.PipelineRecord.Status, targetURL, gitSource.Url, event.PipelineRecord.PRLastCommitSHA)
	if err != nil {
		return err
	}
	return nil
}

func getPipelineRecordWebURL(urlTemplate string, event *api.Event) (string, error) {
	// Create a new template and parse the url template into it.
	t := template.Must(template.New("target url template").Parse(urlTemplate))

	var buf bytes.Buffer
	// Execute the template.
	err := t.Execute(&buf, event)
	if err != nil {
		return "", err
	}

	return buf.String(), err

}

// UpdateEvent updates the event. If it is finished, delete it and trigger the post hook.
//func (em *eventManager) UpdateEvent(event *api.Event) error {
func UpdateEvent(event *api.Event) error {
	eventID := event.ID
	ds := store.NewStore()
	defer ds.Close()

	if IsEventFinished(event) {
		log.Infof("Event %s has finished", eventID)
		if err := ds.DeleteEvent(string(eventID)); err != nil {
			log.Errorf("fail to delete the event %s", eventID)
			return err
		}
		event.PipelineRecord.EndTime = time.Now()
		postHookEvent(event)
	} else {
		if err := ds.UpdateEvent(event); err != nil {
			log.Errorf("fail to update the event %s", eventID)
			return err
		}
	}

	err := ds.UpdatePipelineRecord(event.PipelineRecord)
	if err != nil {
		log.Errorf("Fail to update the pipeline record %s", event.PipelineRecord.ID)
		return err
	}

	return nil
}

// IsEventFinished return true if event is finished.
func IsEventFinished(event *api.Event) bool {
	status := event.PipelineRecord.Status
	if status == api.Success ||
		status == api.Failed ||
		status == api.Aborted {
		return true
	}
	return false
}

// DeleteEventByRecordID deletes the event. If it is running, delete its worker at the same time.
func DeleteEventByRecordID(id string) error {
	ds := store.NewStore()
	defer ds.Close()

	event, err := ds.GetEventByRecordID(id)
	if err != nil {
		log.Errorf("fail to get the event by record %s as %s", id, err.Error())
		return err
	}

	// Delete the related worker pod
	log.Infof("terminate the worker for event %s", id)
	terminateEventWorker(event)

	// Delete the event in event queue.
	if err := ds.DeleteEvent(id); err != nil {
		log.Errorf("fail to delete the event %s", id)
		return err
	}

	return nil
}

// GetEvent get the event by ID.
func GetEvent(id string) (*api.Event, error) {
	ds := store.NewStore()
	defer ds.Close()

	return ds.GetEventByID(id)
}

// CheckWorkerTimeout ...
func CheckWorkerTimeout(event *api.Event) {
	if IsEventFinished(event) {
		return
	}
	// worker, err := CloudController.LoadWorker(event.Worker)
	// if err != nil {
	// 	log.Errorf("load worker %v with error: %v", event.Worker, err)
	// 	return
	// }

	pipelineRecord := event.PipelineRecord
	now := time.Now()
	workerInfo := event.WorkerInfo
	if now.After(workerInfo.DueTime) {
		pipelineRecord.Status = api.Failed
		err := UpdateEvent(event)
		if err != nil {
			log.Errorf("update event %s error: %v", event.ID, err)
		}
		return
	}

	time.Sleep(workerInfo.DueTime.Sub(now))

	ds := store.NewStore()
	defer ds.Close()
	e, err := ds.GetEventByID(event.ID)
	if err != nil {
		return
	}

	if !IsEventFinished(e) {
		e.PipelineRecord.Status = api.Aborted
		err = UpdateEvent(e)
		if err != nil {
			log.Errorf("update event %s error: %v", e.ID, err)
		}
	}
}

func terminateEventWorker(event *api.Event) {
	cp, err := getCloudProvider(event.Project.Worker)
	if err != nil {
		logdog.Warnf("load worker %v with error: %v", event.WorkerInfo, err)
	} else {
		if event.WorkerInfo != nil {
			err = cp.TerminateWorker(event.WorkerInfo.Name)
			if err != nil {
				logdog.Warnf("Terminate worker err: %v", err)
			}
		}
	}
}

func getCloudProvider(w *api.WorkerConfig) (cloud.Provider, error) {

	log.Infof("worker: %+v", w)

	ds := store.NewStore()
	defer ds.Close()

	// Use the incluster cloud by default.
	var cloudName string

	if w != nil && w.Location != nil {
		cloudName = w.Location.CloudName
	} else {
		cloudName = cloud.DefaultCloudName
		log.Warningf("No cloud specified for this project, will use the defult cloud %s", cloudName)
	}

	c, err := ds.FindCloudByName(cloudName)
	if err != nil {
		err = fmt.Errorf("fail to find cloud %s as %v", cloudName, err)
		log.Error(err)
		return nil, err
	}

	if c.Kubernetes != nil {
		if c.Kubernetes.InCluster {
			// default cluster, get default namespace.
			c.Kubernetes.Namespace = cloud.DefaultNamespace
		} else {
			c.Kubernetes.Namespace = w.Location.Namespace
		}
	}

	return cloud.NewCloudProvider(c)
}
