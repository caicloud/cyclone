package models

import (
	"fmt"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/caicloud/nirvana/log"
)

var Replication Replicationer

type Replicationer interface {
	EnsureIndexes()
	Save(repinfo *ReplicationInfo) error
	CountAsSoure(tenant, registry, project string) (int, error)
	CountAsTarget(tenant, registry, project string) (int, error)
	FindByName(tenant string, name string) (*ReplicationInfo, error)
	FindOnePage(tenant string, start, limit int) (int, []*ReplicationInfo, error)
	FindOnePageBySourceRegistry(tenant string, sregistry string, param *FindOnePageParams) (int, []*ReplicationInfo, error)
	FindOnePageByTargetRegistry(tenant string, tregistry string, param *FindOnePageParams) (int, []*ReplicationInfo, error)
	FindAllBySourceRegistry(sregistry string) ([]*ReplicationInfo, error)
	FindAllByTargetRegistry(tregistry string) ([]*ReplicationInfo, error)
	FindAllByProject(project string) ([]*ReplicationInfo, error)
	UpdateReplication(tenant string, name string, alias string, targetRegistry string, triggerKind string, uptime time.Time) error
	UpdateLastListRecordsTime(tenant string, repliation string, lastListRecordsTime time.Time) error
	UpdateLastUpdateTime(tenant string, repliation string, lastUpdateTime time.Time) error
	UpdateLastTriggerTime(tenant string, repliation string, lastTriggerTime time.Time) error
	Delete(tenant string, name string) error
	IsExist(tenant string, name string) (bool, error)
}

type _Replication struct {
	*mgo.Collection
}

var replicationIndexes = []mgo.Index{
	{Key: []string{"name"}, Unique: true},
	{Key: []string{"project"}},
	{Key: []string{"source_registry"}},
	{Key: []string{"target_registry"}},
}

func (r *_Replication) EnsureIndexes() {
	EnsureIndexes(r.Collection, replicationIndexes)
}

// Replication Collection 中要存哪些内容、不存哪些内容
// 实际上，Replication Collectoin 只存 Name、Tenant、SourceRegistry、ReplicationPolicyId 和 Alias 和这五个字段，
// 基本可以满足 replication CRUD 的需求。

// 但是，
// 1. 为了满足 正向查询和反向查询的需求，TargetRegistry, SourceProject, TargetProject 也是必须要加到，否则很难实现分页；
// 2. 为了满足 LSIT replications 时使用 riggerKind 对结果进行过滤后分页，TriggerKindn 也是必须要加的。

// 此外，
// 为了方便实现 LIST /replicatoins/{replication}/records 接口，四个时间字段起着极为关键的作用，必不可少。
// 因此 Replication collection 中记录的字段会多一些。
// 但是，不建议记录 harbor replication 的全部信息，否则很难处理数据不一致的情况。
type ReplicationInfo struct {
	Name                string    `bson:"name"`
	Tenant              string    `bson:"tenant"`
	ReplicationPolicyId int64     `bson:"replication_policy_id"`
	Alias               string    `bson:"alias"`
	Project             string    `bson:"project"`
	TriggerKind         string    `bson:"trigger_kind"`
	SourceRegistry      string    `bson:"source_registry"`
	TargetRegistry      string    `bson:"target_registry"`
	CreationTime        time.Time `bson:"creation_time"`
	LastUpdateTime      time.Time `bson:"last_update_time"`
	LastTriggerTime     time.Time `bson:"last_trigger_time"`
	LastListRecordsTime time.Time `bson:"last_list_records_time"`
}

func (r *_Replication) Save(repinfo *ReplicationInfo) error {
	return r.Insert(repinfo)
}

func (r *_Replication) CountAsSoure(tentant, registry, project string) (int, error) {
	total, err := r.Find(bson.M{"tenant": tentant, "source_registry": registry, "project": project}).Count()
	if err != nil {
		log.Errorf("Get rep count as source for registry %s and project %s error: %v", registry, project, err)
		return 0, fmt.Errorf("get count error: %s, %s", registry, project)
	}
	return total, nil
}

func (r *_Replication) CountAsTarget(tentant, registry, project string) (int, error) {
	total, err := r.Find(bson.M{"tenant": tentant, "target_registry": registry, "project": project}).Count()
	if err != nil {
		log.Errorf("Get rep count as target for registry %s and project %s error: %v", registry, project, err)
		return 0, fmt.Errorf("get count error: %s, %s", registry, project)
	}
	return total, nil
}

func (r *_Replication) FindAllByProject(project string) ([]*ReplicationInfo, error) {
	replications := make([]*ReplicationInfo, 0)
	err := r.Find(bson.M{"project": project}).All(&replications)
	if err != nil {
		log.Errorf("mongo query all error: %v", err)
		return nil, err
	}
	return replications, nil
}

func (r *_Replication) FindByName(tenant string, name string) (*ReplicationInfo, error) {
	repinfo := &ReplicationInfo{}
	err := r.Find(bson.M{"tenant": tenant, "name": name}).One(repinfo)
	return repinfo, err
}

func (r *_Replication) FindOnePage(tenant string, start, limit int) (int, []*ReplicationInfo, error) {
	repinfos := make([]*ReplicationInfo, 0)
	query := r.Find(bson.M{"tenant": tenant})
	total, err := query.Count()
	if err != nil {
		log.Errorf("mongo query count error: %v", err)
		return 0, nil, err
	}
	log.Infof("mongo query count total: %d", total)
	err = query.Skip(start).Limit(limit).Sort("-creation_time").All(&repinfos)
	if err != nil {
		log.Errorf("mongo query all error: %v", err)
		return 0, nil, err
	}

	return total, repinfos, nil
}

type FindOnePageParams struct {
	Project     string
	TriggerType string
	Prefix      string
	Start       int
	Limit       int
}

func (r *_Replication) FindOnePageBySourceRegistry(tenant string, sregistry string, param *FindOnePageParams) (int, []*ReplicationInfo, error) {
	repinfos := make([]*ReplicationInfo, 0)
	queryBson := bson.M{"tenant": tenant, "source_registry": sregistry, "name": bson.M{"$regex": "^" + param.Prefix}}
	if param.Project != "" {
		queryBson["project"] = param.Project
	}
	if param.TriggerType != "" {
		queryBson["trigger_kind"] = param.TriggerType
	}
	query := r.Find(queryBson)

	// 此处 query.Count() 一定要在 query.Skip() 和 query.Limit() 调用，否则会出错
	total, err := query.Count()
	if err != nil {
		log.Errorf("mongo query count error: %v", err)
		return 0, nil, err
	}
	log.Infof("mongo query count total: %d", total)

	err = query.Skip(param.Start).Limit(param.Limit).Sort("-creation_time").All(&repinfos)
	if err != nil {
		log.Errorf("mongo query all error: %v", err)
		return 0, nil, err
	}
	return total, repinfos, nil
}

func (r *_Replication) FindOnePageByTargetRegistry(tenant string, tregistry string, param *FindOnePageParams) (int, []*ReplicationInfo, error) {
	repinfos := make([]*ReplicationInfo, 0)
	queryBson := bson.M{"tenant": tenant, "target_registry": tregistry, "name": bson.M{"$regex": "^" + param.Prefix}}
	if param.Project != "" {
		queryBson["project"] = param.Project
	}
	if param.TriggerType != "" {
		queryBson["trigger_kind"] = param.TriggerType
	}
	query := r.Find(queryBson)

	// 此处 query.Count() 一定要在 query.Skip() 和 query.Limit() 调用，否则会出错
	total, err := query.Count()
	if err != nil {
		log.Errorf("mongo query count error: %v", err)
		return 0, nil, err
	}
	log.Infof("mongo query count total: %d", total)

	err = query.Skip(param.Start).Limit(param.Limit).Sort("-creation_time").All(&repinfos)
	if err != nil {
		log.Errorf("mongo query all error: %v", err)
		return 0, nil, err
	}
	return total, repinfos, nil
}

func (r *_Replication) FindAllBySourceRegistry(sregistry string) ([]*ReplicationInfo, error) {
	repinfos := make([]*ReplicationInfo, 0)
	err := r.Find(bson.M{"source_registry": sregistry}).All(&repinfos)
	if err != nil {
		return nil, err
	}
	return repinfos, nil
}

func (r *_Replication) FindAllByTargetRegistry(tregistry string) ([]*ReplicationInfo, error) {
	repinfos := make([]*ReplicationInfo, 0)
	err := r.Find(bson.M{"target_registry": tregistry}).All(&repinfos)
	if err != nil {
		return nil, err
	}
	return repinfos, nil
}

func (r *_Replication) UpdateReplication(tenant string, name string, alias string,
	targetRegistry string, triggerKind string, uptime time.Time) error {
	return r.Update(bson.M{"tenant": tenant, "name": name}, bson.M{"$set": bson.M{"alias": alias, "target_registry": targetRegistry, "trigger_kind": triggerKind, "last_update_time": uptime}})
}

func (r *_Replication) UpdateLastListRecordsTime(tenant string, repliation string, lastListRecordsTime time.Time) error {
	return r.Update(bson.M{"name": repliation}, bson.M{"$set": bson.M{"last_list_records_time": lastListRecordsTime}})
}

func (r *_Replication) UpdateLastUpdateTime(tenant string, repliation string, lastUpdateTime time.Time) error {
	return r.Update(bson.M{"name": repliation}, bson.M{"$set": bson.M{"last_update_time": lastUpdateTime}})
}

func (r *_Replication) UpdateLastTriggerTime(tenant string, repliation string, lastTriggerTime time.Time) error {
	return r.Update(bson.M{"tenant": tenant, "name": repliation}, bson.M{"$set": bson.M{"last_trigger_time": lastTriggerTime}})
}

func (r *_Replication) Delete(tenant string, name string) error {
	return r.Remove(bson.M{"tenant": tenant, "name": name})
}

func (r *_Replication) IsExist(tenant string, name string) (bool, error) {
	return IsExist(r.Collection, bson.M{"tenant": tenant, "name": name})
}
