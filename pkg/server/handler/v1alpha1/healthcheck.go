package v1alpha1

import (
	"context"

	"github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
)

// HealthCheck checks the health status of Cyclone server.
func HealthCheck(ctx context.Context) (v1alpha1.HealthStatus, error) {
	return v1alpha1.HealthStatus{
		Status: "healthy",
	}, nil
}
