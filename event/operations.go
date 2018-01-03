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
	"encoding/json"

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/etcd"
	log "github.com/zoumo/logdog"

	"github.com/caicloud/cyclone/store"
)

// SendCreateServiceEvent is a helper method which sends a create service event
// to etcd and wait for the event to be acked.
func SendCreateServiceEvent(service *api.Service) error {
	eventID := api.EventID(service.ServiceID)

	ds := store.NewStore()
	defer ds.Close()

	// tok, _ := ds.FindtokenByUserID(service.UserID, service.Repository.SubVcs)
	project, err := ds.FindProjectByServiceID(service.ServiceID)
	if err != nil {
		log.Errorf("fail to get token for service %s", service.Name)
		return err
	}

	event := api.Event{
		EventID:   eventID,
		Service:   *service,
		Operation: CreateServiceOps,
		Status:    api.EventStatusPending,
		Data:      map[string]interface{}{"Token": project.SCM.Token},
	}

	// Set the namespace for worker in k8s cloud.
	if project.Worker != nil && len(project.Worker.Namespace) != 0 {
		event.Data["namespace"] = project.Worker.Namespace
	}

	log.Infof("send create service event: %v", event)

	ds.CreateMassage(&event)

	return nil
}

// SendCreateVersionEvent is a helper method which sends a create version event
// to etcd and wait for the event to be acked.
func SendCreateVersionEvent(service *api.Service, version *api.Version) error {
	username := service.Username
	serviceName := service.Name
	versionName := version.Name
	eventID := api.EventID(version.VersionID)

	ds := store.NewStore()
	defer ds.Close()

	// tok, _ := ds.FindtokenByUserID(service.UserID, service.Repository.SubVcs)
	project, err := ds.FindProjectByServiceID(service.ServiceID)
	if err != nil {
		log.Errorf("fail to get token for service %s", serviceName)
		return err
	}

	event := api.Event{
		EventID:   eventID,
		Service:   *service,
		Version:   *version,
		Operation: CreateVersionOps,
		Data: map[string]interface{}{
			"service-name": serviceName,
			"version-name": versionName,
			"username":     username,
			"Token":        project.SCM.Token,
		},
		Status: api.EventStatusPending,
	}

	// Set the namespace for worker in k8s cloud.
	if project.Worker != nil && len(project.Worker.Namespace) != 0 {
		event.Data["namespace"] = project.Worker.Namespace
	}

	log.Infof("send create version event: %v", event)

	ds.CreateMassage(&event)

	return nil
}
