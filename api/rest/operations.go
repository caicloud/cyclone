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

package rest

import (
	"encoding/json"

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/etcd"
	"github.com/caicloud/cyclone/event"
	"github.com/caicloud/cyclone/pkg/log"
	"github.com/caicloud/cyclone/store"
)

const (
	// EventsUnfinished is the prefix for etcd.
	EventsUnfinished = "/events/unfinished/"
)

// sendCreateServiceEvent is a helper method which sends a create service event
// to etcd and wait for the event to be acked.
func sendCreateServiceEvent(service *api.Service) error {
	eventID := api.EventID(service.ServiceID)

	ds := store.NewStore()
	defer ds.Close()
	tok, _ := ds.FindtokenByUserID(service.UserID, service.Repository.SubVcs)

	event := api.Event{
		EventID:   eventID,
		Service:   *service,
		Operation: event.CreateServiceOps,
		Status:    api.EventStatusPending,
		Data:      map[string]interface{}{"Token": tok.Vsctoken.AccessToken},
	}

	log.Infof("send create service event: %v", event)

	etcdClient := etcd.GetClient()
	jsEvent, err := json.Marshal(&event)
	if err != nil {
		log.Errorf("create service event marshal err: %v", err)
		return err
	}

	err = etcdClient.Set(EventsUnfinished+string(eventID), string(jsEvent))
	if err != nil {
		log.Errorf("send create service event err: %v", err)
		return err
	}

	return nil
}

// sendCreateVersionEvent is a helper method which sends a create version event
// to etcd and wait for the event to be acked.
func sendCreateVersionEvent(service *api.Service, version *api.Version) error {
	username := service.Username
	serviceName := service.Name
	versionName := version.Name
	eventID := api.EventID(version.VersionID)

	ds := store.NewStore()
	defer ds.Close()
	tok, _ := ds.FindtokenByUserID(service.UserID, service.Repository.SubVcs)

	event := api.Event{
		EventID:   eventID,
		Service:   *service,
		Version:   *version,
		Operation: event.CreateVersionOps,
		Data: map[string]interface{}{
			"service-name": serviceName,
			"version-name": versionName,
			"username":     username,
			"Token":        tok.Vsctoken.AccessToken,
		},
		Status: api.EventStatusPending,
	}

	log.Infof("send create version event: %v", event)

	etcdClient := etcd.GetClient()
	jsEvent, err := json.Marshal(&event)
	if err != nil {
		log.Errorf("create version event marshal err: %v", err)
		return err
	}

	err = etcdClient.Set(EventsUnfinished+string(eventID), string(jsEvent))
	if err != nil {
		log.Errorf("send create version event err: %v", err)
		return err
	}

	return nil
}
