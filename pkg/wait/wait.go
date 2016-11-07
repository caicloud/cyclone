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

package wait

import (
	"errors"
	"time"
)

// WaitFunc creates a channel that receives an item every time a test
// should be executed and is closed when the last test should be invoked.
type WaitFunc func(done <-chan struct{}) <-chan struct{}

// ConditionFunc returns true if the condition is satisfied, or an error
// if the loop should be aborted.
type ConditionFunc func() (done bool, err error)

// NoErrorConditionFunc returns the error, or return nil if the loop
// should be aborted.
type NoErrorConditionFunc func() error

// Poll tries a condition func until it returns true, an error, or the timeout
// is reached. condition will always be invoked at least once but some intervals
// may be missed if the condition takes too long or the time window is too short.
// If you want to Poll something forever, see PollInfinite.
// Poll always waits the interval before the first check of the condition.
func Poll(interval, timeout time.Duration, condition ConditionFunc) error {
	done := make(chan struct{})
	defer close(done)
	return waitfor(poller(interval, timeout), condition, done)
}

// PollUntilNoError tries a condition func until it returns no error, or the timeout
// is reached.
func PollUntilNoError(interval, timeout time.Duration, condition NoErrorConditionFunc) error {
	done := make(chan struct{})
	defer close(done)
	return waitforNoError(poller(interval, timeout), condition, done)
}

// poller returns a WaitFunc. The wait function sends signal to its returned
// channel every 'interval', and close the channel after 'timeout'.
func poller(interval, timeout time.Duration) WaitFunc {
	return WaitFunc(func(done <-chan struct{}) <-chan struct{} {
		ch := make(chan struct{})

		go func() {
			defer close(ch)

			t := time.NewTicker(interval)
			var after *time.Timer
			if timeout != 0 {
				after = time.NewTimer(timeout)
				defer after.Stop()
			}

			for {
				select {
				case <-done:
					return
				case <-after.C:
					return
				case <-t.C:
					ch <- struct{}{}
				}
			}
		}()
		return ch
	})
}

// Waitfor waits for a condition (as in parameter 'fn') to become true. Instead
// of polling forever, it gets a channel from wait(), and then invokes 'fn' once
// for every value placed on the channel and once more when the channel is closed.
func waitfor(wait WaitFunc, fn ConditionFunc, done <-chan struct{}) error {
	c := wait(done)
	for {
		// Block until we have a signal from wait, or channel is closed.
		_, open := <-c
		ok, err := fn()
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
		if !open {
			return errors.New("Timeout waiting for condition")
		}
	}
}

// waitforNoError waits for a condition to become no error. Instead
// of polling forever, it gets a channel from wait(), and then invokes 'fn' once
// for every value placed on the channel and once more when the channel is closed.
func waitforNoError(wait WaitFunc, fn NoErrorConditionFunc, done <-chan struct{}) error {
	c := wait(done)
	for {
		// Block until we have a signal from wait, or channel is closed.
		_, open := <-c
		err := fn()
		if err == nil {
			return nil
		}
		if !open {
			return errors.New("Timeout waiting for condition")
		}
	}
}
