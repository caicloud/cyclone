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

package store

import (
	"fmt"
	"time"

	"github.com/caicloud/cyclone/api"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	// retryThreshold represents the threshold for retry time of event.
	// If the retry time is not more than this threshold, the retry interval will be retry * retryInterval.
	// the retry interval will be retryThreshold * retryInterval.
	retryThreshold = 10

	// retryInterval represents the interval unit for retry.
	retryInterval = 1 * time.Minute
)

var (
	// outTimeout represents the timeout for event to keep "out" status.
	// After the event is out, then its status will be "handling" after it starts to be handled,
	// otherwise, it will be re-enqueued and its status will change back to "in".
	outTimeout, _ = time.ParseDuration("-3m")
)

// CreateEvent creates the event with the initial status `in`.
func (d *DataStore) CreateEvent(event *api.Event) (*api.Event, error) {
	event.QueueStatus = api.InQueue
	event.InTime = time.Now()

	if err := d.eventCollection.Insert(event); err != nil {
		return nil, err
	}

	return event, nil
}

// GetEventByID gets the event by id.
func (d *DataStore) GetEventByID(id string) (*api.Event, error) {
	event := &api.Event{}
	query := bson.M{"event_id": id}
	if err := d.eventCollection.Find(query).One(event); err != nil {
		return nil, err
	}

	return event, nil
}

// NextEvent get the next event with conditions:
// * sorted by `inTime`, first in first out;
// * with status status `in`;
// * has been out more than the threshold.
func (d *DataStore) NextEvent() (*api.Event, error) {
	query := bson.M{"$or": []bson.M{bson.M{"queueStatus": api.InQueue}, bson.M{"queueStatus": api.OutQueue, "outTime": bson.M{"$lte": time.Now().Add(outTimeout)}}}}
	change := mgo.Change{
		Upsert:    false,
		Remove:    false,
		ReturnNew: false,
		Update:    bson.M{"$set": bson.M{"queueStatus": api.OutQueue, "outTime": time.Now()}},
	}
	result := &api.Event{}

	changeInfo, err := d.eventCollection.Find(query).Sort("inTime").Apply(change, result)
	if err != nil {
		return nil, err
	}

	if changeInfo.Matched > 1 || changeInfo.Updated > 1 {
		return nil, fmt.Errorf("more than 1 same events in queue")
	}

	if changeInfo.Matched == 0 {
		return nil, nil
	}

	return result, nil
}

func (d *DataStore) DeleteEvent(id string) error {
	query := bson.M{"event_id": id}
	return d.eventCollection.Remove(query)
}

func (d *DataStore) UpdateEvent(event *api.Event) error {
	query := bson.M{"event_id": event.EventID}
	return d.eventCollection.Update(query, event)
}

func (d *DataStore) ResetEvent(event *api.Event) error {
	event.Retry = event.Retry + 1
	event.QueueStatus = api.InQueue

	var intervalNum int
	if event.Retry <= retryThreshold {
		intervalNum = event.Retry
	} else {
		intervalNum = retryThreshold
	}

	event.InTime = event.InTime.Add(time.Duration(time.Duration(intervalNum) * retryInterval))

	query := bson.M{"event_id": event.EventID}

	change := mgo.Change{
		Upsert:    false,
		Remove:    false,
		ReturnNew: false,
		Update:    bson.M{"$set": bson.M{"retry": event.Retry, "queueStatus": event.QueueStatus, "inTime": event.InTime}},
	}
	result := &api.Event{}

	changeInfo, err := d.eventCollection.Find(query).Apply(change, result)
	if err != nil {
		return err
	}

	if changeInfo.Matched > 1 || changeInfo.Updated > 1 {
		return fmt.Errorf("more than 1 same events in queue")
	}

	if changeInfo.Matched == 0 {
		return nil
	}

	return nil
}
