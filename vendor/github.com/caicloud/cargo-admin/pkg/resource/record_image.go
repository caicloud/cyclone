package resource

import (
	"context"
	"fmt"

	"github.com/caicloud/cargo-admin/pkg/api/admin/types"
	. "github.com/caicloud/cargo-admin/pkg/errors"
	"github.com/caicloud/cargo-admin/pkg/models"

	"github.com/caicloud/nirvana/log"
)

type ListRecordImagesParams struct {
	Replication string
	Record      string
	Status      string
	Start       int
	Limit       int
}

// =================================================================================================
func ListRecordImages(ctx context.Context, tenant string, param *ListRecordImagesParams) (int, []*types.RecordImage, error) {
	log.Infof("list replication: %s records: %s 's images, status: %s,  start: %d, limit: %d",
		param.Replication, param.Record, param.Status, param.Start, param.Limit)

	repinfo, err := models.Replication.FindByName(tenant, param.Replication)
	if err != nil {
		return 0, nil, ErrorUnknownInternal.Error(err)
	}

	err = updateSyncingRecordStatus(repinfo)
	if err != nil {
		log.Errorf("updateSyncingRecordStatus error: %v", err)
		return 0, nil, err
	}
	total, recimginfos, err := models.RecordImage.FindOnePage(param.Record, revertStatus(param.Status), param.Start, param.Limit)
	if err != nil {
		return 0, nil, ErrorUnknownInternal.Error(err)
	}

	ret := make([]*types.RecordImage, 0, len(recimginfos))
	for _, recimginfo := range recimginfos {
		operation := "transfer"
		if recimginfo.Operation != "" {
			operation = recimginfo.Operation
		}
		ret = append(ret, &types.RecordImage{
			Image:     fmt.Sprintf("%s:%s", recimginfo.Repository, recimginfo.Tag),
			Operation: operation,
			Status:    recimginfo.Status,
		})
	}

	return total, ret, nil
}
