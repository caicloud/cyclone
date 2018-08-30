package handlers

import (
	"context"

	"github.com/caicloud/cargo-admin/pkg/api/token/types"
	"github.com/caicloud/cargo-admin/pkg/resource"

	"github.com/caicloud/nirvana/log"
)

func HealthCheck(ctx context.Context, registry string) (*types.HealthCheckResult, error) {
	ret, err := resource.HealthCheck(ctx, registry)
	if err != nil {
		log.Errorf("health check error: %v", err)
		return nil, err
	}
	return ret, nil
}
