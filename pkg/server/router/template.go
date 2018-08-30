/*
Copyright 2018 caicloud authors. All rights reserved.

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
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/emicklei/go-restful"
	log "github.com/golang/glog"
	"gopkg.in/yaml.v2"

	"github.com/caicloud/cyclone/pkg/api"
	httputil "github.com/caicloud/cyclone/pkg/util/http"
)

const TemplatePath = "/config/templates/templates.yaml"

var templates []*api.Template
var once sync.Once

// listTemplates handles the request to list all cyclone built-in pipeline templates.
func (router *router) listTemplates(request *restful.Request, response *restful.Response) {
	once.Do(
		func() {
			log.Infof("read pipeline templates from %s", TemplatePath)
			result, err := getTemplates(TemplatePath)
			if err != nil {
				log.Errorf("unmarshal yaml template file %s error: %v", TemplatePath, err)
				return
			}
			templates = result
			log.Info("read dockerfile templates succeed")
		},
	)

	response.WriteHeaderAndEntity(http.StatusOK, httputil.ResponseWithList(templates, len(templates)))

}

func getTemplates(templatePath string) ([]*api.Template, error) {
	data, err := ioutil.ReadFile(templatePath)
	if err != nil {
		log.Errorf("read template file %s error: %v", templatePath, err)
		return nil, err
	}

	templates := make([]*api.Template, 0)
	err = yaml.Unmarshal(data, &templates)
	if err != nil {
		log.Errorf("unmarshal yaml template file %s error: %v", templatePath, err)
		return nil, err
	}

	return templates, nil
}
