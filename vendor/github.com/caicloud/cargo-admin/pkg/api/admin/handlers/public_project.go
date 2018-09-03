package handlers

import (
	"context"

	"github.com/caicloud/cargo-admin/pkg/api/admin/form"
	"github.com/caicloud/cargo-admin/pkg/api/admin/types"
	"github.com/caicloud/cargo-admin/pkg/env"
	. "github.com/caicloud/cargo-admin/pkg/errors"
	"github.com/caicloud/cargo-admin/pkg/models"
	"github.com/caicloud/cargo-admin/pkg/resource"

	"github.com/caicloud/nirvana/log"
)

func CreatePublicProject(ctx context.Context, tid string, registry string, cpwReq *types.CreatePublicProjectReq) (*types.PublicProject, error) {
	if !env.IsSystemTenant(tid) {
		return nil, ErrorSystemTenantAllowed.Error("create public project")
	}

	exist, err := models.Project.IsExist(tid, registry, cpwReq.Metadata.Name)
	if err != nil {
		return nil, ErrorUnknownInternal.Error(err)
	}
	if exist {
		return nil, ErrorProjectAlreadyExist.Error(cpwReq.Metadata.Name)
	}

	desc := ""
	if cpwReq.Spec != nil {
		desc = cpwReq.Spec.Description
	}

	log.Infof("start create public project: %s into registry: %s", cpwReq.Metadata.Name, registry)
	project, err := resource.CreatePublicProject(ctx, tid, registry, cpwReq.Metadata.Name, desc)
	if err != nil {
		log.Errorf("create public project: %s into registry: %s error: %v", cpwReq.Metadata.Name, registry, err)
		return nil, err
	}
	log.Infof("finish create public project: %s into registry: %s", cpwReq.Metadata.Name, registry)

	return project, nil
}

func ListPublicProjects(ctx context.Context, seqID, tid string, registry string, p *form.Pagination) (*types.ListResponse, map[string]string, error) {
	headers := make(map[string]string)
	headers[types.HeaderSeqID] = seqID
	total, list, err := resource.ListPublicProjects(ctx, tid, registry, p.Start, p.Limit)
	if err != nil {
		log.Errorf("list public projects from registry: %s error: %v", registry, err)
		return nil, headers, err
	}
	return types.NewListResponse(total, list), headers, nil
}

func UpdatePublicProject(ctx context.Context, tid string, registry string, pname string, uwReq *types.UpdatePublicProjectReq) (*types.Project, error) {
	desc := ""
	if uwReq.Spec != nil {
		desc = uwReq.Spec.Description
	}
	return resource.UpdatePublicProject(ctx, tid, registry, pname, desc)
}

func DeletePublicProject(ctx context.Context, tid, registry string, pname string) error {
	log.Infof("start delete public project: %s from registry: %s", pname, registry)
	return resource.DeletePublicProject(ctx, tid, registry, pname)
}
