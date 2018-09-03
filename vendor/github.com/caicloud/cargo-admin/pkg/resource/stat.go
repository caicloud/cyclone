package resource

import (
	"context"
	"time"

	"github.com/caicloud/cargo-admin/pkg/api/admin/types"
	"github.com/caicloud/cargo-admin/pkg/env"
	"github.com/caicloud/cargo-admin/pkg/harbor"
	"github.com/caicloud/cargo-admin/pkg/models"

	"github.com/caicloud/nirvana/log"
)

func ListRegistryStats(ctx context.Context, tid, registry, action string, startTime, endTime int64) (int, []*types.StatItem, error) {
	if tid == env.SystemTenant {
		log.Infof("tenant: %s", tid)
	}

	cli, err := harbor.ClientMgr.GetLogClient(registry)
	if err != nil {
		return 0, nil, err
	}
	hlogs, err := cli.ListLogs(startTime, endTime, action)
	if err != nil {
		return 0, nil, err
	}

	ret := make([]*types.StatItem, 0)
	start := time.Unix(startTime, 0)
	end := time.Unix(endTime, 0)
	day := end.Truncate(time.Hour * 24)
	c := 0

	for {
		item := &types.StatItem{
			TimeStamp: day.Unix(),
			Count:     0,
		}
		for _, hlog := range hlogs[c:] {
			if hlog.OpTime.Unix() >= day.Unix() {
				item.Count++
				c++
			} else {
				break
			}
		}
		ret = append(ret, item)
		day = day.Add(-time.Hour * 24)
		if day.Unix() < start.Unix() {
			break
		}
	}

	return len(ret), ret, nil
}

func ListProjectStats(ctx context.Context, tid, registry, pname, action string, startTime, endTime int64) (int, []*types.StatItem, error) {
	if tid == env.SystemTenant {
	}

	log.Infof("get project stats: tenant=%s, project=%s, action=%s", tid, pname, action)
	cli, err := harbor.ClientMgr.GetLogClient(registry)
	if err != nil {
		log.Errorf("get client for registry %s error: %v", registry, err)
		return 0, nil, err
	}
	pinfo, err := models.Project.FindByNameWithoutTenant(registry, pname)
	if err != nil {
		log.Errorf("find project %s error: %v", pname, err)
		return 0, nil, err
	}

	hlogs, err := cli.ListProjectLogs(pinfo.ProjectId, startTime, endTime, action)
	if err != nil {
		log.Errorf("list project logs error: %v", err)
		return 0, nil, err
	}

	ret := make([]*types.StatItem, 0)
	start := time.Unix(startTime, 0)
	end := time.Unix(endTime, 0)
	day := end.Truncate(time.Hour * 24)
	c := 0

	for {
		item := &types.StatItem{
			TimeStamp: day.Unix(),
			Count:     0,
		}
		for _, hlog := range hlogs[c:] {
			if hlog.OpTime.Unix() >= day.Unix() {
				item.Count++
				c++
			} else {
				break
			}
		}
		ret = append(ret, item)
		day = day.Add(-time.Hour * 24)
		if day.Unix() < start.Unix() {
			break
		}
	}

	return len(ret), ret, nil
}
