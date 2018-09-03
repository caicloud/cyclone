package handlers

import (
	"context"

	"github.com/caicloud/cargo-admin/pkg/api/admin/types"
	"github.com/caicloud/cargo-admin/pkg/errors"
	"github.com/caicloud/cargo-admin/pkg/resource"
)

func GetOrNewDockerAccount(ctx context.Context, registry, username string) (*types.DockerAccount, error) {
	if username == "" {
		return nil, errors.ErrorUnknownRequest.Error("user name is empty")
	}

	return resource.GetOrNewDockerAccount(ctx, registry, username)
}
