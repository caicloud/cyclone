package resource

import (
	"context"
	"time"

	"code.cloudfoundry.org/bytefmt"

	"github.com/caicloud/cargo-admin/pkg/api/admin/types"
	. "github.com/caicloud/cargo-admin/pkg/errors"
	"github.com/caicloud/cargo-admin/pkg/harbor"
	"github.com/caicloud/cargo-admin/pkg/models"

	"github.com/caicloud/cargo-admin/pkg/env"
	"github.com/caicloud/nirvana/errors"
	"github.com/caicloud/nirvana/log"
)

type CreateRegistryReq struct {
	Alias    string
	Name     string
	Host     string
	Domain   string
	Username string
	Password string
}

func CreateRegistry(ctx context.Context, tenant string, crReq *CreateRegistryReq) (*types.Registry, error) {
	err := checkRegistry(crReq.Host, crReq.Username, crReq.Password)
	if err != nil {
		return nil, err
	}
	err = createAllTargetsToOneRegistry(crReq.Host, crReq.Username, crReq.Password)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	err = createOneTargetToAllRegistries(crReq.Name, crReq.Host, crReq.Username, crReq.Password)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	now := time.Now()
	err = models.Registry.Save(&models.RegistryInfo{
		Name:           crReq.Name,
		Alias:          crReq.Alias,
		Host:           crReq.Host,
		Domain:         crReq.Domain,
		Username:       crReq.Username,
		Password:       crReq.Password,
		CreationTime:   now,
		LastUpdateTime: now,
	})

	for pName, project := range DefaultPublicProjectsMap {
		err := EnsureDefaultPublicProject(&env.DefaultPublicProject{Name: pName, IfExists: project.IfExists, Harbor: crReq.Name})
		if err != nil {
			log.Warningf("create default public project '%s' for registry '%s' error: %v", pName, crReq.Name, err)
		}
	}

	ret, err := GetRegistry(ctx, crReq.Name, tenant)
	if err != nil {
		log.Error(err)
	}
	return ret, nil
}

// Check whether a registry is valid to be added:
// - Whether user name and password are valid
// - Whether the registry is already added to another platform
func checkRegistry(host, username, password string) error {
	cli, err := harbor.NewClient(host, username, password)
	if err != nil {
		return errors.BadRequest.Build(types.AccessFailed, "access to ${url} as user ${user} failed").Error(host, username)
	}
	htargets, err := cli.ListTargets()
	if err != nil {
		log.Errorf("list targets error: %v", err)
		return ErrorUnknownInternal.Error(err)
	}
	if len(htargets) != 0 {
		return errors.BadRequest.Build(types.UsedAlready, "registry ${url} has been used in other Cargo-Admin already").Error(host)
	}
	return nil
}

func ListRegistries(ctx context.Context, tenant string, start, limit int) (int, []*types.Registry, error) {
	total, registries, err := models.Registry.FindOnePage(start, limit)
	if err != nil {
		return 0, nil, ErrorUnknownInternal.Error(err)
	}

	ret := make([]*types.Registry, 0, len(registries))
	for _, r := range registries {
		reg, err := getRetRegistry(r, tenant)
		if err != nil {
			log.Errorf("getRetRegistry error: %v", err)
			reg = &types.Registry{
				Metadata: &types.RegistryMetadata{
					Name:           r.Name,
					Alias:          r.Alias,
					CreationTime:   r.CreationTime,
					LastUpdateTime: r.LastUpdateTime,
				},
				Spec: &types.RegistrySpec{
					Host:     r.Host,
					Domain:   r.Domain,
					Username: r.Username,
					Password: r.Password,
				},
				Status: &types.RegistryStatus{
					ProjectCount: &types.ProjectCount{
						Public:  -1,
						Private: -1,
					},
					RepositoryCount: &types.RepositoryCount{
						Public:  -1,
						Private: -1,
					},
					StorageStatics: &types.StorageStatics{
						Used:  "\\",
						Total: "\\",
					},
					Healthy: false,
				},
			}
		} else {
			reg.Status.Healthy = true
		}
		ret = append(ret, reg)
	}
	return total, ret, nil
}

func GetRegistry(ctx context.Context, registry, tenant string) (*types.Registry, error) {
	regInfo, err := models.Registry.FindByName(registry)
	if err != nil {
		return nil, ErrorUnknownInternal.Error(err)
	}
	ret, err := getRetRegistry(regInfo, tenant)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// Get project counts and repository count for tenant. It's inefficient to get real repository statistics
// since Cargo-Admin doesn't have repository data. We need to send request to Harbor for each project in
// order to get repository count. But for system tenant (or in single tenant mode), we can rely on Harbor's
// own statistic API.
func getCounts(registry, tenant string) (*types.TenantStatistic, error) {
	counts := types.TenantStatistic{
		ProjectCount: &types.ProjectCount{},
		RepoCount:    &types.RepositoryCount{},
	}
	cli, err := harbor.ClientMgr.GetClient(registry)
	if err != nil {
		return nil, err
	}

	if tenant == env.SystemTenant {
		statistics, err := cli.GetStatistics()
		if err != nil {
			return nil, err
		}
		counts.ProjectCount.Public = statistics.PubPC
		counts.ProjectCount.Private = statistics.PriPC
		counts.RepoCount.Public = statistics.PubRC
		counts.RepoCount.Private = statistics.PriRC
		return &counts, nil
	}

	pInfos, err := models.Project.GetTenantProjects(registry, tenant)
	if err != nil {
		return nil, err
	}

	for _, p := range pInfos {
		repoCount, err := cli.GetRepoCount(p.ProjectId)
		if err != nil {
			log.Warningf("get repo count for project %d error: %v", p.ProjectId, err)
		}
		if p.IsPublic {
			counts.ProjectCount.Public++
			counts.RepoCount.Public += int64(repoCount)
		} else {
			counts.ProjectCount.Private++
			counts.RepoCount.Private += int64(repoCount)
		}
	}
	return &counts, nil
}

func getRetRegistry(r *models.RegistryInfo, tenant string) (*types.Registry, error) {
	cli, err := harbor.ClientMgr.GetClient(r.Name)
	if err != nil {
		return nil, err
	}
	volumes, err := cli.GetVolumes()
	if err != nil {
		return nil, err
	}

	counts, err := getCounts(r.Name, tenant)
	if err != nil {
		return nil, err
	}

	return &types.Registry{
		Metadata: &types.RegistryMetadata{
			Name:           r.Name,
			Alias:          r.Alias,
			CreationTime:   r.CreationTime,
			LastUpdateTime: r.LastUpdateTime,
		},
		Spec: &types.RegistrySpec{
			Host:     r.Host,
			Domain:   r.Domain,
			Username: r.Username,
			Password: r.Password,
		},
		Status: &types.RegistryStatus{
			ProjectCount: &types.ProjectCount{
				Public:  counts.ProjectCount.Public,
				Private: counts.ProjectCount.Private,
			},
			RepositoryCount: &types.RepositoryCount{
				Public:  counts.RepoCount.Public,
				Private: counts.RepoCount.Private,
			},
			StorageStatics: &types.StorageStatics{
				Used:  bytefmt.ByteSize(volumes.Storage.Total - volumes.Storage.Free),
				Total: bytefmt.ByteSize(volumes.Storage.Total),
			},
			Healthy: true,
		},
	}, nil
}

type UpdateRegistryReq struct {
	Alias    string
	Username string
	Password string
}

func UpdateRegistry(ctx context.Context, registry, tenant string, urReq *UpdateRegistryReq) (*types.Registry, error) {
	regInfo, err := models.Registry.FindByName(registry)
	if err != nil {
		return nil, ErrorUnknownInternal.Error(err)
	}

	_, err = harbor.NewClient(regInfo.Host, urReq.Username, urReq.Password)
	if err != nil {
		return nil, errors.BadRequest.Build(types.AccessFailed, "access to ${url} as user ${user} failed").Error(regInfo.Host, urReq.Username)
	}

	err = models.Registry.Update(registry, urReq.Alias, urReq.Username, urReq.Password)
	if err != nil {
		log.Errorf("mongo error: %v", err)
		return nil, ErrorUnknownInternal.Error(err)
	}

	// If user name or password changed, update target information in other registries.
	if (regInfo.Username != urReq.Username) || (regInfo.Password != urReq.Password) {
		err = updateOneTargetToAllRegistries(registry, urReq.Username, urReq.Password)
		if err != nil {
			log.Error(err)
			return nil, err
		}
	}

	return GetRegistry(ctx, registry, tenant)
}

func updateOneTargetToAllRegistries(registry, username, password string) error {
	registries, err := models.Registry.FindAll()
	if err != nil {
		return err
	}
	for _, r := range registries {
		if r.Name == registry {
			continue
		}
		cli, err := harbor.ClientMgr.GetClient(r.Name)
		if err != nil {
			return err
		}

		targets, err := cli.ListTargets()
		if err != nil {
			return err
		}

		for _, t := range targets {
			if t.Name == registry {
				err = cli.UpdateTarget(t.ID, t.Name, t.URL, username, password)
				if err != nil {
					log.Errorf("update target '%s' in registry '%s', error %v", t.Name, r.Name, err)
					return ErrorUnknownInternal.Error(err)
				}
				log.Errorf("update target '%s' in registry '%s' successfully", t.Name, r.Name)
			}
		}
	}
	return nil
}

func DeleteRegistry(ctx context.Context, registry string) error {
	err := DeleteAllReplicationsBySourceRegistry(ctx, registry)
	if err != nil {
		log.Errorf("delete all replications from registry: %s error: %v", registry, err)
		return err
	}

	err = DeleteAllReplicationsByTargetRegistry(ctx, registry)
	if err != nil {
		log.Errorf("delete all replications to registry: %s error: %v", registry, err)
		return err
	}

	err = deleteAllHarborTargets(ctx, registry)
	if err != nil {
		log.Errorf("delete all harbor targets from registry: %s error: %v", registry, err)
		return err
	}

	err = deleteOneTargetFromAllRegistries(registry)
	if err != nil {
		log.Errorf("deleteOneTargetFromAllRegistries error: %v", err)
		return ErrorUnknownInternal.Error(err)
	}

	err = models.Project.DeleteAllByRegistry(registry)
	if err != nil {
		log.Errorf("delete all projects from mongo error: %v", err)
		return ErrorUnknownInternal.Error(err)
	}

	err = models.DockerAccount.DeleteAll(registry)
	if err != nil {
		log.Errorf("delete all docker account from mongo error: %v", err)
		return ErrorUnknownInternal.Error(err)
	}

	err = models.Registry.Delete(registry)
	if err != nil {
		return ErrorUnknownInternal.Error(err)
	}
	return nil
}

func IsExistRegistryHost(ctx context.Context, host string) (bool, error) {
	b, err := models.Registry.IsExistHost(host)
	if err != nil {
		return false, ErrorUnknownInternal.Error(err)
	}
	return b, err
}
