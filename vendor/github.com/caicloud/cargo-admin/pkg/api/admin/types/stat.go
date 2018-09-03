package types

type StatItem struct {
	TimeStamp int64 `json:"timeStamp"`
	Count     int   `json:"count"`
}

// TODO: 这个地方要改一下
type TenantStatistic struct {
	Tenant       string           `json:tenant`
	ProjectCount *ProjectCount    `json:projectCount`
	RepoCount    *RepositoryCount `json:repositoryCount`
}

type SortableStaticstic []*TenantStatistic

func (s SortableStaticstic) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s SortableStaticstic) Less(i, j int) bool {
	return s[i].Tenant < s[j].Tenant
}

func (s SortableStaticstic) Len() int {
	return len(s)
}
