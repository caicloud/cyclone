package resource

import (
	"context"
	"fmt"
	"strings"

	"github.com/caicloud/cargo-admin/pkg/api/admin/types"
	. "github.com/caicloud/cargo-admin/pkg/errors"
	"github.com/caicloud/cargo-admin/pkg/harbor"
	"github.com/caicloud/cargo-admin/pkg/models"

	"github.com/caicloud/nirvana/log"

	"gopkg.in/mgo.v2"
)

func ListTags(ctx context.Context, tenant, registry, project, repo, query string, start, limit int) (int, []*types.Tag, error) {
	rInfo, err := models.Registry.FindByName(registry)
	if err != nil {
		return 0, nil, err
	}
	cli, err := harbor.ClientMgr.GetClient(registry)
	if err != nil {
		return 0, nil, err
	}

	tags, err := cli.ListTags(project, repo)
	if err != nil {
		log.Errorf("list tags from harbor error: %v, projectName: %s, repoName: %s", err, project, repo)
		return 0, nil, err
	}

	if query != "" {
		tags = filterTags(tags, query)
	}
	total := len(tags)

	ret := make([]*types.Tag, 0, len(tags))
	tags = PageHarborTags(ctx, tags, start, limit)
	for _, tag := range tags {
		vulnerabilities, err := cli.GetTagVulnerabilities(project, repo, tag.Name)
		if err != nil {
			log.Errorf("get vulnerability from harbor error: %v", err)
			return 0, nil, err
		}
		ret = append(ret, &types.Tag{
			Metadata: &types.TagMetadata{
				Name:         tag.Name,
				CreationTime: tag.Created,
			},
			Spec: &types.TagSpec{
				Image: fmt.Sprintf("%s/%s/%s:%s", rInfo.Domain, project, repo, tag.Name),
			},
			Status: &types.TagStatus{
				Author:               tag.Author,
				VulnerabilitiesCount: len(vulnerabilities),
				Vulnerabilities:      convertVulnerabilities(vulnerabilities),
			},
		})
	}
	return total, ret, nil
}

func filterTags(tags []*harbor.HarborTag, q string) []*harbor.HarborTag {
	if q == "" {
		return tags
	}

	ret := make([]*harbor.HarborTag, 0)
	for _, t := range tags {
		if strings.Index(t.Name, q) == -1 {
			continue
		}
		ret = append(ret, t)
	}
	return ret
}

func convertVulnerabilities(hvs []*harbor.HarborVulnerability) []*types.Vulnerability {
	ret := make([]*types.Vulnerability, 0, len(hvs))
	for _, hv := range hvs {
		ret = append(ret, &types.Vulnerability{
			Name:        hv.ID,
			Package:     hv.Pkg,
			Description: hv.Description,
			Link:        hv.Link,
			Severity:    hv.Severity.String(),
			Version:     hv.Version,
			Fixed:       hv.Fixed,
		})
	}
	return ret
}

func GetTag(ctx context.Context, tid, registry, pname, repoName, tag string) (*types.Tag, error) {
	rinfo, err := models.Registry.FindByName(registry)
	if err != nil {
		return nil, err
	}
	cli, err := harbor.ClientMgr.GetClient(registry)
	if err != nil {
		return nil, err
	}
	ret, err := cli.GetTag(pname, repoName, tag)
	if err != nil {
		log.Errorf("get tag: %s from harbor error: %v, projectName: %s, repoName: %s", tag, err, pname, repoName)
		return nil, err
	}

	vulnerabilities, err := cli.GetTagVulnerabilities(pname, repoName, tag)
	if err != nil {
		log.Errorf("get vulnerability from harbor error: %v", err)
		return nil, err
	}

	return &types.Tag{
		Metadata: &types.TagMetadata{
			Name:         ret.Name,
			CreationTime: ret.Created,
		},
		Spec: &types.TagSpec{
			Image: fmt.Sprintf("%s/%s/%s:%s", rinfo.Domain, pname, repoName, ret.Name),
		},
		Status: &types.TagStatus{
			Author:               ret.Author,
			VulnerabilitiesCount: len(vulnerabilities),
			Vulnerabilities:      convertVulnerabilities(vulnerabilities),
		},
	}, nil
}

func DeleteTag(ctx context.Context, tenant, registry, pname, repoName, tag string) error {
	wsinfo, err := models.Project.FindByNameWithoutTenant(registry, pname)
	if err != nil {
		if err == mgo.ErrNotFound {
			return ErrorContentNotFound.Error(fmt.Sprintf("project: %v", pname))
		}
		return ErrorUnknownInternal.Error(err)
	}
	if wsinfo.IsPublic && wsinfo.Tenant != tenant {
		return ErrorDeleteFailed.Error(pname, fmt.Sprintf("project %v is public project, the images in public project can only be deleted by system-admin tenant", pname))
	}

	cli, err := harbor.ClientMgr.GetClient(registry)
	if err != nil {
		return err
	}
	err = cli.DeleteTag(pname, repoName, tag)
	if err != nil {
		log.Errorf("delete tag: %s from harbor error: %v, projectName: %s, repoName: %s", tag, err, pname, repoName)
		return err
	}
	return nil
}
