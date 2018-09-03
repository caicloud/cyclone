/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package store

import (
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	v1 "github.com/caicloud/devops-admin/pkg/api/v1"
)

var Workspace *_Workspace

type _Workspace struct {
	*mgo.Collection
}

var workspaceIndexs = []mgo.Index{
	{Key: []string{"name"}},
	{Key: []string{"tenant"}},
	{Key: []string{"cycloneProject"}, Unique: true},
}

func (w *_Workspace) EnsureIndexes() error {
	for _, index := range workspaceIndexs {
		err := w.EnsureIndex(index)
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *_Workspace) Save(workspace *v1.Workspace) error {
	workspace.ID = bson.NewObjectId().Hex()
	currentTime := time.Now().Format(time.RFC3339)
	workspace.CreationTime = currentTime
	workspace.LastUpdateTime = currentTime

	return w.Insert(workspace)
}

func (w *_Workspace) FindOnePage(tenant string, start, limit int) ([]v1.Workspace, int, error) {
	workspaces := make([]v1.Workspace, 0)
	query := w.Find(bson.M{"tenant": tenant})

	total, err := query.Count()
	if err != nil {
		return workspaces, 0, err
	}

	// If there is no limit, return all.
	if limit != 0 {
		query.Limit(limit)
	}

	if err = query.Skip(start).Sort("-creationTime").All(&workspaces); err != nil {
		return workspaces, 0, err
	}

	return workspaces, total, err
}

func (w *_Workspace) FindByName(tenant string, name string) (v1.Workspace, error) {
	workspace := v1.Workspace{}
	err := w.Find(bson.M{"tenant": tenant, "name": name}).One(&workspace)
	return workspace, err
}

func (w *_Workspace) FindByNameWithoutTenant(name string) (v1.Workspace, error) {
	workspace := v1.Workspace{}
	err := w.Find(bson.M{"name": name}).One(&workspace)
	return workspace, err
}

func (w *_Workspace) Delete(tenant string, name string) error {
	return w.Remove(bson.M{"tenant": tenant, "name": name})
}

func (w *_Workspace) Update(tenant string, name string, desc string) error {
	return w.Collection.Update(bson.M{"tenant": tenant, "name": name},
		bson.M{
			"$set": bson.M{
				"description":    desc,
				"lastUpdateTime": time.Now().Format(time.RFC3339),
			},
		},
	)
}

func (w *_Workspace) FindAllSortByName(tenant string) ([]v1.Workspace, error) {
	workspaces := make([]v1.Workspace, 0)
	err := w.Find(bson.M{"tenant": tenant}).Sort("name").All(&workspaces)
	if err != nil {
		return nil, err
	}
	return workspaces, nil
}

func (w *_Workspace) IsExist(name string) (bool, error) {
	return IsExist(w.Collection, bson.M{"name": name})
}
