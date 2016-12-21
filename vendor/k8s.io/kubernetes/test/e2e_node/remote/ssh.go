/*
Copyright 2016 The Kubernetes Authors.

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

package remote

import (
	"flag"
	"fmt"
	"os/exec"
	"os/user"
	"strings"
	"sync"

	"github.com/golang/glog"
)

var sshOptions = flag.String("ssh-options", "", "Commandline options passed to ssh.")
var sshEnv = flag.String("ssh-env", "", "Use predefined ssh options for environment.  Options: gce")

var sshOptionsMap map[string]string

func init() {
	usr, err := user.Current()
	if err != nil {
		glog.Fatal(err)
	}
	sshOptionsMap = map[string]string{
		"gce": fmt.Sprintf("-i %s/.ssh/google_compute_engine -o UserKnownHostsFile=/dev/null -o IdentitiesOnly=yes -o CheckHostIP=no -o StrictHostKeyChecking=no -o ServerAliveInterval=30 -o LogLevel=ERROR", usr.HomeDir),
	}
}

var hostnameIpOverrides = struct {
	sync.RWMutex
	m map[string]string
}{m: make(map[string]string)}

func AddHostnameIp(hostname, ip string) {
	hostnameIpOverrides.Lock()
	defer hostnameIpOverrides.Unlock()
	hostnameIpOverrides.m[hostname] = ip
}

func GetHostnameOrIp(hostname string) string {
	hostnameIpOverrides.RLock()
	defer hostnameIpOverrides.RUnlock()
	if ip, found := hostnameIpOverrides.m[hostname]; found {
		return ip
	}
	return hostname
}

// getSSHCommand handles proper quoting so that multiple commands are executed in the same shell over ssh
func getSSHCommand(sep string, args ...string) string {
	return fmt.Sprintf("'%s'", strings.Join(args, sep))
}

// SSH executes ssh command with runSSHCommand as root. The `sudo` makes sure that all commands
// are executed by root, so that there won't be permission mismatch between different commands.
func SSH(host string, cmd ...string) (string, error) {
	return runSSHCommand("ssh", append([]string{GetHostnameOrIp(host), "--", "sudo"}, cmd...)...)
}

// SSHNoSudo executes ssh command with runSSHCommand as normal user. Sometimes we need this,
// for example creating a directory that we'll copy files there with scp.
func SSHNoSudo(host string, cmd ...string) (string, error) {
	return runSSHCommand("ssh", append([]string{GetHostnameOrIp(host), "--"}, cmd...)...)
}

// runSSHCommand executes the ssh or scp command, adding the flag provided --ssh-options
func runSSHCommand(cmd string, args ...string) (string, error) {
	if env, found := sshOptionsMap[*sshEnv]; found {
		args = append(strings.Split(env, " "), args...)
	}
	if *sshOptions != "" {
		args = append(strings.Split(*sshOptions, " "), args...)
	}
	output, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("command [%s %s] failed with error: %v", cmd, strings.Join(args, " "), err)
	}
	return string(output), nil
}
