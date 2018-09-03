package handlers

import (
	"context"
	"github.com/caicloud/cargo-admin/pkg/harbor"
	"github.com/caicloud/cargo-admin/pkg/models"
)

func HealthCheck(ctx context.Context) error {
	_, err := models.Registry.FindAll()
	if err != nil {
		return err
	}

	_, err = harbor.ClientMgr.GetClient("default")
	if err != nil {
		return err
	}

	return nil
}
