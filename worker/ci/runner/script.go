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

package runner

import (
	"bytes"
	"encoding/base64"
	"fmt"

	"github.com/caicloud/cyclone/worker/ci/parser"
	docker_client "github.com/fsouza/go-dockerclient"
)

// entrypoint is the default bash entrypoint command
// slice, used to execute a build script in string format.
var entrypoint = []string{"/bin/sh", "-e", "-c"}

// traceScript is a helper script that is added
// to the build script to trace a command.
const traceScript = `
echo %s | base64 -d
%s
`

// Encode encodes the build script as a command in the
// provided Container config. For linux, the build script
// is embedded as the container entrypoint command, base64
// encoded as a one-line script.
func Encode(c *docker_client.CreateContainerOptions, n *parser.DockerNode) {
	var buf bytes.Buffer

	buf.WriteString(writeCmds(n.Commands))

	c.Config.Entrypoint = entrypoint
	c.Config.Cmd = []string{encode(buf.Bytes())}
}

// writeCmds is a helper fuction that writes a slice
// of bash commands as a single script.
func writeCmds(cmds []string) string {
	var buf bytes.Buffer
	for _, cmd := range cmds {
		buf.WriteString(trace(cmd))
	}
	return buf.String()
}

// trace is a helper function that allows us to echo
// commands back to the console for debugging purposes.
func trace(cmd string) string {
	echo := fmt.Sprintf("$ %s\n", cmd)
	base := base64.StdEncoding.EncodeToString([]byte(echo))
	return fmt.Sprintf(traceScript, base, cmd)
}

// encode is a helper function that base64 encodes
// a shell command (or entire script)
func encode(script []byte) string {
	encoded := base64.StdEncoding.EncodeToString(script)
	return fmt.Sprintf("echo %s | base64 -d | /bin/sh", encoded)
}
