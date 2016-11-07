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

package etcd

import (
	"time"

	"github.com/caicloud/cyclone/pkg/log"
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

// Client is the type for etcd client.
type Client struct {
	client client.Client
}

var (
	c   Client
	ctx context.Context
)

const (
	// Actions about watcher.
	Watch_Action_Create string = "create"
	Watch_Action_Set    string = "set"
	Watch_Action_Delete string = "delete"
)

// Init initializes a client, connected to etcd server.
func Init(endpoints []string) {
	cfg := client.Config{
		Endpoints: endpoints,
		Transport: client.DefaultTransport,
		// Set timeout per request to fail fast when the target endpoint is unavailable.
		HeaderTimeoutPerRequest: time.Second * 5,
	}

	var err error
	c.client, err = client.New(cfg)
	if err != nil {
		log.Fatalf("connect to etcd err: %v", err)
	}
	ctx = context.Background()
}

// GetClient gets a etcd client.
func GetClient() *Client {
	return &c
}

// IsDirExist gets if the path is a dir in etcd server.
func (ec *Client) IsDirExist(dir string) bool {
	kapi := client.NewKeysAPI(ec.client)
	resp, err := kapi.Get(ctx, dir, nil)
	if err != nil {
		return false
	}

	return resp.Node.Dir
}

// CreateDir creates a dir in etcd server.
func (ec *Client) CreateDir(dir string) error {
	kapi := client.NewKeysAPI(ec.client)
	_, err := kapi.Set(ctx, dir, "", &client.SetOptions{Dir: true})
	return err
}

// Set sets value to key in etcd server.
func (ec *Client) Set(key, value string) error {
	kapi := client.NewKeysAPI(ec.client)
	_, err := kapi.Set(ctx, key, value, nil)
	return err
}

// Get gets value from key in etcd server.
func (ec *Client) Get(key string) (value string, err error) {
	kapi := client.NewKeysAPI(ec.client)
	resp, err := kapi.Get(ctx, key, nil)
	if err != nil {
		return "", err
	}
	return resp.Node.Value, nil
}

// Delete deletes a key.
func (ec *Client) Delete(key string) (err error) {
	kapi := client.NewKeysAPI(ec.client)
	_, err = kapi.Delete(ctx, key, nil)
	if err != nil {
		return err
	}
	return nil
}

// List lists keys in a dir.
func (ec *Client) List(dir string) ([]string, error) {
	var values []string
	kapi := client.NewKeysAPI(ec.client)
	resp, err := kapi.Get(ctx, dir, nil)
	if err != nil {
		return values, err
	}

	for _, node := range resp.Node.Nodes {
		respNode, err := kapi.Get(ctx, node.Key, nil)
		if err != nil {
			return values, err
		}
		values = append(values, respNode.Node.Value)
	}
	return values, nil
}

// CreateWatcher creates a watcher to watch a dir.
func (ec *Client) CreateWatcher(dir string) (client.Watcher, error) {
	kapi := client.NewKeysAPI(ec.client)
	respGet, err := kapi.Get(ctx, dir, nil)
	if err != nil {
		return nil, err
	}
	log.Infof("star watch %s after %d", dir, respGet.Index)
	w := kapi.Watcher(dir, &client.WatcherOptions{AfterIndex: respGet.Index,
		Recursive: true})
	return w, err
}
