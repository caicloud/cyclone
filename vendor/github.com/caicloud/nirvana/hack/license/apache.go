/*
Copyright 2017 Caicloud Authors

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
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/caicloud/nirvana/log"
	"github.com/spf13/cobra"
)

var sentinels = []string{
	"Copyright",
	"Caicloud",
	`Licensed under the Apache License, Version 2.0 (the "License");`,
}

var dryRun = false
var root = ""
var goHeaderFile = ""

// LoadGoBoilerplate loads the boilerplate file passed to --go-header-file.
func loadGoBoilerplate(filepath string) ([]byte, error) {
	b, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	b = bytes.Replace(b, []byte("YEAR"), []byte(strconv.Itoa(time.Now().Year())), -1)
	return b, nil
}

// Run ...
func Run() {

	boilerplate, err := loadGoBoilerplate(goHeaderFile)
	if err != nil {
		log.Fatal(err)
		return
	}
	boilerplate = bytes.TrimSpace(boilerplate)
	// add one empty line
	license := append(boilerplate, '\n', '\n')

	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// skip vendor
		if info.IsDir() &&
			(strings.Contains(path, "vendor") || strings.Contains(path, ".git")) {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}

		// skip not go file
		// TODO: support bash and python file
		if ext := filepath.Ext(path); ext != ".go" {
			// log.Infof("Skip file: %s", path)
			return nil
		}

		allFile, err := ioutil.ReadFile(path)
		if err != nil {
			return nil
		}

		src := allFile[:150]

		needLicense := false

		for _, sentinel := range sentinels {
			if !bytes.Contains(src, []byte(sentinel)) {
				needLicense = true
			}
		}

		if needLicense {
			log.Infof("Add License to file: %s", path)

			i := bytes.Index(allFile, []byte("package"))

			if !dryRun {
				if err := ioutil.WriteFile(path, append(license, allFile[i:]...), 0655); err != nil {
					panic(err)
				}
			}
			return nil
		}

		log.Infof("Skip file: %s", path)

		return nil
	})

	if err != nil {
		log.Error(err)
	}
}

func main() {
	cmd := &cobra.Command{
		Use:  "license",
		Long: "Read Apache 2.0 LICENSE content and add to to all go source code header",
		Run: func(cmd *cobra.Command, args []string) {
			Run()
		},
	}
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "")
	cmd.Flags().StringVarP(&root, "root", "r", "./", "")
	cmd.Flags().StringVarP(&goHeaderFile, "go-header-file", "", "", "")

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
