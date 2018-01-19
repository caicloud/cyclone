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

package server

import (
	"time"

	cli "gopkg.in/urfave/cli.v1"
)

// api server env
const (
	// Debug mode
	Debug = "DEBUG"
	// MongoDBHost ...
	MongoDBHost = "MONGODB_HOST"
	// KafkaHost ...
	KafkaHost = "KAFKA_HOST"
	// CloudAutoDiscovery ...
	CloudAutoDiscovery = "CLOUD_AUTO_DISCOVERY"
)

// APIServerOptions contains all options(config) for api server
type APIServerOptions struct {
	MongoDBHost         string
	KafkaHost           string
	MongoGracePeriod    time.Duration
	CyclonePort         int
	CycloneAddrTemplate string
	ShowAPIDoc          bool
	Debug               bool
	CloudAutoDiscovery  bool
	LogForceColor       bool
}

// NewAPIServerOptions returns a new APIServerOptions
func NewAPIServerOptions() *APIServerOptions {
	return &APIServerOptions{
		MongoGracePeriod:    30 * time.Second,
		CyclonePort:         7099,
		CycloneAddrTemplate: "http://localhost:%v",
	}
}

// AddFlags adds flags to cli.App
func (opts *APIServerOptions) AddFlags(app *cli.App) {

	flags := []cli.Flag{
		cli.StringFlag{
			Name:        "mongodb-host",
			Value:       "localhost",
			Usage:       "mongdb host",
			EnvVar:      MongoDBHost,
			Destination: &opts.MongoDBHost,
		},
		cli.StringFlag{
			Name:        "kafka-host",
			Value:       "127.0.0.1:9092",
			Usage:       "kafka host",
			EnvVar:      KafkaHost,
			Destination: &opts.KafkaHost,
		},
		cli.BoolTFlag{
			Name:        "show-api-doc",
			Usage:       "show the api doc at http://<cyclone instance>/apidocs/#/api/v0.1",
			Destination: &opts.ShowAPIDoc,
		},
		cli.BoolFlag{
			Name:        "debug",
			Usage:       "Debug mode, default to false",
			EnvVar:      Debug,
			Destination: &opts.Debug,
		},
		cli.BoolTFlag{
			Name:        "cloud-auto-discovery",
			Usage:       "auto discovery cloud by k8s serviceAccount, default to true",
			EnvVar:      CloudAutoDiscovery,
			Destination: &opts.CloudAutoDiscovery,
		},
		cli.BoolFlag{
			Name:        "log-force-color",
			Usage:       "force log use color output",
			Destination: &opts.LogForceColor,
		},
	}

	app.Flags = append(app.Flags, flags...)

}
