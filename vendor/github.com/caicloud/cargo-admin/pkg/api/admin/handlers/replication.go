package handlers

import (
	"context"
	"fmt"

	"github.com/caicloud/cargo-admin/pkg/api/admin/form"
	"github.com/caicloud/cargo-admin/pkg/api/admin/types"
	. "github.com/caicloud/cargo-admin/pkg/errors"
	"github.com/caicloud/cargo-admin/pkg/models"
	"github.com/caicloud/cargo-admin/pkg/resource"
	"github.com/caicloud/cargo-admin/pkg/utils/slugify"

	"github.com/caicloud/nirvana/log"
)

// TODO: 参数检查
func CreateReplication(ctx context.Context, tid string, crReq *types.CreateReplicationReq) (*types.Replication, error) {
	name := slugify.Slugify(crReq.Metadata.Alias, true)
	log.Infof("replication alias: %s convert to replication name: %s", crReq.Metadata.Alias, name)
	rep, err := resource.CreateReplication(ctx, tid, &resource.CreateReplicationReq{
		Alias:             crReq.Metadata.Alias,
		Name:              name,
		Project:           crReq.Spec.Project,
		ReplicateNow:      crReq.Spec.ReplicateNow,
		ReplicateDeletion: crReq.Spec.ReplicateDeletion,
		SourceRegistry:    crReq.Spec.Source.Name,
		TargetRegistry:    crReq.Spec.Target.Name,
		Trigger:           crReq.Spec.Trigger,
	})
	if err != nil {
		log.Infof("create replication from registry: %s error: %v", crReq.Spec.Source.Name, err)
		return nil, err
	}
	return rep, nil
}

func ListReplications(ctx context.Context, seqID, tid string, direction string, registry string, project string, triggerType string, q string, p *form.Pagination) (*types.ListResponse, map[string]string, error) {
	headers := make(map[string]string)
	headers[types.HeaderSeqID] = seqID
	if direction == "" {
		return nil, headers, ErrorUnknownRequest.Error("direction is required for query")
	}
	if registry == "" {
		return nil, headers, ErrorUnknownRequest.Error("registry is required for query")
	}

	total, list, err := resource.ListReplications(ctx, tid, &resource.ListReplicationsParams{
		Direction:   direction,
		Registry:    registry,
		Project:     project,
		TriggerType: triggerType,
		Prefix:      q,
		Start:       p.Start,
		Limit:       p.Limit,
	})
	if err != nil {
		return nil, headers, err
	}
	return types.NewListResponse(total, list), headers, nil
}

func GetReplication(ctx context.Context, tid string, replication string) (*types.Replication, error) {
	return resource.GetReplication(ctx, tid, replication)
}

func UpdateReplication(ctx context.Context, tid string, replication string, urReq *types.UpdateReplicationReq) (*types.Replication, error) {
	b, err := models.Replication.IsExist(tid, replication)
	if err != nil {
		log.Errorf("mongo error: %v", err)
		return nil, ErrorUnknownInternal.Error(err)
	}
	if !b {
		return nil, ErrorContentNotFound.Error(fmt.Sprintf("replication: %s", replication))
	}

	repinfo, err := models.Replication.FindByName(tid, replication)
	if err != nil {
		log.Errorf("mongo error: %v", err)
		return nil, err
	}
	if urReq.Spec.Project != repinfo.Project {
		return nil, ErrorUnknownRequest.Error("project can not be modified")
	}

	if urReq.Spec.Source.Name != repinfo.SourceRegistry {
		return nil, ErrorUnknownRequest.Error("source registry can not be moidfied")
	}

	newRep, err := resource.UpdateReplication(ctx, tid, repinfo, urReq)
	if err != nil {
		log.Errorf("update replication: %s error: %v", urReq.Metadata.Name, err)
		return nil, err
	}
	return newRep, nil
}

func DeleteReplication(ctx context.Context, tid string, replication string) error {
	err := resource.DeleteReplication(ctx, tid, replication)
	if err != nil {
		log.Errorf("delete replication: %s error: %v", replication, err)
		return err
	}
	return nil
}

func TriggerReplication(ctx context.Context, tid string, replication string, ctrRep *types.TriggerReplicationReq) error {
	log.Infof("start trigger replication: %s", replication)
	err := resource.TriggerReplication(ctx, tid, replication, ctrRep)
	if err != nil {
		log.Errorf("trigger replication: %s error: %v", replication, err)
		return err
	}
	log.Infof("finish trigger replication: %s", replication)
	return nil
}
