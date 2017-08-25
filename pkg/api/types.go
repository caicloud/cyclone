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

package api

import "time"

// Project represents a group to manage a set of related applications. It maybe a real project, which contains several or many applications.
type Project struct {
	ID          string    `bson:"_id,omitempty" json:"id,omitempty" description:"id of the project"`
	Name        string    `bson:"name,omitempty" json:"name,omitempty" description:"name of the project, should be unique"`
	Description string    `bson:"description,omitempty" json:"description,omitempty" description:"description of the project"`
	Owner       string    `bson:"owner,omitempty" json:"owner,omitempty" description:"owner of the project"`
	CreatedTime time.Time `bson:"createdTime,omitempty" json:"createdTime,omitempty" description:"created time of the project"`
	UpdatedTime time.Time `bson:"updatedTime,omitempty" json:"updatedTime,omitempty" description:"updated time of the project"`
}
