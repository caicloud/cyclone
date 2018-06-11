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

package router

import (
	"net/http"

	restful "github.com/emicklei/go-restful"
	log "github.com/golang/glog"

	"github.com/caicloud/cyclone/pkg/cloud"
)

// healthCheck handles the request to check the health of Cyclone server.
func (router *router) healthCheck(request *restful.Request, response *restful.Response) {
	status := http.StatusInternalServerError

	// Check the health of MongoDB connection.
	dbHealth := false
	if err := router.dataStore.Ping(); err != nil {
		log.Errorf("Fail to ping database as %v", err)
	} else {
		dbHealth = true
	}

	// Check the health of cloud providers.
	cloudHealth := false
	cs, err := router.dataStore.FindAllClouds()
	if err != nil {
		log.Errorf("Fail to list all clouds %v", err)
	} else {
		for _, c := range cs {
			cp, err := cloud.NewCloudProvider(&c)
			if err != nil {
				log.Error(err)
				break
			}

			if err := cp.Ping(); err != nil {
				log.Errorf("Cloud %v is not health as %v", cp, err)
				break
			}

			cloudHealth = true
		}
	}

	if dbHealth && cloudHealth {
		status = http.StatusOK
	}

	response.WriteHeader(status)
}
