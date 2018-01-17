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

package cloud

import (
	"os"
	"reflect"
	"testing"

	cli "gopkg.in/urfave/cli.v1"
)

func TestWorkerOptions_DeepCopy(t *testing.T) {
	opts := NewWorkerOptions()
	dep := opts.DeepCopy()
	deep := &dep

	if !reflect.DeepEqual(opts, deep) {
		t.Errorf("WorkerOptions.DeepCopy() = %v, want %v", deep, opts)
	}
	if deep == opts {
		t.Errorf("WorkerOptions.DeepCopy() = %v, want %v", &deep, &opts)
	}

	// change it
	deep.Quota[ResourceLimitsCPU] = ZeroQuantity
	deep.Quota[ResourceLimitsMemory] = ZeroQuantity

	if reflect.DeepEqual(opts, deep) {
		t.Errorf("WorkerOptions.DeepCopy() = %v, want %v", deep, opts)
	}

}

func TestWorkerOptions_AddFlags(t *testing.T) {

	opts := NewWorkerOptions()
	app := &cli.App{
		Action: func(c *cli.Context) {
			tests := []struct {
				name string
				got  interface{}
				want interface{}
			}{
				{CycloneServer, opts.WorkerEnvs.CycloneServer, "http://127.0.0.1:7099"},
				{ConsoleWebEndpoint, opts.WorkerEnvs.ConsoleWebEndpoint, "http://127.0.0.1:3000"},
				{RegistryLocation, opts.WorkerEnvs.RegistryLocation, "cargo.caicloud.io"},
				{GitlabURL, opts.WorkerEnvs.GitlabURL, "https://gitlab.com"},
				{WorkerImage, opts.WorkerEnvs.WorkerImage, "cargo.caicloud.io/caicloud/cyclone-worker"},
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
