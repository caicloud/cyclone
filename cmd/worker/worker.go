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

package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/golang/glog"
	"gopkg.in/urfave/cli.v1"

	"github.com/caicloud/cyclone/cmd/worker/options"
	_ "github.com/caicloud/cyclone/pkg/scm/provider/github"
	_ "github.com/caicloud/cyclone/pkg/scm/provider/gitlab"
	_ "github.com/caicloud/cyclone/pkg/scm/provider/svn"
	"github.com/caicloud/cyclone/pkg/worker"
)

// newCliApp create a new server cli app
func newCliApp() *cli.App {
	// Log to standard error instead of files.
	flag.Set("logtostderr", "true")

	// Flushes all pending log I/O.
	defer glog.Flush()

	app := cli.NewApp()

	app.Name = "cyclone-worker"

	opts := options.NewWorkerOptions()
	opts.AddFlags(app)

	app.Action = func(c *cli.Context) error {
		glog.Info("worker options: %v", opts)

		worker := worker.NewWorker(opts)
		return worker.Run()
	}

	// sort flags by name
	sort.Sort(cli.FlagsByName(app.Flags))

	return app
}

func main() {
	app := newCliApp()
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
