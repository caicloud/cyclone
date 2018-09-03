package resource

import (
	"context"
	"encoding/json"

	"github.com/caicloud/cargo-admin/pkg/api/admin/types"
	"github.com/caicloud/cargo-admin/pkg/harbor"
	"github.com/caicloud/cargo-admin/pkg/models"

	"github.com/caicloud/nirvana/log"
)

const (
	slashSep = "/"
	colonSep = ":"
)

func projectsArrToMap(projects []*harbor.HarborProject) map[int64]*harbor.HarborProject {
	ret := make(map[int64]*harbor.HarborProject)
	for i, _ := range projects {
		ret[projects[i].ProjectID] = projects[i]
	}
	return ret
}

func getRepoAccess(wsinfo models.ProjectInfo, tenant string) string {
	if wsinfo.Tenant == tenant {
		return ReadWriteAccess
	}
	return ReadAccess
}

func unmarshalVulnerabilities(vs string) ([]*types.Vulnerability, error) {
	ret := make([]*types.Vulnerability, 0)
	if vs == "null" {
		return ret, nil
	}
	err := json.Unmarshal([]byte(vs), &ret)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return ret, nil
}

type filter struct {
	ProjectPrefix string
	HashSlash     bool
	RepoPefix     string
	HasColon      bool
	TagPrefix     string
}

// StartCursor 和 EndCursor 分别为 PageResult[StartCursor:EndCursor]
type page struct {
	StartCursor int
	EndCursor   int
	Page        int
	PageSize    int
}

// start/limit => page/pageSize
// 正常情况下，start/limit => page/pageSize 完全没问题
//  0      5   =>   1     5
//  5      5   =>   2     5
// 10      5   =>   3     5
// 非正常情况，start/limit => page/pageSize 会有问题，根本原因是 start/limit 本质上不是严格的分页
//  1     10   =>   1    10
//  5      8   =>   1     8

// 此函数，不仅可以适配正常情况下的转换，也可以实现非正常情况下的转换，以实现严格的分页
// @param rstart: start
// @param rlimit: limit
// @param step：可以理解为 pageSize,
// @param length: 需要进行分页的 items 的 length
func optimizePageParams(rstart, rlimit, step, length int) []*page {
	// 此处表示：非正常情况下，items 的 length 小于 2 倍 step，那么将 items 全部 list 出来，再去截取需要的部分即可
	if length < step*2 {
		if rstart+rlimit > length {
			rlimit = length - rstart
		}
		return []*page{
			&page{
				StartCursor: rstart,
				EndCursor:   rstart + rlimit,
				Page:        1,
				PageSize:    length,
			},
		}
	}

	// 此处表示：在正常情况下，直接将 start/limit 转换为 page/pageSize 即可
	if rstart%step == 0 && rlimit == step {
		return []*page{
			&page{
				StartCursor: 0,
				EndCursor:   rlimit,
				Page:        rstart/step + 1,
				PageSize:    step,
			},
		}
	}
	// 此处表示：非正常情况下，直接截取 第一页 的结果即可
	pztotal := length/step + 1
	if (rstart + rlimit) < step {
		return []*page{
			&page{
				StartCursor: rstart % step,
				EndCursor:   rstart%step + rlimit,
				Page:        rstart/step + 1,
				PageSize:    step,
			},
		}
	}

	// 此处表示：非正常情况下，直接截取 最后一页 的结果即可
	if (rstart/step + 1) == pztotal {
		return []*page{
			&page{
				StartCursor: 0,
				EndCursor:   length - rstart,
				Page:        rstart/step + 1,
				PageSize:    step,
			},
		}
	}

	// 此处表示：非正常情况下，需要两次 page 操作，然后分别从两次 page 截取所需的结果，拼凑最最终结果
	lastEndCursor := 0
	if (length - ((rstart/step + 1) * step)) < (rstart%step + 1) {
		lastEndCursor = length - ((rstart/step + 1) * step)
	} else {
		lastEndCursor = rstart%step + 1
	}
	return []*page{
		&page{
			StartCursor: rstart % step,
			EndCursor:   step,
			Page:        rstart/step + 1,
			PageSize:    step,
		},
		&page{
			StartCursor: 0,
			EndCursor:   lastEndCursor,
			Page:        rstart/step + 2,
			PageSize:    step,
		},
	}
}

// 此函数只适用于 start/limit => page/pageSize 在正常情况下的转换
func convertPageParams(start, limit int) (page int, pageSize int) {
	pageSize = limit
	page = start/pageSize + 1
	return page, pageSize
}

func reconvertPageParams(page, pageSize int) (start, limit int) {
	start = (page - 1) * pageSize
	limit = pageSize
	return start, limit
}

func PageRepos(ctx context.Context, inputs []*types.Repository, start, limit int) []*types.Repository {
	if start < 0 || start >= len(inputs) || limit <= 0 {
		return inputs
	}
	end := start + limit
	if end > len(inputs) {
		end = len(inputs)
	}
	return inputs[start:end]
}

func PageHarborTags(ctx context.Context, inputs []*harbor.HarborTag, start, limit int) []*harbor.HarborTag {
	if start < 0 || start >= len(inputs) || limit <= 0 {
		return inputs
	}
	end := start + limit
	if end > len(inputs) {
		end = len(inputs)
	}
	return inputs[start:end]
}
