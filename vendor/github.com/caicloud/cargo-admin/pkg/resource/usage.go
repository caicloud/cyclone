package resource

import (
	"context"

	"github.com/caicloud/cargo-admin/pkg/api/admin/types"
	"github.com/caicloud/cargo-admin/pkg/harbor"
	"github.com/caicloud/cargo-admin/pkg/models"

	"github.com/caicloud/nirvana/log"
)

// TODO: 这个函数可能会比较慢，需要改进
func ListRegistryUsages(ctx context.Context, registry string) ([]*types.TenantStatistic, error) {
	cli, err := harbor.ClientMgr.GetProjectClient(registry)
	if err != nil {
		return nil, err
	}

	pGroups, err := models.Project.GetGroupedProjects(registry)
	if err != nil {
		log.Errorf("get projects for registry %s error: %v", registry, err)
		return nil, err
	}

	statistics := make([]*types.TenantStatistic, 0)
	for _, group := range pGroups {
		statistic := types.TenantStatistic{
			Tenant:       group.Tenant,
			ProjectCount: &types.ProjectCount{},
			RepoCount:    &types.RepositoryCount{},
		}

		if len(group.PIDs) != len(group.Publics) {
			log.Errorf("observe broken group data for tenant: %s", group.Tenant)
			continue
		}

		for i, pid := range group.PIDs {
			repoCount, err := cli.GetRepoCount(pid)
			if err != nil {
				log.Warningf("get repo count for project %d error: %v", pid, err)
			}
			if group.Publics[i] {
				statistic.ProjectCount.Public++
				statistic.RepoCount.Public += int64(repoCount)
			} else {
				statistic.ProjectCount.Private++
				statistic.RepoCount.Private += int64(repoCount)
			}
		}
		statistics = append(statistics, &statistic)
	}

	return statistics, nil
}
