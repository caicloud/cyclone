package handlers

import (
	"context"
	"fmt"

	"github.com/caicloud/cargo-admin/pkg/api/admin/form"
	"github.com/caicloud/cargo-admin/pkg/api/admin/types"
	. "github.com/caicloud/cargo-admin/pkg/errors"
	"github.com/caicloud/cargo-admin/pkg/models"
	"github.com/caicloud/cargo-admin/pkg/resource"

	"github.com/caicloud/nirvana/log"
)

func CreateProject(ctx context.Context, tid string, registry string, cpReq *types.CreateProjectReq) (*types.Project, error) {
	pname := ""
	if resource.IsMultiTenantEnabled() {
		pname = fmt.Sprintf("%s_%s", tid, cpReq.Metadata.Name)
	} else {
		pname = fmt.Sprintf("%s", cpReq.Metadata.Name)
	}

	exist, err := models.Project.IsExist(tid, registry, pname)
	if err != nil {
		return nil, ErrorUnknownInternal.Error(err)
	}
	if exist {
		return nil, ErrorProjectAlreadyExist.Error(pname)
	}

	desc := ""
	if cpReq.Spec != nil {
		desc = cpReq.Spec.Description
	}

	log.Infof("start create project: %s into registry: %s", pname, registry)
	project, err := resource.CreateProject(ctx, tid, registry, pname, desc, false)
	if err != nil {
		log.Errorf("create project: %s into registry: %s error: %v", pname, registry, err)
		return nil, err
	}
	log.Infof("finish create project: %s into registry: %s", pname, registry)

	return project, nil
}

func ListProjects(ctx context.Context, seqID, tid string, registry string, includePublic bool, q string, p *form.Pagination) (*types.ListResponse, map[string]string, error) {
	headers := make(map[string]string)
	headers[types.HeaderSeqID] = seqID
	total, list, err := resource.ListProjects(ctx, tid, registry, includePublic, q, p.Start, p.Limit)
	if err != nil {
		log.Errorf("list projects from registry: %s error: %v", registry, err)
		return nil, headers, err
	}
	return types.NewListResponse(total, list), headers, nil
}

func GetProject(ctx context.Context, tid, registry, pname string) (*types.Project, error) {
	project, err := resource.GetProject(ctx, tid, registry, pname)
	if err != nil {
		log.Errorf("get project: %s from registry: %s error: %v", pname, registry, err)
		return nil, err
	}
	return project, nil
}

func UpdateProject(ctx context.Context, tid string, registry string, pname string, uwReq *types.UpdateProjectReq) (*types.Project, error) {
	desc := ""
	if uwReq.Spec != nil {
		desc = uwReq.Spec.Description
	}
	return resource.UpdateProject(ctx, tid, registry, pname, desc)
}

func DeleteProject(ctx context.Context, tid, registry, pname string) error {
	log.Infof("start delete project: %s from registry: %s", pname, registry)
	return resource.DeleteProject(ctx, tid, registry, pname)
}

func ListProjectStats(ctx context.Context, tid, registry, pname, action string, startTime, endTime int64) (*types.ListResponse, error) {
	total, list, err := resource.ListProjectStats(ctx, tid, registry, pname, action, startTime, endTime)
	if err != nil {
		return nil, err
	}
	return types.NewListResponse(total, list), nil
}
