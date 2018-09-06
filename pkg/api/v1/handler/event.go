/*
Copyright 2017 caicloud authors. All rights reserved.

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

package handler

import (
	"context"
	"fmt"

	"github.com/caicloud/cyclone/pkg/api"
	contextutil "github.com/caicloud/cyclone/pkg/util/context"
)

// GetEvent handles the request to get a event.
func GetEvent(ctx context.Context, eventID string) (*api.Event, error) {
	event, err := eventManager.GetEvent(eventID)
	if err != nil {
		return nil, err
	}

	return event, nil
}

// SetEvent handles the request to set the event.
func SetEvent(ctx context.Context, eventID string) (*api.Event, error) {
	event := &api.Event{}
	err := contextutil.GetJsonPayload(ctx, event)
	if err != nil {
		return nil, err
	}
	if eventID != event.ID {
		err := fmt.Errorf("The event IDs in the request path and request body are not same")
		return nil, err
	}

	event, err = eventManager.SetEvent(event)
	if err != nil {
		return nil, err
	}

	return event, nil
}
