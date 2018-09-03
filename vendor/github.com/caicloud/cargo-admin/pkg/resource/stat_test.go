package resource

import (
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/caicloud/cargo-admin/pkg/api/admin/types"
	"github.com/caicloud/cargo-admin/pkg/harbor"
	"github.com/caicloud/cargo-admin/pkg/mock"
	"github.com/caicloud/cargo-admin/pkg/models"

	"fmt"
	"github.com/davecgh/go-spew/spew"
)

func initClientData() {
	mocked := harbor.ClientMgr.(*mock.MockedClients)
	mocked.Add("r1", mock.NewProjectClient(), mock.ProClient)
	mocked.Add("r2", mock.NewProjectClient(), mock.ProClient)
	mocked.Add("r1", mock.NewLogClient(), mock.LgClient)
	mocked.Add("r2", mock.NewLogClient(), mock.LgClient)
}

func initProjectData() {
	projects := []*models.ProjectInfo{
		{Name: "p1", Registry: "r2", Tenant: "t1", ProjectId: 1, IsPublic: true},
		{Name: "p2", Registry: "r2", Tenant: "t2", ProjectId: 2, IsPublic: true},
		{Name: "p3", Registry: "r2", Tenant: "t2", ProjectId: 3, IsPublic: false},
		{Name: "p4", Registry: "r1", Tenant: "t3", ProjectId: 4, IsPublic: true},
		{Name: "p5", Registry: "r1", Tenant: "t3", ProjectId: 5, IsPublic: false},
		{Name: "p6", Registry: "r1", Tenant: "t3", ProjectId: 6, IsPublic: false},
	}
	for _, p := range projects {
		models.Project.Save(p)
	}

	client, _ := harbor.ClientMgr.GetProjectClient("r1")
	mocked, _ := client.(*mock.ProjectClient)
	mocked.Add("p1", true, 1)
	mocked.Add("p2", true, 2)
	mocked.Add("p3", false, 3)
	mocked.Add("p4", true, 4)
	mocked.Add("p5", false, 5)
	mocked.Add("p6", false, 6)
}

func initLogData() {
	client, _ := harbor.ClientMgr.GetLogClient("r2")
	mocked, _ := client.(*mock.LogClient)
	startTime := time.Date(2018, time.January, 1, 12, 0, 0, 0, time.UTC)
	logs := []*harbor.HarborAccessLog{
		{LogID: 1, RepoName: "repo1", Operation: "pull", OpTime: startTime.AddDate(0, 0, 2)},
		{LogID: 2, RepoName: "repo1", Operation: "pull", OpTime: startTime.AddDate(0, 0, 2)},
		{LogID: 2, RepoName: "repo1", Operation: "pull", OpTime: startTime.AddDate(0, 0, 2)},
		{LogID: 3, RepoName: "repo1", Operation: "pull", OpTime: startTime.AddDate(0, 0, 1)},
		{LogID: 4, RepoName: "repo1", Operation: "pull", OpTime: startTime.AddDate(0, 0, 1)},
		{LogID: 5, RepoName: "repo1", Operation: "pull", OpTime: startTime},
	}
	for _, log := range logs {
		mocked.Add(log)
	}
}

func TestListRegistryUsages(t *testing.T) {
	originClients := harbor.ClientMgr
	harbor.ClientMgr = mock.NewMockedClients()
	originProject := models.Project
	models.Project = mock.NewProjectModel()
	initClientData()
	initProjectData()
	defer func() {
		harbor.ClientMgr = originClients
		models.Project = originProject
	}()

	statistics, err := ListRegistryUsages(nil, "r1")
	if err != nil {
		t.Errorf("Unexpected err: %v", err)
		return
	}

	sort.Sort(types.SortableStaticstic(statistics))
	truth := []*types.TenantStatistic{
		{"t1", &types.ProjectCount{1, 0}, &types.RepositoryCount{1, 0}},
		{"t2", &types.ProjectCount{1, 1}, &types.RepositoryCount{2, 3}},
		{"t3", &types.ProjectCount{1, 2}, &types.RepositoryCount{4, 11}},
	}
	if !reflect.DeepEqual(statistics, truth) {
		t.Errorf("Get registry usage error, expected %s, but got %s", spew.Sdump(truth), spew.Sdump(statistics))
	}
}

func TestListProjectStats(t *testing.T) {
	originClients := harbor.ClientMgr
	harbor.ClientMgr = mock.NewMockedClients()
	originProject := models.Project
	models.Project = mock.NewProjectModel()
	initClientData()
	initProjectData()
	initLogData()
	defer func() {
		harbor.ClientMgr = originClients
		models.Project = originProject
	}()
	days := 3
	startTime := time.Date(2018, time.January, 1, 12, 0, 0, 0, time.UTC).Truncate(time.Hour * 24)
	endTime := startTime.AddDate(0, 0, days).Add(-time.Millisecond)
	cnt, stats, err := ListProjectStats(nil, "t1", "r2", "p1", "pull", startTime.Unix(), endTime.Unix())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if cnt != days {
		t.Errorf("Get project status error, expected %d, but got %d", days, cnt)
		return
	}

	day := startTime.Truncate(time.Hour * 24)
	truth := []*types.StatItem{
		{day.AddDate(0, 0, 2).Unix(), 3},
		{day.AddDate(0, 0, 1).Unix(), 2},
		{day.Unix(), 1},
	}
	if !reflect.DeepEqual(stats, truth) {
		t.Errorf("Get project status error, expected %s, but got %s", spew.Sdump(truth), spew.Sdump(stats))
	}
}

func TestListRegistryStats(t *testing.T) {
	originClients := harbor.ClientMgr
	harbor.ClientMgr = mock.NewMockedClients()
	originProject := models.Project
	models.Project = mock.NewProjectModel()
	initClientData()
	initProjectData()
	initLogData()
	defer func() {
		harbor.ClientMgr = originClients
		models.Project = originProject
	}()

	days := 3
	startTime := time.Date(2018, time.January, 1, 12, 0, 0, 0, time.UTC).Truncate(time.Hour * 24)
	endTime := startTime.AddDate(0, 0, days).Add(-time.Millisecond)
	cnt, stats, err := ListRegistryStats(nil, "t1", "r2", "pull", startTime.Unix(), endTime.Unix())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if cnt != days {
		fmt.Errorf("Get registry stats error, expected %d, but got %d", days, cnt)
		return
	}

	day := startTime.Truncate(time.Hour * 24)
	truth := []*types.StatItem{
		{day.AddDate(0, 0, 2).Unix(), 3},
		{day.AddDate(0, 0, 1).Unix(), 2},
		{day.Unix(), 1},
	}
	if !reflect.DeepEqual(stats, truth) {
		t.Errorf("Get registry status error, expected %s, but got %s", spew.Sdump(truth), spew.Sdump(stats))
	}
}
