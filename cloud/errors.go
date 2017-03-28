package cloud

import (
	"errors"
	"fmt"
)

var (
	// ErrNoEnoughResource occurs when clouds have no enough resource to provison as worker
	ErrNoEnoughResource = errors.New("worker required resources are out of quota limit")
)

// ErrCloudProvision contains all clouds provision errors
type ErrCloudProvision struct {
	// cloud name maps to err
	errs map[string]error
}

// NewErrCloudProvision return a new CloudProvisionErr
func NewErrCloudProvision() *ErrCloudProvision {
	return &ErrCloudProvision{
		errs: make(map[string]error),
	}
}

// Err returns nil if CloudProvisionErr contains 0 error
func (cpe *ErrCloudProvision) Err() error {
	if len(cpe.errs) == 0 {
		return nil
	}
	return cpe
}

// Add adds an error to CloudProvisionErr
func (cpe *ErrCloudProvision) Add(name string, err error) {
	cpe.errs[name] = err
}

func (cpe *ErrCloudProvision) Error() string {
	str := "the following clouds can not provison workers:\n"
	for name, err := range cpe.errs {
		str += fmt.Sprintf("  cloud[%s]: %v\n", name, err)
	}
	return str
}

// IsAllCloudsBusyErr check whether all cloud is too busy to provision a worker
func IsAllCloudsBusyErr(errs error) bool {
	if perr, ok := errs.(*ErrCloudProvision); ok {
		for _, err := range perr.errs {
			if err != ErrNoEnoughResource {
				return false
			}
		}
		return true
	}

	if errs == ErrNoEnoughResource {
		return true
	}
	return false
}
