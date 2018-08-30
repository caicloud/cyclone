package models

import (
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/caicloud/nirvana/log"
)

var Registry Registryer

type Registryer interface {
	EnsureIndexes()
	Save(reginfo *RegistryInfo) error
	FindOnePage(start, limit int) (int, []*RegistryInfo, error)
	FindAll() ([]*RegistryInfo, error)
	FindByName(name string) (*RegistryInfo, error)
	FindByHost(host string) (*RegistryInfo, error)
	FindByDomain(domain string) (*RegistryInfo, error)
	Delete(name string) error
	Update(name, alias, username, password string) error
	IsExistHost(host string) (bool, error)
	IsExist(name string) (bool, error)
}

type _Registry struct {
	*mgo.Collection
}

var registryIndexes = []mgo.Index{
	{Key: []string{"name"}, Unique: true},
	{Key: []string{"host"}, Unique: true},
	{Key: []string{"domain"}, Unique: true},
}

func (r *_Registry) EnsureIndexes() {
	EnsureIndexes(r.Collection, registryIndexes)
}

type RegistryInfo struct {
	Name           string    `bson:"name"`
	Alias          string    `bson:"alias"`
	Host           string    `bson:"host"`
	Domain         string    `bson:"domain"`
	Username       string    `bson:"username"`
	Password       string    `bson:"password"`
	CreationTime   time.Time `bson:"creation_time"`
	LastUpdateTime time.Time `bson:"last_update_time"`
}

type RegTargetInfo struct {
	TargetId int64  `json:"targe_id"`
	Registry string `json:"registry"`
}

// 注：List Registries(FindOnePage 和 FindAll 两个函数) 时，返回结果的排序和其他处不同，此处是按照时间正序排序，其余绝大多数都按时间倒序排序

func (r *_Registry) Save(reginfo *RegistryInfo) error {
	return r.Insert(reginfo)
}

func (r *_Registry) FindOnePage(start, limit int) (int, []*RegistryInfo, error) {
	reginfos := make([]*RegistryInfo, 0)
	query := r.Find(bson.M{})
	total, err := query.Count()
	if err != nil {
		log.Errorf("mongo query count error: %v", err)
		return 0, nil, err
	}
	log.Infof("mongo query count total: %d", total)

	err = query.Skip(start).Limit(limit).Sort("creation_time").All(&reginfos)
	if err != nil {
		log.Errorf("mongo query all error: %v", err)
		return 0, nil, err
	}

	return total, reginfos, nil
}

func (r *_Registry) FindAll() ([]*RegistryInfo, error) {
	reginfos := make([]*RegistryInfo, 0)
	err := r.Find(bson.M{}).Sort("creation_time").All(&reginfos)
	if err != nil {
		log.Errorf("mongo query all error: %v", err)
		return nil, err
	}

	return reginfos, nil
}

func (r *_Registry) FindByName(name string) (*RegistryInfo, error) {
	reginfo := &RegistryInfo{}
	err := r.Find(bson.M{"name": name}).One(reginfo)
	return reginfo, err
}

func (r *_Registry) FindByHost(host string) (*RegistryInfo, error) {
	reginfo := &RegistryInfo{}
	err := r.Find(bson.M{"host": host}).One(reginfo)
	return reginfo, err
}

func (r *_Registry) FindByDomain(domain string) (*RegistryInfo, error) {
	reginfo := &RegistryInfo{}
	err := r.Find(bson.M{"domain": domain}).One(reginfo)
	return reginfo, err
}

func (r *_Registry) Delete(name string) error {
	return r.Remove(bson.M{"name": name})
}

func (r *_Registry) Update(name, alias, username, password string) error {
	return r.Collection.Update(bson.M{"name": name},
		bson.M{
			"$set": bson.M{
				"alias":            alias,
				"username":         username,
				"password":         password,
				"last_update_time": time.Now().Format(time.RFC3339),
			},
		},
	)
}

func (r *_Registry) IsExistHost(host string) (bool, error) {
	return IsExist(r.Collection, bson.M{"host": host})
}

func (r *_Registry) IsExist(name string) (bool, error) {
	return IsExist(r.Collection, bson.M{"name": name})
}
