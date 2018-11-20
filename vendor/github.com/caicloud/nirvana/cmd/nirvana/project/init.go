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

package project

import (
	"bytes"
	"fmt"
	"go/build"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/caicloud/nirvana/cmd/nirvana/utils"
	"github.com/caicloud/nirvana/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func newInitCommand() *cobra.Command {
	options := &initOptions{}
	cmd := &cobra.Command{
		Use:   "init /path/to/project",
		Short: "Create a basic project structure",
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

type templateData struct {
	Boilerplate    string
	ProjectAbsDir  string
	ProjectPackage string
	ProjectName    string
	Version        string
	Registries     string
	ImagePrefix    string
	ImageSuffix    string
	BuildImage     string
	RuntimeImage   string
}

// GoBoilerplate returns boilerplate in go style.
func (t *templateData) GoBoilerplate() string {
	return "/*\n" + t.Boilerplate + "\n*/\n"
}

// SharpBoilerplate returns boilerplate in sharp style.
func (t *templateData) SharpBoilerplate() string {
	return "# " + strings.Replace(t.Boilerplate, "\n", "\n# ", -1)
}

type initOptions struct {
	Boilerplate  string
	Version      string
	Registries   []string
	ImagePrefix  string
	ImageSuffix  string
	BuildImage   string
	RuntimeImage string
}

func (o *initOptions) Install(flags *pflag.FlagSet) {
	flags.StringVar(&o.Boilerplate, "boilerplate", "", "Path to boilerplate")
	flags.StringVar(&o.Version, "version", "v0.1.0", "First version of the project")
	flags.StringSliceVar(&o.Registries, "registries", []string{}, "Docker image registries")
	flags.StringVar(&o.ImagePrefix, "image-prefix", "", "Docker image prefix")
	flags.StringVar(&o.ImageSuffix, "image-suffix", "", "Docker image suffix")
	flags.StringVar(&o.BuildImage, "build-image", "golang:latest", "Golang image for building the project")
	flags.StringVar(&o.RuntimeImage, "runtime-image", "debian:jessie", "Docker base image for running the project")

}

func (o *initOptions) Validate(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("must specify a project path")
	}
	if len(args) > 1 {
		return fmt.Errorf("must not specify multiple project paths")
	}
	return nil
}

func (o *initOptions) Run(cmd *cobra.Command, args []string) error {
	pathToProject, err := filepath.Abs(args[0])
	if err != nil {
		return fmt.Errorf("can't get absolute path for %s: %v", args[0], err)
	}
	projectName := filepath.Base(pathToProject)

	registries := strings.Join(o.Registries, " ")
	if registries == "" {
		registries = `""`
	}

	td := &templateData{
		ProjectAbsDir: filepath.Dir(pathToProject),
		ProjectName:   projectName,
		Version:       o.Version,
		Registries:    registries,
		ImagePrefix:   o.ImagePrefix,
		ImageSuffix:   o.ImageSuffix,
		BuildImage:    o.BuildImage,
		RuntimeImage:  o.RuntimeImage,
	}
	td.ProjectPackage, err = utils.PackageForPath(pathToProject)
	if err != nil {
		return err
	}

	if td.ProjectPackage == "" {
		return fmt.Errorf("project %s is not in GOPATH %s", pathToProject, build.Default.GOPATH)
	}

	if o.Boilerplate != "" {
		data, err := ioutil.ReadFile(o.Boilerplate)
		if err != nil {
			return fmt.Errorf("can't read boilerplate file %s: %v", o.Boilerplate, err)
		}
		data = bytes.Replace(data, []byte("YEAR"), []byte(strconv.Itoa(time.Now().Year())), -1)
		data = bytes.TrimSpace(data)
		td.Boilerplate = string(data)
	}

	directories := o.directories(projectName)
	for i, dir := range directories {
		dir = filepath.Join(pathToProject, dir)
		_, err := os.Stat(dir)
		if !os.IsNotExist(err) {
			if err != nil {
				return fmt.Errorf("can't get stat for %s: %v", dir, err)
			}
			return fmt.Errorf("%s already exists", dir)
		}
		directories[i] = dir
	}
	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0775); err != nil {
			return fmt.Errorf("can't create directory %s: %v", dir, err)
		}
	}

	files := map[string][]byte{}
	for file, tpl := range o.templates(projectName) {
		tmpl, err := template.New(file).Parse(tpl)
		if err != nil {
			return fmt.Errorf("can't create template %s: %v", file, err)
		}
		buf := bytes.NewBuffer(nil)
		if err := tmpl.Execute(buf, td); err != nil {
			return fmt.Errorf("can't execute template %s: %v", file, err)
		}
		if strings.HasSuffix(file, ".go") {
			files[file], err = format.Source(buf.Bytes())
			if err != nil {
				return fmt.Errorf("can't format go source file %s: %v", file, err)
			}
		} else {
			files[file] = buf.Bytes()
		}
	}
	for file, data := range files {
		file = filepath.Join(pathToProject, file)
		if err := ioutil.WriteFile(file, data, 0664); err != nil {
			return fmt.Errorf("can't write file %s: %v", file, err)
		}
	}
	return nil
}

func (o *initOptions) directories(project string) []string {
	return []string{
		"bin",
		fmt.Sprintf("cmd/%s", project),
		fmt.Sprintf("build/%s", project),
		"pkg/api/filters",
		"pkg/api/middlewares",
		"pkg/api/modifiers",
		"pkg/api/v1/descriptors",
		"pkg/api/v1/converters",
		"pkg/api/v1/middlewares",
		"pkg/message",
		"pkg/version",
		"vendor",
	}
}

func (o *initOptions) templates(project string) map[string]string {
	return map[string]string{
		fmt.Sprintf("cmd/%s/main.go", project):      o.templateMain(),
		fmt.Sprintf("build/%s/Dockerfile", project): o.templateDockerfile(),
		"pkg/api/filters/filters.go":                o.templateFilters(),
		"pkg/api/middlewares/middlewares.go":        o.templateMiddlewares(),
		"pkg/api/modifiers/modifiers.go":            o.templateModifiers(),
		"pkg/api/v1/descriptors/descriptors.go":     o.templateDescriptors(),
		"pkg/api/v1/descriptors/message.go":         o.templateMessageAPI(),
		"pkg/api/v1/middlewares/middlewares.go":     o.templateMiddlewares(),
		"pkg/api/api.go":                            o.templateAPI(),
		"pkg/message/message.go":                    o.templateMessage(),
		"pkg/version/version.go":                    o.templateVersion(),
		"Gopkg.toml":                                o.templateGopkg(),
		"Makefile":                                  o.templateMakefile(),
		"nirvana.yaml":                              o.templateProject(),
		"README.md":                                 o.templateReadme(),
	}
}

func (o *initOptions) templateMain() string {
	return `
{{- if .Boilerplate -}}
{{ .GoBoilerplate }}
{{- end }}

package main

import (
	"fmt"

	"{{ .ProjectPackage }}/pkg/api"
	"{{ .ProjectPackage }}/pkg/api/filters"
	"{{ .ProjectPackage }}/pkg/api/modifiers"
	"{{ .ProjectPackage }}/pkg/version"

	"github.com/caicloud/nirvana"
	"github.com/caicloud/nirvana/config"
	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/plugins/logger"
	"github.com/caicloud/nirvana/plugins/metrics"
	"github.com/caicloud/nirvana/plugins/reqlog"
)

func main() {
	// Print nirvana banner.
	fmt.Println(nirvana.Logo, nirvana.Banner)

	// Create nirvana command.
	cmd := config.NewNamedNirvanaCommand("server", config.NewDefaultOption())

	// Create plugin options.
	metricsOption := metrics.NewDefaultOption() // Metrics plugin.
	loggerOption := logger.NewDefaultOption()   // Logger plugin.
	reqlogOption := reqlog.NewDefaultOption()   // Request log plugin.

	// Enable plugins.
	cmd.EnablePlugin(metricsOption, loggerOption, reqlogOption)

	// Create server config.
	serverConfig := nirvana.NewConfig()

	// Configure APIs. These configurations may be changed by plugins.
	serverConfig.Configure(
		nirvana.Logger(log.DefaultLogger()), // Will be changed by logger plugin.
		nirvana.Filter(filters.Filters()...),
		nirvana.Modifier(modifiers.Modifiers()...),
		nirvana.Descriptor(api.Descriptor()),
	)

	// Set nirvana command hooks.
	cmd.SetHook(&config.NirvanaCommandHookFunc{
		PreServeFunc: func(config *nirvana.Config, server nirvana.Server) error {
			// Output project information.
			config.Logger().Infof("Package:%s Version:%s Commit:%s", version.Package, version.Version, version.Commit)
			return nil
		},
	})

	// Start with server config.
	if err := cmd.ExecuteWithConfig(serverConfig); err != nil {
		serverConfig.Logger().Fatal(err)
	}
}
`
}

func (o *initOptions) templateDescriptors() string {
	return `
{{- if .Boilerplate -}}
{{ .GoBoilerplate }}
{{- end }}

package descriptors

import (
	"{{ .ProjectPackage }}/pkg/api/v1/middlewares"

	def "github.com/caicloud/nirvana/definition"
)

// descriptors describe APIs of current version.
var descriptors []def.Descriptor

// register registers descriptors.
func register(ds ...def.Descriptor) {
	descriptors = append(descriptors, ds...)
}

// Descriptor returns a combined descriptor for current version.
func Descriptor() def.Descriptor {
	return def.Descriptor{
		Description: "v1 APIs",
		Path:        "/v1",
		Middlewares: middlewares.Middlewares(),
		Children:    descriptors,
	}
}
`
}

func (o *initOptions) templateFilters() string {
	return `
{{- if .Boilerplate -}}
{{ .GoBoilerplate }}
{{- end }}

package filters

import "github.com/caicloud/nirvana/service"

// Filters returns a list of filters.
func Filters() []service.Filter {
	return []service.Filter{
		service.RedirectTrailingSlash(),
		service.FillLeadingSlash(),
		service.ParseRequestForm(),
	}
}
`
}

func (o *initOptions) templateMiddlewares() string {
	return `
{{- if .Boilerplate -}}
{{ .GoBoilerplate }}
{{- end }}

package middlewares

import def "github.com/caicloud/nirvana/definition"

// Middlewares returns a list of middlewares.
func Middlewares() []def.Middleware {
	return []def.Middleware{}
}
`
}

func (o *initOptions) templateModifiers() string {
	return `
{{- if .Boilerplate -}}
{{ .GoBoilerplate }}
{{- end }}

// +nirvana:api=modifiers:"Modifiers"

package modifiers

import "github.com/caicloud/nirvana/service"

// Modifiers returns a list of modifiers.
func Modifiers() []service.DefinitionModifier {
	return []service.DefinitionModifier{
		service.FirstContextParameter(),
		service.ConsumeAllIfConsumesIsEmpty(),
		service.ProduceAllIfProducesIsEmpty(),
		service.ConsumeNoneForHTTPGet(),
		service.ConsumeNoneForHTTPDelete(),
		service.ProduceNoneForHTTPDelete(),
	}
}
`
}

func (o *initOptions) templateAPI() string {
	return `
{{- if .Boilerplate -}}
{{ .GoBoilerplate }}
{{- end }}

// +nirvana:api=descriptors:"Descriptor"

package api

import (
	"{{ .ProjectPackage }}/pkg/api/middlewares"
	v1 "{{ .ProjectPackage }}/pkg/api/v1/descriptors"

	def "github.com/caicloud/nirvana/definition"
)

// Descriptor returns a combined descriptor for APIs of all versions.
func Descriptor() def.Descriptor {
	return def.Descriptor{
		Description: "APIs",
		Path:        "/api",
		Middlewares: middlewares.Middlewares(),
		Consumes:    []string{def.MIMEJSON},
		Produces:    []string{def.MIMEJSON},
		Children: []def.Descriptor{
			v1.Descriptor(),
		},
	}
}
`
}

func (o *initOptions) templateMessageAPI() string {
	return `
{{- if .Boilerplate -}}
{{ .GoBoilerplate }}
{{- end }}

package descriptors

import (
	"{{ .ProjectPackage }}/pkg/message"

	def "github.com/caicloud/nirvana/definition"
)

func init() {
	register([]def.Descriptor{{ print "{{" }}
		Path:        "/messages",
		Definitions: []def.Definition{listMessages},
	}, {
		Path:        "/messages/{message}",
		Definitions: []def.Definition{getMessage},
	},
	}...)
}

var listMessages = def.Definition{
	Method:   def.Get,
	Summary: "List Messages",
	Description: "Query a specified number of messages and returns an array",
	Function: message.ListMessages,
	Parameters: []def.Parameter{
		{
			Source:      def.Query,
			Name:        "count",
			Default:     10,
			Description: "Number of messages",
		},
	},
	Results: def.DataErrorResults("A list of messages"),
}

var getMessage = def.Definition{
	Method:   def.Get,
	Summary: "Get Message",
	Description: "Get a message by id",
	Function: message.GetMessage,
	Parameters: []def.Parameter{
		def.PathParameterFor("message", "Message id"),
	},
	Results: def.DataErrorResults("A message"),
}
`
}

func (o *initOptions) templateMessage() string {
	return `
{{- if .Boilerplate -}}
{{ .GoBoilerplate }}
{{- end }}

package message

import (
	"context"
	"fmt"
)

// Message describes a message entry.
type Message struct {
	ID      int    ` + "`json:\"id\"`" + `
	Title   string ` + "`json:\"title\"`" + `
	Content string ` + "`json:\"content\"`" + `
}

// ListMessages returns all messages.
func ListMessages(ctx context.Context, count int) ([]Message, error) {
	messages := make([]Message, count)
	for i := 0; i < count; i++ {
		messages[i].ID = i
		messages[i].Title = fmt.Sprintf("Example %d", i)
		messages[i].Content = fmt.Sprintf("Content of example %d", i)
	}
	return messages, nil
}

// GetMessage return a message by id.
func GetMessage(ctx context.Context, id int) (*Message, error) {
	return &Message{
		ID:      id,
		Title:   "This is an example",
		Content: "Example content",
	}, nil
}
`
}

func (o *initOptions) templateVersion() string {
	return `
{{- if .Boilerplate -}}
{{ .GoBoilerplate }}
{{- end }}

package version

// Following values should be substituted with a real value during build.
var (
	Version = "Unknown"
	Commit = "Unknown"
	Package = "{{ .ProjectPackage }}"
)
`
}

func (o *initOptions) templateDockerfile() string {
	return `
{{- if .Boilerplate -}}
{{ .SharpBoilerplate }}

{{ end -}}
FROM {{ .BuildImage }}

WORKDIR /go/src/{{ .ProjectPackage }}

COPY . .

ENV GOPATH /go

ARG CMD_DIR=./cmd

ARG ROOT={{ .ProjectPackage }}

ARG TARGET={{ .ProjectName }}

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64                      \
	go build -i -v -o /tmp/${TARGET}                  \
	-ldflags "-s -w -X ${ROOT}/pkg/version.Version=${VERSION}  \
	-X ${ROOT}/pkg/version.Commit=${COMMIT}                    \
	-X ${ROOT}/pkg/version.Package=${ROOT}"                    \
	${CMD_DIR}/${TARGET};

FROM {{ .RuntimeImage }}

ARG TARGET={{ .ProjectName }}

COPY --from=0 /tmp/${TARGET} /${TARGET}

ENTRYPOINT [/${TARGET}]
`
}

func (o *initOptions) templateGopkg() string {
	return `
{{- if .Boilerplate -}}
{{ .SharpBoilerplate }}
#
{{ end -}}
# Gopkg.toml example
#
# Refer to https://github.com/golang/dep/blob/master/docs/Gopkg.toml.md
# for detailed Gopkg.toml documentation.
#
# required = ["github.com/user/thing/cmd/thing"]
# ignored = ["github.com/user/project/pkgX", "bitbucket.org/user/project/pkgA/pkgY"]
#
# [[constraint]]
#   name = "github.com/user/project"
#   version = "1.0.0"
#
# [[constraint]]
#   name = "github.com/user/project2"
#   branch = "dev"
#   source = "github.com/myfork/project2"
#
# [[override]]
#   name = "github.com/x/y"
#   version = "2.4.0"
#
# [prune]
#   non-go = false
#   go-tests = true
#   unused-packages = true

required = ["github.com/caicloud/nirvana/utils/api"]

[prune]
  go-tests = true
  unused-packages = true

[[override]]
  name = "github.com/caicloud/nirvana"
  branch = "master"
`
}

func (o *initOptions) templateMakefile() string {
	return `
{{- if .Boilerplate -}}
{{ .SharpBoilerplate }}
#
{{ end -}}
# The old school Makefile, following are required targets. The Makefile is written
# to allow building multiple binaries. You are free to add more targets or change
# existing implementations, as long as the semantics are preserved.
#
#   make              - default to 'build' target
#   make test         - run unit test
#   make build        - build local binary targets
#   make container    - build containers
#   make push         - push containers
#   make clean        - clean up targets
#
# The makefile is also responsible to populate project version information.

#
# Tweak the variables based on your project.
#

# Current version of the project.
VERSION ?= {{ .Version }}

# Target binaries. You can build multiple binaries for a single project.
TARGETS := {{ .ProjectName }}

# Container registries.
REGISTRIES ?= {{ .Registries }}

# Container image prefix and suffix added to targets.
# The final built images are:
#   $[REGISTRY]$[IMAGE_PREFIX]$[TARGET]$[IMAGE_SUFFIX]:$[VERSION]
# $[REGISTRY] is an item from $[REGISTRIES], $[TARGET] is an item from $[TARGETS].
IMAGE_PREFIX ?= $(strip {{ .ImagePrefix }})
IMAGE_SUFFIX ?= $(strip {{ .ImageSuffix }})

# This repo's root import path (under GOPATH).
ROOT := {{ .ProjectPackage }}

# Project main package location (can be multiple ones).
CMD_DIR := ./cmd

# Project output directory.
OUTPUT_DIR := ./bin

# Build direcotory.
BUILD_DIR := ./build

# Git commit sha.
COMMIT := $(strip $(shell git rev-parse --short HEAD 2>/dev/null))
COMMIT := $(if $(COMMIT),$(COMMIT),"Unknown")

#
# Define all targets. At least the following commands are required:
#

.PHONY: build container push test clean

build:
	@for target in $(TARGETS); do                                                      \
	  go build -i -v -o $(OUTPUT_DIR)/$${target}                                       \
	    -ldflags "-s -w -X $(ROOT)/pkg/version.Version=$(VERSION)                      \
	    -X $(ROOT)/pkg/version.Commit=$(COMMIT)                                        \
	    -X $(ROOT)/pkg/version.Package=$(ROOT)"                                        \
	    $(CMD_DIR)/$${target};                                                         \
	done

container:
	@for target in $(TARGETS); do                                                      \
	  for registry in $(REGISTRIES); do                                                \
	    image=$(IMAGE_PREFIX)$${target}$(IMAGE_SUFFIX);                                \
	    docker build -t $${registry}$${image}:$(VERSION)                               \
	      --build-arg ROOT=$(ROOT) --build-arg TARGET=$${target}                       \
	      --build-arg CMD_DIR=$(CMD_DIR)                                               \
	      -f $(BUILD_DIR)/$${target}/Dockerfile .;                                     \
	  done                                                                             \
	done

push: container
	@for target in $(TARGETS); do                                                      \
	  for registry in $(REGISTRIES); do                                                \
	    image=$(IMAGE_PREFIX)$${target}$(IMAGE_SUFFIX);                                \
	    docker push $${registry}$${image}:$(VERSION);                                  \
	  done                                                                             \
	done

test:
	@go test ./...

clean:
	@rm -vrf ${OUTPUT_DIR}/*
`
}

func (o *initOptions) templateProject() string {
	return `
{{- if .Boilerplate -}}
{{ .SharpBoilerplate }}
#
{{ end -}}
# This file describes your project. It's used to generate api docs and
# clients. All fields in this file won't affect nirvana configurations.

project: {{ .ProjectName }}
description: This project uses nirvana as API framework
schemes:
- http
hosts:
- localhost:8080
contacts:
- name: nobody
  email: nobody@nobody.io
  description: Maintain this project
versions:
- name: v1
  description: The v1 version is the first version of this project
  rules:
  - "^/api/v1.*"
`
}

func (o *initOptions) templateReadme() string {
	return `
# Project {{ .ProjectName }}

<!-- Write one paragraph of this project description here -->

## Getting Started

### Prerequisites

<!-- Describe packages, tools and everything we needed here -->

### Building

<!-- Describe how to build this project -->

### Running

<!-- Describe how to run this project -->

## Versioning

<!-- Place versions of this project and write comments for every version -->

## Contributing

<!-- Tell others how to contribute this project -->

## Authors

<!-- Put authors here -->

## License

<!-- A link to license file -->

`
}

func (o *initOptions) Manuals() string {
	return `
This command generates standard nirvana project structure.
.
├── bin
├── build
│   └── <project-name>
│       └── Dockerfile
├── cmd
│   └── <project-name>
│       └── main.go
├── pkg
│   ├── api
│   │   ├── api.go
│   │   ├── middlewares
│   │   │   └── middlewares.go
│   │   └── v1
│   │       ├── converters
│   │       ├── descriptors
│   │       │   ├── descriptors.go
│   │       │   └── message.go
│   │       └── middlewares
│   │           └── middlewares.go
│   ├── message
│   │   └── message.go
│   └── version
│       └── version.go
├── vendor
├── Gopkg.toml
├── nirvana.yaml
├── Makefile
└── README.md
`
}
