package handlers

import (
	"context"

	"github.com/caicloud/cargo-admin/pkg/api/admin/form"
	"github.com/caicloud/cargo-admin/pkg/api/admin/types"
	"github.com/caicloud/cargo-admin/pkg/resource"
)

func ListRepositories(ctx context.Context, seqID, tenant, registry, project, query, sort string, p *form.Pagination) (*types.ListResponse, map[string]string, error) {
	headers := make(map[string]string)
	headers[types.HeaderSeqID] = seqID
	total, list, err := resource.ListRepositories(ctx, tenant, registry, project, query, sort, p.Start, p.Limit)
	if err != nil {
		return nil, headers, err
	}

	return types.NewListResponse(total, list), headers, nil
}

func GetRepository(ctx context.Context, tid, registry, pname, repoName string) (*types.Repository, error) {
	return resource.GetRepository(ctx, tid, registry, pname, repoName)
}

func UpdateRepository(ctx context.Context, tid string, registry string, pname string, repoName string, urr *types.UpdateRepositoryReq) (*types.Repository, error) {
	return resource.UpdateRepository(ctx, tid, registry, pname, repoName, urr.Spec.Description)
}

func DeleteRepository(ctx context.Context, tid, registry, pname, repoName string) error {
	return resource.DeleteRepository(ctx, tid, registry, pname, repoName)
}
