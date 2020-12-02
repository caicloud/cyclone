/*
Copyright 2018 Caicloud Authors

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

package buildutils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/utils/api"
	"github.com/caicloud/nirvana/utils/builder"
	"github.com/caicloud/nirvana/utils/project"
)

// Build finds project config and generates definitions by paths. If there is
// no project config, default config is used (default root path is current path).
// All paths should be under root path.
func Build(paths ...string) (*project.Config, *api.Definitions, error) {
	var config *project.Config
	var err error
	for _, path := range paths {
		config, err = project.LoadDefaultProjectFile(path)
		if err == nil {
			break
		}
	}
	if err != nil {
		dir, err := os.Getwd()
		if err != nil {
			return nil, nil, err
		}
		dir, err = filepath.Abs(dir)
		if err != nil {
			return nil, nil, err
		}
		config = &project.Config{
			Root:        dir,
			Project:     "Unknown Project",
			Description: "This project does not have a project config.",
			Schemes:     []string{"http"},
			Hosts:       []string{"localhost"},
			Versions: []project.Version{
				{
					Name: "unversioned",
					PathRules: []project.PathRule{
						{
							Prefix: "/",
						},
					},
				},
			},
		}
		log.Warning("can't find project file, instead by default config")
	}
	for _, path := range paths {
		dir, err := filepath.Abs(path)
		if err != nil {
			return nil, nil, err
		}
		if !strings.HasPrefix(dir, config.Root) {
			return nil, nil, fmt.Errorf("path %s is not in root dir %s", path, config.Root)
		}
	}
	builder := builder.NewAPIBuilder(config.Root, project.Subdirectories(false, paths...)...)
	definitions, err := builder.Build()
	if err != nil {
		return nil, nil, err
	}
	return config, definitions, nil
}
