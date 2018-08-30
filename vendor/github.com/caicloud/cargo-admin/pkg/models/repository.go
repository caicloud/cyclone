package models

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var Repository *_Repository

type Repositoryer interface {
	EnsureIndexes()
	Save(repoInfo *RepositoryInfo) error
	FindByName(registry string, projectId int64, name string) (*RepositoryInfo, error)
	Delete(registry string, projectId int64, name string) error
	Upsert(registry string, projectId int64, name string, des string) error
	IsExist(registry string, projectId int64, name string) (bool, error)
}

type _Repository struct {
	*mgo.Collection
}

var repositoryIndexes = []mgo.Index{
	{Key: []string{"name"}},
	{Key: []string{"project_id"}},
	{Key: []string{"registry"}},
}

func (r *_Repository) EnsureIndexes() {
	EnsureIndexes(r.Collection, repositoryIndexes)
}

type RepositoryInfo struct {
	Registry    string `bson:"registry"`
	ProjectId   int64  `bson:"project_id"`
	Name        string `bson:"name"`
	Description string `bson:"description"`
}

func (r *_Repository) Save(repoInfo *RepositoryInfo) error {
	return r.Insert(repoInfo)
}

func (r *_Repository) FindByName(registry string, projectId int64, name string) (*RepositoryInfo, error) {
	repoInfo := &RepositoryInfo{}
	err := r.Find(bson.M{"registry": registry, "project_id": projectId, "name": name}).One(repoInfo)
	return repoInfo, err
}

func (r *_Repository) Delete(registry string, projectId int64, name string) error {
	return r.Remove(bson.M{"registry": registry, "project_id": projectId, "name": name})
}

func (r *_Repository) Upsert(registry string, projectId int64, name string, des string) error {
	_, err := r.Collection.Upsert(bson.M{"registry": registry, "project_id": projectId, "name": name},
		bson.M{
			"registry":    registry,
			"project_id":  projectId,
			"name":        name,
			"description": des,
		},
	)
	return err
}

func (r *_Repository) IsExist(registry string, projectId int64, name string) (bool, error) {
	return IsExist(r.Collection, bson.M{"registry": registry, "project_id": projectId, "name": name})
}
