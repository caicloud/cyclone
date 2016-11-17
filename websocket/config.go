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

package websocket

const (
	// Localhost is the ip of localhost.
	Localhost = "127.0.0.1"
	// WSServerPort is the port of websocket server.
	WSServerPort = 8000
	// MaxConnectionNumber is the default maximum number of connections.
	MaxConnectionNumber = 4096
	// IdleSessionTimeOut is the default timeout for idle sessions.
	IdleSessionTimeOut = 120
	// IdleCheckInterval is the default interval for idle checks.
	IdleCheckInterval = 30
)

//ServerConfig is the config of push-log websocket server
type ServerConfig struct {
	// ServerIP is the IP of server.
	ServerIP string `yml:"ServerIp"`
	// Port is the port of server.
	Port int `yml:"ServerPort"`
	// MaxConnectionNumber is maximum number of connections.
	MaxConnectionNumber int `yml:"MaxConnectionNumber"`
	// IdleSessionTimeOut is the timeout for idle sessions.
	IdleSessionTimeOut int64 `yml:"IdleSessionTimeOut"`
	// IdleCheckInterval is the interval for idle checks.
	IdleCheckInterval int `yml:"IdleCheckInterval"`
	// ServerCertificate is the certificate of server.
	ServerCertificate string `yml:"ServerCertificate"`
	// ServerKey is the key of server.
	ServerKey string `yml:"ServerKey"`
}

var mScServer *ServerConfig

//GetConfig get config of push-log websocket server
func GetConfig() *ServerConfig {
	return mScServer
}

//LoadServerConfig load config of push-log websocket server
func LoadServerConfig() error {
	mScServer = &ServerConfig{
		ServerIP:            Localhost,
		Port:                WSServerPort,
		MaxConnectionNumber: MaxConnectionNumber,
		IdleSessionTimeOut:  IdleSessionTimeOut,
		IdleCheckInterval:   IdleCheckInterval,
		ServerCertificate:   "",
		ServerKey:           "",
	}
	return nil
}
