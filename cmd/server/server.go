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
	"fmt"
	"os"
	"sort"

	"gopkg.in/urfave/cli.v1"

	_ "github.com/caicloud/cyclone/pkg/scm/provider"
)

// NeverStop may be passed to Until to make it never stop.
var NeverStop <-chan struct{} = make(chan struct{})

// RunServer starts an api server
func RunServer(opts *ServerOptions, stopCh <-chan struct{}) error {
	s := opts.NewAPIServer()
	ps, err := s.PrepareRun()
	if err != nil {
		return err
	}
	ps.Run(stopCh)
	return nil
}

// newCliApp create a new server cli app
func newCliApp() *cli.App {

	app := cli.NewApp()

	app.Name = "cyclone"

	opts := NewServerOptions()
	opts.AddFlags(app)

	app.Action = func(c *cli.Context) error {
		// start server
		return RunServer(opts, NeverStop)
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
