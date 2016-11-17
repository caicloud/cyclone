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
package osutil

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMustGet(t *testing.T) {
	// test string
	var str string
	strEnv := "cyclone_string_env"

	str = GetStringEnv(strEnv, "test")
	assert.Equal(t, str, "test")

	os.Setenv(strEnv, "STRING")
	defer os.Unsetenv(strEnv)

	str = GetStringEnv(strEnv, "")
	assert.Equal(t, str, "STRING")

	// test int
	var i int
	int_env := "cyclone_int_env"

	i = GetIntEnv(int_env, 1)
	assert.Equal(t, i, 1)

	os.Setenv(int_env, "10")
	defer os.Unsetenv(int_env)

	i = GetIntEnv(int_env, 0)
	assert.Equal(t, i, 10)

	// test float
	var f float64
	float_env := "cyclone_float_env"

	f = GetFloat64Env(float_env, 1.1)
	assert.Equal(t, f, 1.1)

	os.Setenv(float_env, "11.1")
	defer os.Unsetenv(float_env)

	f = GetFloat64Env(float_env, 0)
	assert.Equal(t, f, 11.1)

}
