package handlers

import (
	"context"

	"github.com/caicloud/cargo-admin/pkg/api/admin/form"
	"github.com/caicloud/cargo-admin/pkg/api/admin/types"
	"github.com/caicloud/cargo-admin/pkg/resource"
)

func ListTags(ctx context.Context, tid string, registry string, pname string, repoName string, q string, p *form.Pagination) (*types.ListResponse, error) {
	total, list, err := resource.ListTags(ctx, tid, registry, pname, repoName, q, p.Start, p.Limit)
	if err != nil {
		return nil, err
	}
	return types.NewListResponse(total, list), nil
}

func GetTag(ctx context.Context, tid, registry, pname, repoName, tag string) (*types.Tag, error) {
	return resource.GetTag(ctx, tid, registry, pname, repoName, tag)
}

func DeleteTag(ctx context.Context, tid, registry, pname, repoName, tag string) error {
	return resource.DeleteTag(ctx, tid, registry, pname, repoName, tag)
}
