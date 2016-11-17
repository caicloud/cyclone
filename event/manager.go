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

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/etcd"
	"github.com/caicloud/cyclone/pkg/log"
	"github.com/caicloud/cyclone/remote"
	"github.com/caicloud/cyclone/resource"
	"golang.org/x/net/context"
)

const (
	// unfinished event dir path in etcd
	Events_Unfinished = "/events/unfinished"
)

var (
	// cert path of worker
	certPathWorker string

	// registry config of worker
	registryWorker api.RegistryCompose

	// remote api manager
	remoteManager *remote.Manager

	resourceManager *resource.Manager
)

// Init init event manager
// Step1: init event operation map
// Step2: new a etcd client
// Step3: load unfinished events from etcd
// Step4: create a unfinished events watcher
// Step5: new a remote api manager
func Init(certPath string, registry api.RegistryCompose) {
	certPathWorker = certPath
	registryWorker = registry

	initOperationMap()

	etcdClient := etcd.GetClient()

	if !etcdClient.IsDirExist(Events_Unfinished) {
		err := etcdClient.CreateDir(Events_Unfinished)
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
	resourceManager = resource.NewManager()
}

// watchEtcd watch unfinished events status change in etcd
func watchEtcd(etcdClient *etcd.Client) {
	watcherUnfinishedEvents, err := etcdClient.CreateWatcher(Events_Unfinished)
	if err != nil {
		log.Fatalf("watch unfinshed events err: %v", err)
	}

	for {
		change, err := watcherUnfinishedEvents.Next(context.Background())
		if err != nil {
			log.Fatalf("watch unfinshed events next err: %v", err)
		}

		switch change.Action {
		case etcd.Watch_Action_Create:
			log.Infof("watch unfinshed events create: %q\n", change.Node)
			event, err := loadEventFromJSON(change.Node.Value)
			if err != nil {
				log.Errorf("analysis create event err: %v", err)
				continue
			}
			eventCreateHandler(&event)

		case etcd.Watch_Action_Set:
			log.Infof("watch unfinshed events set: %q\n", change.Node.Value)
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

		case etcd.Watch_Action_Delete:
			log.Infof("watch finshed events delete: %q\n", change.PrevNode)
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
		err := etcdClient.Delete(Events_Unfinished + "/" + string(event.EventID))
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
	return ec.Set(Events_Unfinished+"/"+string(event.EventID), string(bEvent))
}

// LoadEventFromEtcd loads event from etcd.
func LoadEventFromEtcd(eventID api.EventID) (*api.Event, error) {
	ec := etcd.GetClient()
	sEvent, err := ec.Get(Events_Unfinished + "/" + string(eventID))
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
	jsonEvents, err := etcd.List(Events_Unfinished)
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
			go CheckWorkerTimeOut(event)
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
			if err == resource.ErrUnableSupport {
				log.Info("Waiting for resource to be relaesed...")
				time.Sleep(time.Second * 10)
				continue
			}
			// worker busy
			if err == ErrWorkerBusy {
				log.Info("All system worker are busy, wait for 10 seconds")
				time.Sleep(time.Second * 10)
				continue
			}

			// remove the event from queue which had run
			pendingEvents.Out()

			event.Status = api.EventStatusFail
			event.ErrorMessage = err.Error()
			log.ErrorWithFields("handle event err", log.Fields{"error": err, "event": event})
			postHookEvent(&event)
			etcdClient := etcd.GetClient()
			err := etcdClient.Delete(Events_Unfinished + "/" + string(event.EventID))
			if err != nil {
				log.Errorf("delete finished event err: %v", err)
			}
			continue
		}

		// remove the event from queue which had run
		pendingEvents.Out()
		event.Status = api.EventStatusRunning
		SaveEventToEtcd(&event)
	}
}
