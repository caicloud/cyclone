/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package api

import (
	"github.com/caicloud/devops-admin/pkg/api/v1"
	"github.com/emicklei/go-restful"
)

// Initialize initializes apis of all versions
func Initialize() {
	v1.InstallRouters(restful.DefaultContainer)
	restful.EnableTracing(true)
}
