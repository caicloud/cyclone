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

package resource

import (
	"errors"
	"fmt"
	"sync"

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/pkg/log"
	"github.com/caicloud/cyclone/pkg/osutil"
	"github.com/caicloud/cyclone/store"
)

type Manager struct {
	memoryuser      float64
	memorycontainer float64
	cpuuser         float64
	cpucontainer    float64

	// Protect simultaneous access
	lock sync.Mutex
}

var (
	ErrUnableSupport = errors.New("Unable to support the request resource")
	resourceManager  *Manager
)

// NewManager creates a new resource manager with default resource.
func NewManager() *Manager {
	if resourceManager == nil {
		memoryuser := osutil.GetFloat64Env("MEMORYFORUSER", 4294967296.0) //4G
		cpuuser := osutil.GetFloat64Env("CPUFORUSER", 4096.0)
		memorycontainer := osutil.GetFloat64Env("MEMORYFORCONTAINER", 536870912.0) //512M
		cpucontainer := osutil.GetFloat64Env("CPUFORCONTAINER", 512.0)

		resourceManager = &Manager{
			memoryuser:      memoryuser,
			memorycontainer: memorycontainer,
			cpuuser:         cpuuser,
			cpucontainer:    cpucontainer,
		}
	}
	return resourceManager
}

// ApplyResource uses to apply resource for container.
func (resm *Manager) ApplyResource(event *api.Event) error {
	resm.lock.Lock()
	defer resm.lock.Unlock()

	ds := store.NewStore()
	defer ds.Close()
	resource, err := ds.FindResourceByID(event.Service.UserID)
	if err != nil {
		// Come in, we think that it is the first time to create version according userid, so need add new document
		resource.UserID = event.Service.UserID
		if event.Version.BuildResource.CPU == 0 || event.Version.BuildResource.Memory == 0 {
			event.Version.BuildResource.Memory = resm.memorycontainer
			event.Version.BuildResource.CPU = resm.cpucontainer
		}
		resource.TotalResource.CPU = resm.cpuuser
		resource.TotalResource.Memory = resm.memoryuser
		resource.PerResource.CPU = resm.cpucontainer
		resource.PerResource.Memory = resm.memorycontainer
		resource.LeftResource.Memory = resource.TotalResource.Memory
		resource.LeftResource.CPU = resource.TotalResource.CPU

		if resource.TotalResource.Memory < event.Version.BuildResource.Memory ||
			resource.TotalResource.CPU < event.Version.BuildResource.CPU {
			errResource := fmt.Errorf("the total resource < the request resource")
			log.Errorf("Unable to support the request resource %+v, because > the total resource", event.Service.UserID)
			return errResource
		}

		resource.LeftResource.Memory = resource.LeftResource.Memory - event.Version.BuildResource.Memory
		resource.LeftResource.CPU = resource.LeftResource.CPU - event.Version.BuildResource.CPU
		if err := ds.NewResourceDocument(resource); err != nil {
			log.Errorf("Unable to create new resource document %+v: %v", event.Service.UserID, err)
			return err
		}
	} else {
		if event.Version.BuildResource.CPU == 0 || event.Version.BuildResource.Memory == 0 {
			event.Version.BuildResource.Memory = resource.PerResource.Memory
			event.Version.BuildResource.CPU = resource.PerResource.CPU
		}

		if resource.TotalResource.Memory < event.Version.BuildResource.Memory ||
			resource.TotalResource.CPU < event.Version.BuildResource.CPU {
			errResource := fmt.Errorf("the total resource < the request resource")
			log.Errorf("Unable to support the request resource %+v, because > the total resource", event.Service.UserID)
			return errResource
		}

		if resource.LeftResource.Memory < event.Version.BuildResource.Memory ||
			resource.LeftResource.CPU < event.Version.BuildResource.CPU {
			log.Infof("Unable to support the request resource %+v", event.Service.UserID)
			return ErrUnableSupport
		}

		resource.LeftResource.Memory = resource.LeftResource.Memory - event.Version.BuildResource.Memory
		resource.LeftResource.CPU = resource.LeftResource.CPU - event.Version.BuildResource.CPU
		if err = ds.UpdateResourceStatus(resource.UserID, resource.LeftResource.Memory, resource.LeftResource.CPU); err != nil {
			log.Errorf("Unable to update resource status %+v: %v", event.Service.UserID, err)
			return err
		}
	}

	log.Infof("After apply, the userid %s's left resource  memory %f cpu  %f numberContainers %d", resource.UserID,
		resource.LeftResource.Memory, resource.LeftResource.CPU)
	return nil
}

// ReleaseResource uses to add resource into db.
func (resm *Manager) ReleaseResource(event *api.Event) error {
	ds := store.NewStore()
	defer ds.Close()
	resource, err := ds.FindResourceByID(event.Service.UserID)
	if err != nil {
		return err
	}
	resource.LeftResource.Memory = resource.LeftResource.Memory + event.Version.BuildResource.Memory
	resource.LeftResource.CPU = resource.LeftResource.CPU + event.Version.BuildResource.CPU
	if err = ds.UpdateResourceStatus(resource.UserID, resource.LeftResource.Memory, resource.LeftResource.CPU); err != nil {
		log.ErrorWithFields("Unable to update resource status", log.Fields{"err": err, "usrid": event.Service.UserID})
		return err
	}
	log.Infof("Afer release, the userid %s's left resource  memory %f cpu  %f", resource.UserID,
		resource.LeftResource.Memory, resource.LeftResource.CPU)
	return nil
}

// GetMemorycontainer func that get the default memory for container.
func (resm *Manager) GetMemorycontainer() float64 {
	return resm.memorycontainer
}

// GetCpucontainer func that get the default cpu for container.
func (resm *Manager) GetCpucontainer() float64 {
	return resm.cpucontainer
}

// GetMemoryuser func that get the default memory for user.
func (resm *Manager) GetMemoryuser() float64 {
	return resm.memoryuser
}

// GetCpuuser func that get the default cpu for user.
func (resm *Manager) GetCpuuser() float64 {
	return resm.cpuuser
}
