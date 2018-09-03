package handlers

import (
	"context"

	"github.com/caicloud/cargo-admin/pkg/api/admin/types"
	"github.com/caicloud/cargo-admin/pkg/env"
	. "github.com/caicloud/cargo-admin/pkg/errors"
	"github.com/caicloud/cargo-admin/pkg/resource"

	"github.com/caicloud/nirvana/log"
)

func TriggerImageCopy(ctx context.Context, tid string, username string, ticReq *types.TriggerImageCopcyReq) error {
	log.Infof("source image: %s", ticReq.Source)
	log.Infof("target image: %s", ticReq.Target)

	source, err := resource.ParseImageInfo(ticReq.Source)
	if err != nil {
		return err
	}
	target, err := resource.ParseImageInfo(ticReq.Target)
	if err != nil {
		return err
	}

	if (!env.IsSystemTenant(tid)) && target.Project.IsPublic {
		return ErrorSystemTenantAllowed.Error("copy image into public project")
	}

	log.Infof("start copy image from %s to %s", ticReq.Source, ticReq.Target)
	err = resource.TriggerImageCopy(ctx, tid, username, &resource.ImageCopyReq{Source: source, Target: target})
	if err != nil {
		log.Errorf("copy image error: %v ", err)
		return err
	}
	log.Infof("finish copy image from %s to %s", ticReq.Source, ticReq.Target)
	return nil
}
