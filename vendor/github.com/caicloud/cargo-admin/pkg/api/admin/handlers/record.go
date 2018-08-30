package handlers

import (
	"context"

	"github.com/caicloud/cargo-admin/pkg/api/admin/form"
	"github.com/caicloud/cargo-admin/pkg/api/admin/types"
	"github.com/caicloud/cargo-admin/pkg/resource"

	"github.com/caicloud/nirvana/log"
)

func ListRecords(ctx context.Context, tid string, replication string, status string, triggerType string, p *form.Pagination) (*types.ListResponse, error) {
	total, list, err := resource.ListRecords(ctx, tid, &resource.ListRecordsParams{
		Replication: replication,
		Status:      status,
		TriggerType: triggerType,
		Start:       p.Start,
		Limit:       p.Limit,
	})
	if err != nil {
		log.Errorf("list record error: %v", err)
		return nil, err
	}
	return types.NewListResponse(total, list), nil
}

func GetRecord(ctx context.Context, tid, replication, record string) (*types.Record, error) {
	return resource.GetRecord(ctx, tid, replication, record)
}

func ListRecordImages(ctx context.Context, tid string, replication string, record string, status string, p *form.Pagination) (*types.ListResponse, error) {
	total, list, err := resource.ListRecordImages(ctx, tid, &resource.ListRecordImagesParams{
		Replication: replication,
		Record:      record,
		Status:      status,
		Start:       p.Start,
		Limit:       p.Limit,
	})
	if err != nil {
		log.Errorf("list record images error: %v", err)
		return nil, err
	}

	return types.NewListResponse(total, list), nil
}
