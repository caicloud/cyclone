package handlers

import (
	"context"
	"fmt"
	"sort"

	"github.com/caicloud/cargo-admin/pkg/api/admin/form"
	"github.com/caicloud/cargo-admin/pkg/api/admin/types"
	"github.com/caicloud/cargo-admin/pkg/env"
	. "github.com/caicloud/cargo-admin/pkg/errors"
	"github.com/caicloud/cargo-admin/pkg/resource"
	"github.com/caicloud/cargo-admin/pkg/utils/domain"
	"github.com/caicloud/cargo-admin/pkg/utils/matcher"
	"github.com/caicloud/cargo-admin/pkg/utils/slugify"

	"github.com/caicloud/nirvana/log"

	"github.com/caicloud/nirvana/errors"
	"github.com/vmware/harbor/src/common/utils"
)

func CreateRegistry(ctx context.Context, tid string, crReq *types.CreateRegistryReq) (*types.Registry, error) {
	if !env.IsSystemTenant(tid) {
		return nil, ErrorSystemTenantAllowed.Error("add registry")
	}
	url, err := utils.ParseEndpoint(crReq.Spec.Host)
	if err != nil {
		log.Errorf("parse registry host error: %v", err)
		return nil, errors.BadRequest.Build(types.BadUrl, "parse ${url} error: ${err}").Error(crReq.Spec.Host, err)
	}
	if url.Scheme != "https" {
		return nil, errors.BadRequest.Build(types.BadScheme, "${msg}").Error("only support \"https\" endpoint now")
	}
	host := url.String()
	domain, err := domain.GetDomain(host)
	if err != nil {
		return nil, errors.BadRequest.Build(types.BadUrl, "get domain from ${url} error: ${err}").Error(host, err)
	}
	if matcher.IsIP(domain) {
		return nil, errors.BadRequest.Build(types.NoIPAllowed, "only domain accepted, but got ip ${ip}").Error(domain)
	}
	b, err := resource.IsExistRegistryHost(ctx, host)
	if err != nil {
		return nil, errors.InternalServerError.Build(types.Unknown, "${msg}: ${err}").Error("check host existence error", err)
	}
	if b {
		return nil, errors.Conflict.Build(types.RegistryExisted, "registry ${url} already existed in Cargo-Amin").Error(host)
	}

	reg, err := resource.CreateRegistry(ctx, tid, &resource.CreateRegistryReq{
		Alias:    crReq.Metadata.Alias,
		Name:     slugify.Slugify(crReq.Metadata.Alias, true),
		Host:     host,
		Domain:   domain,
		Username: crReq.Spec.Username,
		Password: crReq.Spec.Password,
	})
	if err != nil {
		log.Infof("create registry error: %v", err)
		return nil, err
	}
	return reg, nil
}

func ListRegistries(ctx context.Context, tid string, p *form.Pagination) (*types.ListResponse, error) {
	total, list, err := resource.ListRegistries(ctx, tid, p.Start, p.Limit)
	if err != nil {
		return nil, err
	}
	return types.NewListResponse(total, list), nil
}

func GetRegistry(ctx context.Context, tid string, registry string) (*types.Registry, error) {
	reg, err := resource.GetRegistry(ctx, registry, tid)
	if err != nil {
		log.Errorf("get registry: %s error: %v", registry, err)
		return nil, err
	}
	return reg, nil
}

func UpdateRegistry(ctx context.Context, tid string, registry string, urReq *types.UpdateRegistryReq) (*types.Registry, error) {
	newReg, err := resource.UpdateRegistry(ctx, registry, tid, &resource.UpdateRegistryReq{
		Alias:    urReq.Metadata.Alias,
		Username: urReq.Spec.Username,
		Password: urReq.Spec.Password,
	})
	if err != nil {
		log.Errorf("update registry: %s error: %v", registry, err)
		return nil, err
	}
	return newReg, nil
}

func DeleteRegistry(ctx context.Context, tid string, registry string) error {
	err := resource.DeleteRegistry(ctx, registry)
	if err != nil {
		log.Errorf("delete registry: %s error: %v", registry, err)
		return err
	}
	return nil
}

func ListRegistryStats(ctx context.Context, tid string, registry, action string, startTime, endTime int64) (*types.ListResponse, error) {
	total, list, err := resource.ListRegistryStats(ctx, tid, registry, action, startTime, endTime)
	if err != nil {
		return nil, err
	}
	return types.NewListResponse(total, list), nil
}

func ListRegistryUsages(ctx context.Context, tid, registry string, p *form.Pagination) (*types.ListResponse, error) {
	list, err := resource.ListRegistryUsages(ctx, registry)
	if err != nil {
		return nil, err
	}

	if len(list) == 0 {
		return types.NewListResponse(0, make([]*types.TenantStatistic, 0)), nil
	}

	if p.Start >= len(list) {
		return nil, ErrorUnknownRequest.Error(fmt.Sprintf("pagination (start: %d, limit: %d) out of bound, total: %d", p.Start, p.Limit, len(list)))
	}

	min := func(a, b int) int {
		if a >= b {
			return b
		}
		return a
	}
	total := len(list)
	sort.Sort(types.SortableStaticstic(list))
	list = list[p.Start:min(p.Start+p.Limit, len(list))]

	return types.NewListResponse(total, list), nil
}
