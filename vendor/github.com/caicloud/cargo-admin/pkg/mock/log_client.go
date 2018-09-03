package mock

import (
	"sort"

	"github.com/caicloud/cargo-admin/pkg/harbor"
)

type LogClient struct {
	logs []*harbor.HarborAccessLog
}

func NewLogClient() *LogClient {
	return &LogClient{
		logs: make([]*harbor.HarborAccessLog, 0),
	}
}

func (client *LogClient) Len() int {
	return len(client.logs)
}

func (client *LogClient) Less(i, j int) bool {
	return client.logs[i].OpTime.After(client.logs[j].OpTime)
}

func (client *LogClient) Swap(i, j int) {
	client.logs[i], client.logs[j] = client.logs[j], client.logs[i]
}

func (client *LogClient) Add(log *harbor.HarborAccessLog) {
	client.logs = append(client.logs, log)
}

func (client *LogClient) ListLogs(startTime int64, endTime int64, op string) ([]*harbor.HarborAccessLog, error) {
	sort.Sort(client)
	return client.logs, nil
}

func (client *LogClient) ListProjectLogs(pid int64, startTime int64, endTime int64, op string) ([]*harbor.HarborAccessLog, error) {
	sort.Sort(client)
	return client.logs, nil
}
