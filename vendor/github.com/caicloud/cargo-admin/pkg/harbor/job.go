package harbor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	. "github.com/caicloud/cargo-admin/pkg/errors"

	"github.com/caicloud/nirvana/log"
)

// 注：JobContinue 这种状态只可能存在于 harbor 的内存中，不会出现在 harbor 的数据库中
// 因此，请求 jobs 时，永远不可能范湖 JobContinue 这种状态
// JobPending, JobRunning 和 JobRetrying 表示这个 Job 在执行中，属于运行时状态
// JobError, JobStopped, JobFinished 和 JobCanceled 表示这个 Job 已经运行结束，属于最终状态
const (
	//JobPending ...
	JobPending string = "pending"
	//JobRunning ...
	JobRunning string = "running"
	//JobError ...
	JobError string = "error"
	//JobStopped ...
	JobStopped string = "stopped"
	//JobFinished ...
	JobFinished string = "finished"
	//JobCanceled ...
	JobCanceled string = "canceled"
	//JobRetrying indicate the job needs to be retried, it will be scheduled to the end of job queue by statemachine after an interval.
	JobRetrying string = "retrying"
	//JobContinue is the status returned by statehandler to tell statemachine to move to next possible state based on trasition table.
	JobContinue string = "_continue"

	JobStopAction string = "stop"
)

type RepJobClienter interface {
	ListRepJobs(params *ListRepoJobsParams) ([]*HarborRepJob, error)
	StopRepJobs(policyId int64) error
}

type ListRepoJobsParams struct {
	PolicyId   int64  `json:"policyId"`
	Repository string `json:"repository"`
	// Status     string    `json:"status"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// 谨记，返回的结果是按照时间倒序排序的！！！
func (c *Client) ListRepJobs(params *ListRepoJobsParams) ([]*HarborRepJob, error) {
	if params.StartTime.Unix() == params.EndTime.Unix() {
		params.EndTime = params.EndTime.Add(time.Second)
	}
	path := APIJobsRepPathWithQuery(params.PolicyId, params.Repository, params.StartTime.Unix(), params.EndTime.Unix())

	log.Infof("%s %s", http.MethodGet, path)
	resp, err := c.do(http.MethodGet, path, nil)
	if err != nil {
		return nil, ErrorUnknownInternal.Error(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrorUnknownInternal.Error(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 == 2 {
		ret := make([]*HarborRepJob, 0)
		err := json.Unmarshal(body, &ret)
		if err != nil {
			return nil, ErrorUnknownInternal.Error(err)
		}
		return ret, nil
	}
	log.Errorf("list harbor replication logs error: %s", body)

	return nil, ErrorUnknownInternal.Error(string(body))
}

func (c *Client) StopRepJobs(policyId int64) error {
	path := APIJobsRepPath()

	req := &HarborStopJobsReq{
		PolicyID: policyId,
		Status:   JobStopAction,
	}
	b, err := json.Marshal(req)
	if err != nil {
		return ErrorUnknownInternal.Error(err)
	}

	log.Infof("%s %s", http.MethodPut, path)
	resp, err := c.do(http.MethodPut, path, bytes.NewReader(b))
	if err != nil {
		return ErrorUnknownInternal.Error(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ErrorUnknownInternal.Error(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 == 2 {
		log.Errorf("stop harbor replication: %d 's jobs successfully", policyId)
		return nil
	}
	log.Errorf("stop harbor replication: %d 's jobs error: %s", policyId, body)
	return ErrorUnknownInternal.Error(string(body))
}

func (c *Client) GetJob(policyID, jobID int64, repository string, creationTime time.Time) (*HarborRepJob, error) {
	params := &ListRepoJobsParams{
		PolicyId:   policyID,
		Repository: repository,
		StartTime:  creationTime.Add(-time.Second),
		EndTime:    creationTime.Add(time.Second),
	}
	jobs, err := c.ListRepJobs(params)
	if err != nil {
		return nil, err
	}

	for _, job := range jobs {
		if job.ID == jobID {
			return job, nil
		}
	}

	return nil, fmt.Errorf("Job %d not found", jobID)
}
