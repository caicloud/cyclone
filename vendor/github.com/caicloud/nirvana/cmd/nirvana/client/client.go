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

package client

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/caicloud/nirvana/cmd/nirvana/buildutils"
	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/utils/generators/golang"
	"github.com/caicloud/nirvana/utils/project"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func newClientCommand() *cobra.Command {
	options := &clientOptions{}
	cmd := &cobra.Command{
		Use:   "client /path/to/apis",
		Short: "Create client for project",
		Long:  options.Manuals(),
		Run: func(cmd *cobra.Command, args []string) {
			if err := options.Validate(cmd, args); err != nil {
				log.Fatalln(err)
			}
			if err := options.Run(cmd, args); err != nil {
				log.Fatalln(err)
			}
		},
	}
	options.Install(cmd.PersistentFlags())
	return cmd
}

type clientOptions struct {
	Output string
	Rest   string
}

func (o *clientOptions) Install(flags *pflag.FlagSet) {
	flags.StringVar(&o.Output, "output", "./client", "Output directory for generated client")
	flags.StringVar(&o.Rest, "rest", "github.com/caicloud/nirvana/rest", "Package of rest client")
}

func (o *clientOptions) Validate(cmd *cobra.Command, args []string) error {
	if o.Output == "" {
		return fmt.Errorf("must specify generated client path")
	}
	return nil
}

func (o *clientOptions) Run(cmd *cobra.Command, args []string) error {
	if len(args) <= 0 {
		defaultAPIsPath := "pkg/apis"
		args = append(args, defaultAPIsPath)
		log.Infof("No packages are specified, defaults to %s", defaultAPIsPath)
	}

	config, definitions, err := buildutils.Build(args...)
	if err != nil {
		return err
	}

	log.Infof("Project root directory is %s", config.Root)

	pkg, err := project.PackageForPath(o.Output)
	if err != nil {
		return err
	}

	rootPkg, err := project.PackageForPath(config.Root)
	if err != nil {
		return err
	}

	generator := golang.NewGenerator(config, definitions, o.Rest, pkg, rootPkg)
	files, err := generator.Generate()
	if err != nil {
		return err
	}

	for path, data := range files {
		path = filepath.Join(o.Output, path+".go")
		dir := filepath.Dir(path)
		_, err := os.Stat(dir)
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0775); err != nil {
				return fmt.Errorf("can't create directory %s: %v", dir, err)
			}
		}
		if err := ioutil.WriteFile(path, data, 0664); err != nil {
			return err
		}
	}
	log.Infof("Generated golang client package %s", pkg)
	return nil
}

func (o *clientOptions) Manuals() string {
	return ""
}
