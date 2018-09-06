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

package handler

import (
	"context"

	"github.com/caicloud/nirvana/log"

	"github.com/caicloud/cyclone/pkg/cloud"
)

// HealthCheck handles the request to check the health of Cyclone server.
func HealthCheck(ctx context.Context) error {

	// Check the health of MongoDB connection.
	if err := ds.Ping(); err != nil {
		log.Errorf("Fail to ping database as %v", err)
		return err
	}

	// Check the health of cloud providers.
	cs, err := ds.FindAllClouds()
	if err != nil {
		log.Errorf("Fail to list all clouds %v", err)
	} else {
		for _, c := range cs {
			cp, err := cloud.NewCloudProvider(&c)
			if err != nil {
				log.Error(err)
				return err
			}

			if err := cp.Ping(); err != nil {
				log.Errorf("Cloud %v is not health as %v", cp, err)
				return err
			}

		}
	}

	return nil
}
