package leaderelection

import (
	"net/http"

	"k8s.io/apiserver/pkg/server/healthz"
	"k8s.io/client-go/tools/leaderelection"
)

type healthChecker struct {
	name    string
	checker func(req *http.Request) error
	le      *leaderelection.LeaderElector
}

// Name returns the name of the current checker.
func (c *healthChecker) Name() string {
	return c.name
}

// Check returns an error if health check failed.
func (c *healthChecker) Check(req *http.Request) error {
	if !c.le.IsLeader() {
		// Currently not concerned with the case that we are hot standby
		return nil
	}
	return c.checker(req)
}

// newNamedChecker returns a HealthzChecker implementation with a specific name.
func newNamedChecker(name string, le *leaderelection.LeaderElector, checker func(req *http.Request) error) healthz.HealthChecker {
	return &healthChecker{
		name:    name,
		checker: checker,
		le:      le,
	}
}
