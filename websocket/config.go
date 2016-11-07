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
	Localhost           = "127.0.0.1"
	WSServerPort        = 8000
	MaxConnectionNumber = 4096
	IdleSessionTimeOut  = 120
	IdleCheckInterval   = 30
)

//ServerConfig is the config of push-log websocket server
type ServerConfig struct {
	ServerIp            string `yml:"ServerIp"`
	Port                int    `yml:"ServerPort"`
	MaxConnectionNumber int    `yml:"MaxConnectionNumber"`
	IdleSessionTimeOut  int64  `yml:"IdleSessionTimeOut"`
	IdleCheckInterval   int    `yml:"IdleCheckInterval"`
	ServerCertificate   string `yml:"ServerCertificate"`
	ServerKey           string `yml:"ServerKey"`
}

var m_scServer *ServerConfig

//GetConfig get config of push-log websocket server
func GetConfig() *ServerConfig {
	return m_scServer
}

//LoadServerConfig load config of push-log websocket server
func LoadServerConfig() error {
	m_scServer = &ServerConfig{
		ServerIp:            Localhost,
		Port:                WSServerPort,
		MaxConnectionNumber: MaxConnectionNumber,
		IdleSessionTimeOut:  IdleSessionTimeOut,
		IdleCheckInterval:   IdleCheckInterval,
		ServerCertificate:   "",
		ServerKey:           "",
	}
	return nil
}
