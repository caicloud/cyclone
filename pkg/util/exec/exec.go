package exec

import "os/exec"

// RunInDir runs a command in given directory. The name and args are as in
// exec.Command. stdout, stderr, and the environment are inherited from the
// current process.
func RunInDir(dir, name string, args ...string) ([]byte, error) {
	c := exec.Command(name, args...)
	c.Dir = dir
	return c.CombinedOutput()
}
