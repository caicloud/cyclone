package retry

import (
	"k8s.io/apimachinery/pkg/util/wait"
)

// OnError executes the provided function repeatedly, retrying if the server returns an error.
func OnError(backoff wait.Backoff, fn func() error) error {
	var lastErr error
	err := wait.ExponentialBackoff(backoff, func() (bool, error) {
		if err := fn(); err != nil {
			lastErr = err
			return false, nil
		}
		return true, nil
	})
	if err == wait.ErrWaitTimeout {
		return lastErr
	}
	return err
}
