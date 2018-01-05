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
	"github.com/caicloud/cyclone/pkg/api"

	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

const (
	inQ  int = 0
	outQ int = 1

	removeTimeout = 240

	JobStatusFailed = "failed"
)

type Massage struct {
	ID      string `bson:"_id" json:"_id,omitempty"`
	Job     *Job   `bson:"job" json:"job,omitempty"`
	State   int    `bson:"state" json:"state,omitempty"`
	Time    int64  `bson:"timestamp" json:"timestamp,omitempty"`
	OutTime int64  `bson:"outTime" json:"outTime,omitempty"`
	Retry   int    `bson:"retry" json:"retry,omitempty"`
}

type Job struct {
	Retry          int
	Pipeline       *api.Pipeline       `bson:"pipeline" json:"pipeline,omitempty"`
	PipelineRecord *api.PipelineRecord `bson:"pipelineRecord" json:"pipeline,omitempty"`
}

// CreateMassage enqueues the job.
func (d *dataStore) CreateMassage(job *Job) {
	d.queueCollection.Insert(&Massage{
		ID:    string(job.PipelineRecord.ID),
		Job:   job,
		State: inQ,
		Time:  time.Now().Unix(),
	})
}

// GetMassage dequeues the job.
func (d *dataStore) GetMassage() (*Massage, error) {
	query := bson.M{"$or": []bson.M{bson.M{"state": inQ}, bson.M{"state": outQ, "outTime": bson.M{"$lte": time.Now().Unix() - removeTimeout}}}}
	change := mgo.Change{
		Upsert:    false,
		Remove:    false,
		ReturnNew: false,
		Update:    bson.M{"$set": bson.M{"state": outQ, "outTime": time.Now().Unix()}},
	}
	result := &Massage{}

	changeInfo, err := d.queueCollection.Find(query).Sort("timestamp").Apply(change, result)
	if err != nil {
		return nil, err
	}

	if changeInfo.Matched > 1 || changeInfo.Updated > 1 {
		return nil, fmt.Errorf("more than 1 same jobs in queue")
	}

	if changeInfo.Matched == 0 {
		return nil, nil
	}

	return result, nil
}

// RemoveMassage remove the used job.
func (d *dataStore) RemoveMassage(id string) {
	d.queueCollection.Remove(id)
}

// ResetMassage re-enqueues the job when busy.
func (d *dataStore) ResetMassage(m *Massage) error {
	m.Retry = m.Retry + 1
	m.State = inQ

	var addTime time.Duration
	if m.Retry > 6 {
		addTime = 3600
	} else {
		addTime = time.Duration(m.Retry) * 600
	}
	m.Time = time.Unix(m.Time, 0).Add(time.Duration(addTime * time.Second)).Unix()

	query := bson.M{"_id": m.ID}

	count, err := d.queueCollection.Find(query).Count()
	if err != nil {
		return err
	}

	if count == 0 {
		return mgo.ErrNotFound
	} else if count > 1 {
		return fmt.Errorf("there are %d items with the same id %s", count, m.ID)
	}

	return d.queueCollection.Update(query, m)
}
