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
	"github.com/caicloud/nirvana/config"
	"github.com/caicloud/nirvana/examples/api-basic/api/v1"
	"github.com/caicloud/nirvana/examples/api-basic/api/v2"
	"github.com/caicloud/nirvana/log"
)

func main() {
	cmd := config.NewDefaultNirvanaCommand()
	if err := cmd.Execute(v1.Descriptor(), v2.Descriptor()); err != nil {
		log.Fatal(err)
	}
}
