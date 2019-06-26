package retry

import (
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
)

// OnError executes the provided function repeatedly, retrying if the server returns an error.
func OnError(backoff wait.Backoff, fn func() error) error {
	var lastErr error
	err := wait.ExponentialBackoff(backoff, func() (bool, error) {
		err := fn()
		switch {
		case err == nil:
			return true, nil
		default:
			lastErr = err
			return false, nil
		}
	})
	if err == wait.ErrWaitTimeout {
		err = lastErr
	}
	return err
}

// OnExceededQuota executes the provided function repeatedly, retrying if the server returns an exceeded quota error.
func OnExceededQuota(backoff wait.Backoff, fn func() error) error {
	var lastErr error
	err := wait.ExponentialBackoff(backoff, func() (bool, error) {
		err := fn()
		switch {
		case err == nil:
			return true, nil
		case errors.IsForbidden(err):
			if strings.Contains(err.Error(), "exceeded quota:") {
				lastErr = err
				return false, nil
			}
			return false, err
		default:
			return false, err
		}
	})
	if err == wait.ErrWaitTimeout {
		err = lastErr
	}
	return err
}
