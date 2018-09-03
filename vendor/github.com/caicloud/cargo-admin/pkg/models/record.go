package models

import (
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/caicloud/nirvana/log"
)

var Record Recorder

type Recorder interface {
	EnsureIndexes()
	Save(recinfo *RecordInfo) error
	SaveBatch(recinfos []*RecordInfo) error
	FindOne(recordID string) (*RecordInfo, error)
	FindOnePage(replication string, triggerType string, status string, start int, limit int) (int, []*RecordInfo, error)
	FindByStatus(replication string, status string) ([]*RecordInfo, error)
	UpdateStatus(replication string, ID bson.ObjectId, status string) error
	DeleteAllByReplication(replication string) error
}

type _Record struct {
	*mgo.Collection
}

var recordIndexes = []mgo.Index{
	{Key: []string{"registry"}},
	{Key: []string{"replication"}},
	{Key: []string{"status"}},
}

func (r *_Record) EnsureIndexes() {
	EnsureIndexes(r.Collection, recordIndexes)
}

type RecordInfo struct {
	Id             bson.ObjectId `bson:"_id"`
	Registry       string        `bson:"registry"`
	Tenant         string        `bson:"tenant"`
	Replication    string        `bson:"replication"`
	RepJobIds      []int64       `bson:"rep_job_ids"`
	Trigger        *Trigger      `bson:"trigger"`
	Status         string        `bson:"status"`
	Reason         string        `bson:"reason"`
	CreationTime   time.Time     `bson:"creation_time"`
	LastUpdateTime time.Time     `bson:"last_update_time"`
}

type Trigger struct {
	Kind          string         `bson:"kind"`
	ScheduleParam *ScheduleParam `bson:"schedule_param"`
}

type ScheduleParam struct {
	Type    string `bson:"kind"`
	Weekday int8   `bson:"weekday"`
	Offtime int64  `bson:"offtime"`
}

func (r *_Record) Save(recinfo *RecordInfo) error {
	return r.Insert(recinfo)
}

func (r *_Record) SaveBatch(recinfos []*RecordInfo) error {
	batch := make([]interface{}, 0, len(recinfos))
	for _, recinfo := range recinfos {
		batch = append(batch, recinfo)
	}
	return r.Insert(batch...)
}

func (r *_Record) FindOne(recordID string) (*RecordInfo, error) {
	record := &RecordInfo{}
	return record, r.Find(bson.M{"_id": recordID}).One(record)
}

func (r *_Record) FindOnePage(replication string, triggerType string, status string, start int, limit int) (int, []*RecordInfo, error) {
	recinfos := make([]*RecordInfo, 0)
	q := bson.M{"replication": replication}
	if triggerType != "" {
		q["trigger.kind"] = triggerType
	}
	if status != "" {
		q["status"] = status
	}
	query := r.Find(q)

	total, err := query.Count()
	if err != nil {
		log.Errorf("mongo query count error: %v", err)
		return 0, nil, err
	}
	log.Infof("mongo query count total: %d", total)

	err = query.Skip(start).Limit(limit).Sort("-creation_time").All(&recinfos)
	if err != nil {
		log.Errorf("mongo query all error: %v", err)
		return 0, nil, err
	}

	return total, recinfos, nil
}

func (r *_Record) UpdateStatus(replication string, ID bson.ObjectId, status string) error {
	err := r.Update(bson.M{"_id": ID, "replication": replication}, bson.M{"$set": bson.M{"status": status, "last_update_time": time.Now()}})
	if err != nil {
		return err
	}

	return nil
}

func (r *_Record) FindByStatus(replication string, status string) ([]*RecordInfo, error) {
	recinfos := make([]*RecordInfo, 0)
	err := r.Find(bson.M{"replication": replication, "status": status}).All(&recinfos)
	if err != nil {
		return nil, err
	}

	return recinfos, nil
}

func (r *_Record) DeleteAllByReplication(replication string) error {
	_, err := r.RemoveAll(bson.M{"replication": replication})
	return err
}
