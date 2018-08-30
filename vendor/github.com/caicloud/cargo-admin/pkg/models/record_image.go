package models

import (
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/caicloud/nirvana/log"
)

var RecordImage RecordImager

type RecordImager interface {
	EnsureIndexes()
	Save(recimginfo *RecordImageInfo) error
	SaveBatch(recimginfos []*RecordImageInfo) error
	UpdateBatchStatus(repJobId int64, status string) error
	FindOnePage(recordId string, status []string, start, limit int) (int, []*RecordImageInfo, error)
	FindAllByReplication(replication string) ([]*RecordImageInfo, error)
	FindAllByRecord(recordID string) ([]*RecordImageInfo, error)
	DeleteAllByReplication(replication string) error
}

type _RecordImage struct {
	*mgo.Collection
}

var recordImageIndexes = []mgo.Index{
	{Key: []string{"record_id"}},
	{Key: []string{"tenant"}},
	{Key: []string{"replication"}},
	{Key: []string{"rep_job_id"}},
	{Key: []string{"registry"}},
	{Key: []string{"status"}},
}

func (ri *_RecordImage) EnsureIndexes() {
	for _, index := range recordImageIndexes {
		err := ri.EnsureIndex(index)
		if err != nil {
			log.Fatal(err)
		}
	}
}

type RecordImageInfo struct {
	RecordId       bson.ObjectId `bson:"record_id"`
	Tenant         string        `bson:"tenant"`
	Registry       string        `bson:"registry"`
	Replication    string        `bson:"replication"`
	RepJobId       int64         `bson:"rep_job_id"`
	Repository     string        `bson:"repository"`
	Tag            string        `bson:"tag"`
	Operation      string        `bson:"operation"`
	Status         string        `bson:"status"`
	CreationTime   time.Time     `bson:"creation_time"`
	LastUpdateTime time.Time     `bson:"last_update_time"`
}

func (ri *_RecordImage) Save(recimginfo *RecordImageInfo) error {
	return ri.Insert(recimginfo)
}

func (ri *_RecordImage) SaveBatch(recimginfos []*RecordImageInfo) error {
	batch := make([]interface{}, 0, len(recimginfos))
	for _, recimginfo := range recimginfos {
		batch = append(batch, recimginfo)
	}
	return ri.Insert(batch...)
}

func (ri *_RecordImage) UpdateBatchStatus(repJobId int64, status string) error {
	_, err := ri.UpdateAll(bson.M{"rep_job_id": repJobId}, bson.M{"$set": bson.M{"status": status}})
	return err
}

func (ri *_RecordImage) FindOnePage(recordIdStr string, status []string, start, limit int) (int, []*RecordImageInfo, error) {
	var query *mgo.Query
	recordId := bson.ObjectIdHex(recordIdStr)
	recimginfos := make([]*RecordImageInfo, 0)
	if len(status) != 0 {
		orquery := make([]bson.M, 0)
		for _, s := range status {
			orquery = append(orquery, bson.M{"status": s})
		}
		query = ri.Find(bson.M{"record_id": recordId, "$or": orquery})
	} else {
		query = ri.Find(bson.M{"record_id": recordId})
	}

	total, err := query.Count()
	if err != nil {
		log.Errorf("mongo query count error: %v", err)
		return 0, nil, err
	}
	log.Infof("mongo query count total: %d", total)

	err = query.Skip(start).Limit(limit).Sort("creation_time").All(&recimginfos)
	if err != nil {
		log.Errorf("mongo query all error: %v", err)
		return 0, nil, err
	}
	return total, recimginfos, nil
}

func (ri *_RecordImage) FindAllByReplication(replication string) ([]*RecordImageInfo, error) {
	recimginfos := make([]*RecordImageInfo, 0)
	err := ri.Find(bson.M{"replication": replication}).All(&recimginfos)
	if err != nil {
		return nil, err
	}
	return recimginfos, nil
}

func (ri *_RecordImage) FindAllByRecord(recordID string) ([]*RecordImageInfo, error) {
	recordImages := make([]*RecordImageInfo, 0)
	err := ri.Find(bson.M{"record_id": recordID}).All(&recordImages)
	if err != nil {
		return nil, err
	}
	return recordImages, nil
}

func (ri *_RecordImage) DeleteAllByReplication(replication string) error {
	_, err := ri.RemoveAll(bson.M{"replication": replication})
	return err
}
