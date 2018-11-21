/*
Copyright 2017 Caicloud Authors

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

package converters

import (
	"context"

	"github.com/caicloud/nirvana/examples/api-basic/application"
	"github.com/caicloud/nirvana/operators/converter"
)

func ConvertApplicationV1ToApplication() converter.Converter {
	return converter.For(func(ctx context.Context, field string, app *application.ApplicationV1) (*application.Application, error) {
		return &application.Application{
			Metadata: application.Metadata{
				Name:      app.Name,
				Partition: app.Partition,
			},
			Spec: application.ApplicationSpec{
				Replica:     app.Replica,
				OtherFields: "Some Default Value",
			},
			Status: application.ApplicationStatus{
				Phase:   app.Phase,
				Message: app.Message,
			},
		}, nil
	})
}

func ConvertApplicationToApplicationV1() converter.Converter {
	return converter.For(func(ctx context.Context, field string, app *application.Application) (*application.ApplicationV1, error) {
		return &application.ApplicationV1{
			Name:      app.Metadata.Name,
			Partition: app.Metadata.Partition,
			Replica:   app.Spec.Replica,
			// Ignore app.Spec.OtherFields
			Phase:   app.Status.Phase,
			Message: app.Status.Message,
		}, nil
	})
}
