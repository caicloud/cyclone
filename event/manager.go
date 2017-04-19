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
	"container/list"
	"encoding/json"
	"sync"
	"time"

	"k8s.io/client-go/rest"

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/cloud"
	"github.com/caicloud/cyclone/etcd"
	"github.com/caicloud/cyclone/remote"
	"github.com/caicloud/cyclone/store"
	"github.com/zoumo/logdog"
	log "github.com/zoumo/logdog"
	"golang.org/x/net/context"
)

const (
	// EventsUnfinished represents unfinished event dir path in etcd
	EventsUnfinished = "/events/unfinished/"
)

var (
	CloudController *cloud.Controller

	workerOptions *cloud.WorkerOptions

	// remote api manager
	remoteManager *remote.Manager
)

// Init init event manager
// Step1: init event operation map
// Step2: new a etcd client
// Step3: load unfinished events from etcd
// Step4: create a unfinished events watcher
// Step5: new a remote api manager
func Init(wopts *cloud.WorkerOptions, cloudAutoDiscovery bool) {

	initCloudController(wopts, cloudAutoDiscovery)

	initOperationMap()

	etcdClient := etcd.GetClient()

	if !etcdClient.IsDirExist(EventsUnfinished) {
		err := etcdClient.CreateDir(EventsUnfinished)
		if err != nil {
			log.Errorf("init event manager create events dir err: %v", err)
			return
		}
	}

	GetList().loadListFromEtcd(etcdClient)
	initPendingQueue()

	go watchEtcd(etcdClient)
	go handlePendingEvents()

	remoteManager = remote.NewManager()
}

// FIXME, so ugly
// load clouds from database
func initCloudController(wopts *cloud.WorkerOptions, cloudAutoDiscovery bool) {
	CloudController = cloud.NewController()
	// load clouds from store
	ds := store.NewStore()
	defer ds.Close()
	clouds, err := ds.FindAllClouds()
	if err != nil {
		logdog.Error("Can not find clouds from ds", logdog.Fields{"err": err})
		return
	}

	CloudController.AddClouds(clouds...)

	if len(CloudController.Clouds) == 0 && cloudAutoDiscovery {
		addInClusterK8SCloud()
	}

	workerOptions = wopts
}

func addInClusterK8SCloud() {
	ds := store.NewStore()
	defer ds.Close()
	_, err := rest.InClusterConfig()
	if err == nil {
		// in k8s cluster
		opt := cloud.Options{
			Kind:         cloud.KindK8SCloud,
			Name:         "_inCluster",
			K8SInCluster: true,
		}
		err := CloudController.AddClouds(opt)
		if err != nil {
			logdog.Warn("Can not add inCluster k8s cloud to database")
		}
		err = ds.InsertCloud(&opt)
		if err != nil {
			logdog.Warn("Can not add inCluster k8s cloud to database")
		}
	}
}

// watchEtcd watch unfinished events status change in etcd
func watchEtcd(etcdClient *etcd.Client) {
	watcherUnfinishedEvents, err := etcdClient.CreateWatcher(EventsUnfinished)
	if err != nil {
		log.Fatalf("watch unfinshed events err: %v", err)
	}

	for {
		change, err := watcherUnfinishedEvents.Next(context.Background())
		if err != nil {
			log.Fatalf("watch unfinshed events next err: %v", err)
		}

		switch change.Action {
		case etcd.WatchActionCreate:
			log.Infof("watch unfinshed events create: %s\n", change.Node)
			event, err := loadEventFromJSON(change.Node.Value)
			if err != nil {
				log.Errorf("analysis create event err: %v", err)
				continue
			}
			eventCreateHandler(&event)

		case etcd.WatchActionSet:
			log.Infof("watch unfinshed events set: %s\n", change.Node.Value)
			event, err := loadEventFromJSON(change.Node.Value)
			if err != nil {
				log.Errorf("analysis set event err: %v", err)
				continue
			}

			if change.PrevNode == nil {
				eventCreateHandler(&event)
			} else {
				preEvent, preErr := loadEventFromJSON(change.PrevNode.Value)
				if preErr != nil {
					log.Errorf("analysis set pre event err: %v", preErr)
					continue
				}
				eventChangeHandler(&event, &preEvent)
			}

		case etcd.WatchActionDelete:
			log.Infof("watch finshed events delete: %s\n", change.PrevNode)
			event, err := loadEventFromJSON(change.PrevNode.Value)
			if err != nil {
				log.Errorf("analysis delete event err: %v", err)
				continue
			}
			eventRemoveHandler(&event)

		default:
			log.Warnf("watch unknow etcd action(%s): %v", change.Action, change)
		}
	}
}

// eventCreateHandler handler when watched a event created
func eventCreateHandler(event *api.Event) {
	GetList().addUnfinshedEvent(event)
	pendingEvents.In(event)

	return
}

// eventChangeHandler handler when watched a event changed
func eventChangeHandler(event *api.Event, preEvent *api.Event) {
	GetList().addUnfinshedEvent(event)

	// event handle finished
	if !IsEventFinished(preEvent) && IsEventFinished(event) {
		postHookEvent(event)
		etcdClient := etcd.GetClient()
		err := etcdClient.Delete(EventsUnfinished + string(event.EventID))
		if err != nil {
			log.Errorf("delete finished event err: %v", err)
		}
	}
}

// SaveEventToEtcd saves event in etcd.
func SaveEventToEtcd(event *api.Event) error {
	ec := etcd.GetClient()
	bEvent, err := json.Marshal(event)
	if nil != err {
		return err
	}
	return ec.Set(EventsUnfinished+string(event.EventID), string(bEvent))
}

// LoadEventFromEtcd loads event from etcd.
func LoadEventFromEtcd(eventID api.EventID) (*api.Event, error) {
	ec := etcd.GetClient()
	sEvent, err := ec.Get(EventsUnfinished + string(eventID))
	if err != nil {
		return nil, err
	}
	var event api.Event
	err = json.Unmarshal([]byte(sEvent), &event)
	if nil != err {
		return nil, err
	}
	return &event, nil
}

// eventChangeHandler handler when watched a event removed
func eventRemoveHandler(event *api.Event) {
	GetList().removeEvent(event.EventID)
}

// IsEventFinished return event is finished
func IsEventFinished(event *api.Event) bool {
	if event.Status == api.EventStatusSuccess ||
		event.Status == api.EventStatusFail ||
		event.Status == api.EventStatusCancel {
		return true
	}
	return false
}

// Delete me
// eventList load unfinished events
var eventList *List

// List define event list
type List struct {
	sync.RWMutex
	events map[api.EventID]*api.Event
}

// GetList get event list
func GetList() *List {
	if eventList == nil {
		eventList = newList()
	}
	return eventList
}

// newList new a event list
func newList() *List {
	return &List{
		events: make(map[api.EventID]*api.Event),
	}
}

// loadListFromEtcd load event list from etcd
func (el *List) loadListFromEtcd(etcd *etcd.Client) {
	jsonEvents, err := etcd.List(EventsUnfinished)
	if err != nil {
		log.Errorf("load event list from etcd err: %v", err)
		return
	}

	for _, jsonEvent := range jsonEvents {
		event, err := loadEventFromJSON(jsonEvent)
		if err != nil {
			log.Errorf("load event from etcd err: %v", err)
			continue
		}
		log.Infof("load event to list: %s", event.EventID)
		if event.Status == api.EventStatusPending {
			el.addUnfinshedEvent(&event)
		} else {
			go CheckWorkerTimeout(&event)
		}
	}
}

// loadEventFromJSON load event info from Json format.
func loadEventFromJSON(jsonEvent string) (api.Event, error) {
	var event api.Event
	err := json.Unmarshal([]byte(jsonEvent), &event)
	if err != nil {
		log.Errorf("load event list from etcd err: %v", err)
		return event, err
	}

	return event, nil
}

// addUnfinshedEvent adds unfinished event to list.
func (el *List) addUnfinshedEvent(event *api.Event) {
	if event.EventID != "" {
		el.Lock()
		defer el.Unlock()
		el.events[event.EventID] = event
	}
}

// removeEvent removes a event form list.
func (el *List) removeEvent(eventID api.EventID) (err error) {
	if eventID != "" {
		el.Lock()
		defer el.Unlock()
		delete(el.events, eventID)
	}
	return err
}

// GetEvent gets a event from list.
func (el *List) GetEvent(eventID api.EventID) *api.Event {
	el.RLock()
	defer el.RUnlock()
	event, ok := el.events[eventID]
	if !ok {
		return nil
	}
	return event
}

var pendingEvents Queue

// Queue is the type for pending events.
type Queue struct {
	sync.RWMutex
	queue *list.List
}

// Init initializes a queue.
func (eq *Queue) Init() {
	eq.queue = list.New()
}

// In enqueues a event.
func (eq *Queue) In(event *api.Event) {
	eq.Lock()
	defer eq.Unlock()

	eq.queue.PushBack(event)
}

// GetFront get the first event in the queue.
func (eq *Queue) GetFront() *api.Event {
	eq.RLock()
	defer eq.RUnlock()

	element := eq.queue.Front()
	return element.Value.(*api.Event)
}

// Out dequeues a event.
func (eq *Queue) Out() {
	eq.Lock()
	defer eq.Unlock()

	element := eq.queue.Front()
	eq.queue.Remove(element)
}

// IsEmpty checks if the queue is empty.
func (eq *Queue) IsEmpty() bool {
	eq.RLock()
	defer eq.RUnlock()

	return eq.queue.Len() == 0
}

// initPendingQueue initializes the pending events queue.
func initPendingQueue() {
	pendingEvents.Init()

	eventList.RLock()
	defer eventList.RUnlock()

	for _, event := range eventList.events {
		if event.Status == api.EventStatusPending {
			pendingEvents.In(event)
		}
	}
}

// handlePendingEvents polls event queues and handle events one by one.
func handlePendingEvents() {
	for {
		if pendingEvents.IsEmpty() {
			time.Sleep(time.Second * 1)
			continue
		}

		event := *pendingEvents.GetFront()
		err := handleEvent(&event)
		if err != nil {
			// if err == resource.ErrUnableSupport {
			// 	log.Info("Waiting for resource to be relaesed...")
			// 	time.Sleep(time.Second * 10)
			// 	continue
			// }
			// worker busy
			// if err == ErrWorkerBusy {
			// 	log.Info("All system worker are busy, wait for 10 seconds")
			// 	time.Sleep(time.Second * 10)
			// 	continue
			// }

			if cloud.IsAllCloudsBusyErr(err) {
				log.Info("All system worker are busy, wait for 10 seconds")
				time.Sleep(time.Second * 10)
				continue
			}

			// remove the event from queue which had run
			pendingEvents.Out()

			event.Status = api.EventStatusFail
			event.ErrorMessage = err.Error()
			log.Error("handle event err", log.Fields{"error": err, "event": event})
			postHookEvent(&event)
			etcdClient := etcd.GetClient()
			err := etcdClient.Delete(EventsUnfinished + string(event.EventID))
			if err != nil {
				log.Errorf("delete finished event err: %v", err)
			}
			continue
		}

		// remove the event from queue which had run
		pendingEvents.Out()
		event.Status = api.EventStatusRunning
		ds := store.NewStore()
		defer ds.Close()
		if event.Operation == "create-version" {
			event.Version.Status = api.VersionRunning
			if err := ds.UpdateVersionDocument(event.Version.VersionID, event.Version); err != nil {
				log.Errorf("Unable to update version status post hook for %+v: %v", event.Version, err)
			}
		}
		SaveEventToEtcd(&event)
	}
}

// CheckWorkerTimeout ...
func CheckWorkerTimeout(event *api.Event) {
	if IsEventFinished(event) {
		return
	}
	worker, err := CloudController.LoadWorker(event.Worker)
	if err != nil {
		log.Error("load worker error")
		return
	}

	ok, left := worker.IsTimeout()
	if ok {
		event.Status = api.EventStatusFail
		SaveEventToEtcd(event)
		return
	}

	time.Sleep(left)

	event, err = LoadEventFromEtcd(event.EventID)
	if err != nil {
		return
	}

	if !IsEventFinished(event) {
		log.Infof("event time out: %v", event)
		event.Status = api.EventStatusCancel
		SaveEventToEtcd(event)
	}
}
