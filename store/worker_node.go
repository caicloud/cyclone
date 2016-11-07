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
	"github.com/satori/go.uuid"
	"gopkg.in/mgo.v2/bson"
)

// NewSystemWorkerNodeDocument creates a new document (record) in mongodb. It returns worker node
// id of the newly created worker node.
func (d *DataStore) NewSystemWorkerNodeDocument(workerNode *api.WorkerNode) (string, error) {
	workerNode.NodeID = uuid.NewV4().String()
	col := d.s.DB(defaultDBName).C(workerNodeCollection)
	_, err := col.Upsert(bson.M{"_id": workerNode.NodeID}, workerNode)
	return workerNode.NodeID, err
}

// FindWorkerNodesByDockerHost finds a list of WorkerNodes via docker host.
func (d *DataStore) FindWorkerNodesByDockerHost(dockerHost string) ([]api.WorkerNode, error) {
	nodes := []api.WorkerNode{}
	filter := bson.M{"docker_host": dockerHost}
	col := d.s.DB(defaultDBName).C(workerNodeCollection)
	err := col.Find(filter).Iter().All(&nodes)
	return nodes, err
}

// FindWorkerNodeByID finds a worker node entity by ID.
func (d *DataStore) FindWorkerNodeByID(nodeID string) (*api.WorkerNode, error) {
	node := &api.WorkerNode{}
	col := d.s.DB(defaultDBName).C(workerNodeCollection)
	err := col.Find(bson.M{"_id": nodeID}).One(node)
	return node, err
}

// FindSystemWorkerNode finds a list of system worker node.
func (d *DataStore) FindSystemWorkerNode() ([]api.WorkerNode, error) {
	workerNodes := []api.WorkerNode{}
	filter := bson.M{"type": api.SystemWorkerNode}
	col := d.s.DB(defaultDBName).C(workerNodeCollection)
	err := col.Find(filter).Iter().All(&workerNodes)
	return workerNodes, err
}

// DeleteWorkerNodeByID removes worker node by node_id.
func (d *DataStore) DeleteWorkerNodeByID(nodeID string) error {
	col := d.s.DB(defaultDBName).C(workerNodeCollection)
	err := col.Remove(bson.M{"_id": nodeID})
	return err
}

// FindSystemWorkerNodeByResource finds a list of system worker node by resouce.
func (d *DataStore) FindSystemWorkerNodeByResource(resource *api.BuildResource) ([]api.WorkerNode, error) {
	workerNodes := []api.WorkerNode{}
	filter := bson.M{
		"type":                 api.SystemWorkerNode,
		"left_resource.memory": bson.M{"$gte": resource.Memory},
		"left_resource.cpu":    bson.M{"$gte": resource.CPU},
	}
	col := d.s.DB(defaultDBName).C(workerNodeCollection)
	err := col.Find(filter).Sort("-left_resource.memory").Iter().All(&workerNodes)
	return workerNodes, err
}

// UpsertWorkerNodeDocument upsert a special woker node document
func (d *DataStore) UpsertWorkerNodeDocument(node *api.WorkerNode) (string, error) {
	col := d.s.DB(defaultDBName).C(workerNodeCollection)
	_, err := col.Upsert(bson.M{"_id": node.NodeID}, node)
	return node.NodeID, err
}
