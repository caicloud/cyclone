package models

import (
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var DockerAccount *_DockerAccount

type _DockerAccount struct {
	*mgo.Collection
}

var dockerAccountIndexes = []mgo.Index{
	{Key: []string{"username"}},
	{Key: []string{"registry"}},
}

func (d *_DockerAccount) EnsureIndexes() {
	EnsureIndexes(d.Collection, dockerAccountIndexes)
}

type DockerAccountInfo struct {
	Username   string    `bson:"username"`
	Password   string    `bson:"password"`
	Registry   string    `bson:"registry"`
	CreateTime time.Time `bson:"create_time"`
}

func (d *_DockerAccount) Save(dainfo *DockerAccountInfo) error {
	return d.Insert(dainfo)
}

func (d *_DockerAccount) FindByName(registry string, username string) (*DockerAccountInfo, error) {
	dainfo := &DockerAccountInfo{}
	err := d.Find(bson.M{"registry": registry, "username": username}).One(dainfo)
	return dainfo, err
}

func (d *_DockerAccount) IsExist(registry string, username string) (bool, error) {
	return IsExist(d.Collection, bson.M{"registry": registry, "username": username})
}

func (d *_DockerAccount) DeleteAll(registry string) error {
	_, err := d.RemoveAll(bson.M{"registry": registry})
	return err
}
