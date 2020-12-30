package retry

import (
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
)

// DefaultRetry is used to retry github operations now.
var DefaultRetry = wait.Backoff{
	Steps:    5,
	Duration: 10 * time.Millisecond,
	Factor:   3.0,
	Jitter:   0.1,
}

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
