package resource

import (
	"context"

	"github.com/caicloud/cargo-admin/pkg/api/token/types"
	. "github.com/caicloud/cargo-admin/pkg/errors"
	"github.com/caicloud/cargo-admin/pkg/harbor"
	"github.com/caicloud/cargo-admin/pkg/models"
	"github.com/caicloud/nirvana/log"
)

func HealthCheck(ctx context.Context, registry string) (*types.HealthCheckResult, error) {
	var regInfo *models.RegistryInfo
	var err error
	log.Infof("Health Check for: %s", registry)
	if registry == "" {
		regInfo, err = models.Registry.FindByName(DefaultRegName)
	} else {
		regInfo, err = models.Registry.FindByDomain(registry)
	}
	if err != nil {
		return nil, ErrorUnknownInternal.Error(err)
	}

	cli, err := harbor.ClientMgr.GetClient(regInfo.Name)
	if err != nil {
		return nil, ErrorUnknownInternal.Error(err)
	}

	_, _, err = cli.ListProjects(1, 1, "", "")
	if err != nil {
		return nil, ErrorUnknownInternal.Error(err)
	}

	return &types.HealthCheckResult{
		Mongo: "healthy",
		Cargo: "healthy",
	}, nil
}
