package retry

import (
	"k8s.io/apimachinery/pkg/util/wait"
)

// OnError executes the provided function repeatedly, retrying if the server returns a error.
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
