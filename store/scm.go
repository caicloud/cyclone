/*
Copyright 2017 caicloud authors. All rights reserved.

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
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// CreateToken creates creates the token, returns the token created.
func (d *DataStore) CreateToken(token *api.ScmToken) (*api.ScmToken, error) {
	if err := d.scmCollection.Insert(token); err != nil {
		return nil, err
	}
	return token, nil
}

// Findtoken finds the token by project id and scm type.
func (d *DataStore) Findtoken(projectID, scmType string) (*api.ScmToken, error) {

	query := bson.M{"projectID": projectID, "scmType": scmType}

	count, err := d.scmCollection.Find(query).Count()
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, mgo.ErrNotFound
	}

	token := &api.ScmToken{}
	if err := d.scmCollection.Find(query).One(token); err != nil {
		return nil, err
	}
	return token, nil
}

// UpdateToken2 updates the token, please make sure the project id and scm type is provided before call this method.
func (d *DataStore) UpdateToken2(token *api.ScmToken) error {
	return d.scmCollection.Update(bson.M{"projectID": token.ProjectID, "scmType": token.ScmType}, token)
}

// DeleteToken deletes the token with the project id and scm type.
func (d *DataStore) DeleteToken(projectID, scmType string) error {
	return d.scmCollection.Remove(bson.M{"projectID": projectID, "scmType": scmType})
}
