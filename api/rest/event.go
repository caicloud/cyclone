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
	"net/http"

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/etcd"
	"github.com/caicloud/cyclone/pkg/log"
	"github.com/caicloud/cyclone/store"
	"github.com/emicklei/go-restful"
)

// getEvent finds a event from EventID.
//
// GET: /api/v0.1/:uid/events/{event_id}
//
// RESPONSE: (GetEventResponse)
//  {
//    "event": (object) api.Event object.
//    "error_msg": (string) set IFF the request fails.
//  }
func getEvent(request *restful.Request, response *restful.Response) {
	var getResponse api.GetEventResponse
	token := request.HeaderParameter("token")
	eventID := request.PathParameter("event_id")

	if !checkToken(token) {
		message := "Invalid token"
		log.ErrorWithFields(message, log.Fields{"event_id": eventID})
		getResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, getResponse)
		return
	}

	etcdClient := etcd.GetClient()
	sEvent, err := etcdClient.Get(EventsUnfinished + eventID)
	if err != nil {
		message := "Unable to get event from etcd"
		log.ErrorWithFields(message, log.Fields{"event_id": eventID, "error": err})
		getResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, getResponse)
		return
	}

	var event api.Event
	err = json.Unmarshal([]byte(sEvent), &event)
	if err != nil {
		message := "Unable to unmarshal event from etcd"
		log.ErrorWithFields(message, log.Fields{"event_id": eventID, "error": err})
		getResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, getResponse)
		return
	}

	getResponse.Event = event
	response.WriteHeaderAndEntity(http.StatusAccepted, getResponse)
}

// setEvent set a event, validates and saves it..
//
// PUT: /api/v0.1/:uid/events/{event_id}
//
// RESPONSE: (SetEventResponse)
//  {
//    "event_id": (string) EventID.
//    "error_msg": (string) set IFF the request fails.
//  }
func setEvent(request *restful.Request, response *restful.Response) {
	// Read out posted service information.
	setEvent := api.SetEvent{}
	err := request.ReadEntity(&setEvent)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusBadRequest, "Unable to parse request body")
		return
	}

	var setResponse api.SetEventResponse
	token := request.HeaderParameter("token")
	eventID := request.PathParameter("event_id")
	if !checkToken(token) {
		message := "Invalid token"
		log.ErrorWithFields(message, log.Fields{"event_id": eventID, "error": err})
		setResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, setResponse)
		return
	}

	etcdClient := etcd.GetClient()
	sEvent, err := etcdClient.Get(EventsUnfinished + eventID)
	if err != nil {
		message := "Unable to get event from etcd"
		log.ErrorWithFields(message, log.Fields{"event_id": eventID, "error": err})
		setResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, setResponse)
		return
	}

	var event api.Event
	err = json.Unmarshal([]byte(sEvent), &event)
	if err != nil {
		message := "Unable to unmarshal event from etcd"
		log.ErrorWithFields(message, log.Fields{"event_id": eventID, "error": err})
		setResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, setResponse)
		return
	}

	event.Service = setEvent.Event.Service
	event.Version = setEvent.Event.Version
	event.Project = setEvent.Event.Project
	event.ProjectVersion = setEvent.Event.ProjectVersion
	event.Status = setEvent.Event.Status
	event.ErrorMessage = setEvent.Event.ErrorMessage

	// Write service/version to mongo.
	ds := store.NewStore()
	defer ds.Close()

	if "" != event.Service.ServiceID && "" == event.Version.VersionID {
		ds.UpsertServiceDocument(&event.Service)
	} else if "" != event.Version.VersionID {
		ds.UpdateVersionDocument(event.Version.VersionID, setEvent.Event.Version)
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		message := "Unable to marshal event from etcd"
		log.ErrorWithFields(message, log.Fields{"event_id": eventID, "error": err})
		setResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, setResponse)
		return
	}

	log.Infof("set etcd: %s", string(eventJSON))
	err = etcdClient.Set(EventsUnfinished+eventID, string(eventJSON))
	if err != nil {
		message := "Unable to set event to etcd"
		log.ErrorWithFields(message, log.Fields{"event_id": eventID, "error": err})
		setResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, setResponse)
		return
	}

	setResponse.EventID = api.EventID(eventID)
	response.WriteHeaderAndEntity(http.StatusAccepted, setResponse)
}

// checkToken check the validity of a token.
func checkToken(token string) bool {
	// TODO Check the token.
	log.Infof("check token: %s", token)
	return true
}
