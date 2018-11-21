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

package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

var v = viper.New()

func init() {
	// Initialize config paths.
	// These paths are searched:
	//   ./
	//   ./config/
	//   {ExecutableFilePath}/
	//   {ExecutableFilePath}/config/
	//   /etc/nirvana/
	//
	// Config file should be one of:
	//   nirvana.yaml
	//   nirvana.toml
	//   nirvana.json
	v.SetConfigName("nirvana")
	file, err := os.Executable()
	if err != nil {
		panic(err)
	}
	path := filepath.Dir(file)
	v.AddConfigPath("./")
	v.AddConfigPath("./config/")
	v.AddConfigPath(path + "/")
	v.AddConfigPath(path + "/config/")
	v.AddConfigPath("/etc/nirvana/")

	// Read configs.
	if err := v.ReadInConfig(); err != nil {
		// Ignore not found error
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			panic(err)
		}
	}
}

// IsSet checks to see if the key has been set in any of the data locations.
// IsSet is case-insensitive for a key
func IsSet(key string) bool {
	return v.IsSet(key)
}

// Set sets the value for the key in the override regiser.
// Set is case-insensitive for a key.
// Will be used instead of values obtained via
// flags, config file, ENV, default, or key/value store.
func Set(key string, value interface{}) {
	v.Set(key, value)
}

// Get can retrieve any value given the key to use.
// Get is case-insensitive for a key.
// Get has the behavior of returning the value associated with the first
// place from where it is set. Viper will check in the following order:
// override, flag, env, config file, key/value store, default
//
// Get returns an interface. For a specific value use one of the Get____ methods.
func Get(key string) interface{} {
	return v.Get(key)
}

// GetBool returns the value associated with the key as a bool.
func GetBool(key string) bool {
	return cast.ToBool(v.Get(key))
}

// GetDuration returns the value associated with the key as a time.Duration.
func GetDuration(key string) time.Duration {
	return cast.ToDuration(v.Get(key))
}

// GetFloat32 returns the value associated with the key as a float32.
func GetFloat32(key string) float32 {
	return cast.ToFloat32(v.Get(key))
}

// GetFloat64 returns the value associated with the key as a float64.
func GetFloat64(key string) float64 {
	return cast.ToFloat64(v.Get(key))
}

// GetInt returns the value associated with the key as a int.
func GetInt(key string) int {
	return cast.ToInt(v.Get(key))
}

// GetInt8 returns the value associated with the key as a int.
func GetInt8(key string) int8 {
	return cast.ToInt8(v.Get(key))
}

// GetInt16 returns the value associated with the key as a int.
func GetInt16(key string) int16 {
	return cast.ToInt16(v.Get(key))
}

// GetInt32 returns the value associated with the key as a int32.
func GetInt32(key string) int32 {
	return cast.ToInt32(v.Get(key))
}

// GetInt64 returns the value associated with the key as a int64.
func GetInt64(key string) int64 {
	return cast.ToInt64(v.Get(key))
}

// GetString returns the value associated with the key as a string.
func GetString(key string) string {
	return cast.ToString(v.Get(key))
}

// GetStringSlice returns the value associated with the key as a []string.
func GetStringSlice(key string) []string {
	return cast.ToStringSlice(v.Get(key))
}

// GetUint returns the value associated with the key as a uint.
func GetUint(key string) uint {
	return cast.ToUint(v.Get(key))
}

// GetUint8 returns the value associated with the key as a uint.
func GetUint8(key string) uint8 {
	return cast.ToUint8(v.Get(key))
}

// GetUint16 returns the value associated with the key as a uint.
func GetUint16(key string) uint16 {
	return cast.ToUint16(v.Get(key))
}

// GetUint32 returns the value associated with the key as a uint32.
func GetUint32(key string) uint32 {
	return cast.ToUint32(v.Get(key))
}

// GetUint64 returns the value associated with the key as a uint64.
func GetUint64(key string) uint64 {
	return cast.ToUint64(v.Get(key))
}
