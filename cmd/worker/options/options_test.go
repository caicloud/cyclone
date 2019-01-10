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

package options

import (
	"os"
	"reflect"
	"testing"

	cli "gopkg.in/urfave/cli.v1"
)

func TestWorkerOptionsAddFlags(t *testing.T) {
	opts := NewWorkerOptions()
	app := &cli.App{
		Action: func(c *cli.Context) {
			tests := []struct {
				name string
				got  interface{}
				want interface{}
			}{
				{CycloneServer, opts.CycloneServer, "http://127.0.0.1:7099"},
				{ConsoleWebEndpoint, opts.ConsoleWebEndpoint, "http://127.0.0.1:3000"},
				{RegistryLocation, opts.RegistryLocation, ""},
				{GitlabURL, opts.GitlabURL, "https://gitlab.com"},
				{WorkerImage, opts.WorkerImage, "cargo.caicloud.io/caicloud/cyclone-worker"},
				{ResourceLimitsMemory, opts.Quota[ResourceLimitsMemory], DefaultLimitMemory},
				{ResourceLimitsCPU, opts.Quota[ResourceLimitsCPU], DefaultLimitCPU},
			}

			for _, tt := range tests {
				if !reflect.DeepEqual(tt.got, tt.want) {
					t.Errorf("WorkerOptions.AddFlags(): %s = %v, want %v", tt.name, tt.got, tt.want)
				}
			}
		},
	}
	opts.AddFlags(app)
	app.Run(os.Args)
}
