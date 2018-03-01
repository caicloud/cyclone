/*
Copyright 2017 caicloud authors. All rights reserved.

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

package manager

import (
	"fmt"

	log "github.com/golang/glog"

	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/event"
	"github.com/caicloud/cyclone/pkg/store"
)

// EventManager represents the interface to manage event.
type EventManager interface {
	GetEvent(id string) (*api.Event, error)
	SetEvent(event *api.Event) (*api.Event, error)
}

// eventManager represents the manager for event.
type eventManager struct {
	ds *store.DataStore
}

// NewEventManager creates a event manager.
func NewEventManager(dataStore *store.DataStore) (EventManager, error) {
	if dataStore == nil {
		return nil, fmt.Errorf("Fail to new event manager as data store is nil")
	}

	return &eventManager{dataStore}, nil
}

func (m *eventManager) GetEvent(id string) (*api.Event, error) {
	event, err := m.ds.GetEventByID(id)
	if err != nil {
		log.Errorf("Fail to get the event %s", id)
		return nil, err
	}

	return event, nil
}

func (m *eventManager) SetEvent(e *api.Event) (*api.Event, error) {
	err := m.ds.UpdatePipelineRecord(e.PipelineRecord)
	if err != nil {
		log.Errorf("Fail to set the pipeline record %s", e.PipelineRecord.ID)
		return nil, err
	}

	event.UpdateEvent(e)
	return e, nil
}
