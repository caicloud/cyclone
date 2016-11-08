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
	"github.com/caicloud/cyclone/api"
	"gopkg.in/mgo.v2/bson"
)

// NewTokenDocument creates a new document (record) in mongodb.
func (d *DataStore) NewTokenDocument(token *api.VscToken) error {
	col := d.s.DB(defaultDBName).C(remoteCollectionName)
	_, err := col.Upsert(bson.M{"vsc": token.Vsc, "userid": token.UserID}, token)
	return err
}

// FindtokenByUserID finds token by UserID.
func (d *DataStore) FindtokenByUserID(userID, urlvsc string) (*api.VscToken, error) {
	col := d.s.DB(defaultDBName).C(remoteCollectionName)
	tok := &api.VscToken{}
	err := col.Find(bson.M{"userid": userID, "vsc": urlvsc}).One(tok)
	return tok, err
}

// UpdateToken update token via user ID.
func (d *DataStore) UpdateToken(token *api.VscToken) error {
	col := d.s.DB(defaultDBName).C(remoteCollectionName)
	err := col.Update(bson.M{"userid": token.UserID, "vsc": token.Vsc},
		bson.M{"$set": bson.M{"vsctoken": token.Vsctoken}})
	return err
}

// RemoveTokeninDB removes token.
func (d *DataStore) RemoveTokeninDB(userID string, urlvsc string) error {
	col := d.s.DB(defaultDBName).C(remoteCollectionName)
	err := col.Remove(bson.M{"userid": userID, "vsc": urlvsc})
	return err
}
