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

package http

import (
	"html/template"
	"net/http"

	"github.com/caicloud/cyclone/pkg/log"
	"github.com/caicloud/cyclone/pkg/osutil"
)

const (
	LOG_HTML_TEMPLATE = "LOG_HTML_TEMPLATE"
)

// Server is the type for log server.
func Server() {
	log.Info("http server start")
	http.HandleFunc("/log", logHandler)
}

func logHandler(w http.ResponseWriter, r *http.Request) {
	filePath := osutil.GetStringEnv(LOG_HTML_TEMPLATE, "/http/web/log.html")
	t, err := template.ParseFiles(filePath)
	if err != nil {
		log.Error(err)
	}
	t.Execute(w, nil)
}
