package resource

import (
	"context"
	"time"

	"github.com/caicloud/cargo-admin/pkg/api/admin/types"
	. "github.com/caicloud/cargo-admin/pkg/errors"
	"github.com/caicloud/cargo-admin/pkg/harbor"
	"github.com/caicloud/cargo-admin/pkg/models"

	"github.com/caicloud/nirvana/log"
)

const (
	StatusSuccess string = "Success"
	StatusFailed  string = "Failed"
	StatusSyncing string = "Syncing"
)

type ListRecordsParams struct {
	Replication string
	Status      string
	TriggerType string
	Start       int
	Limit       int
}

// Get the latest record status
func GetRecord(ctx context.Context, tenant, replication, recordID string) (*types.Record, error) {
	recordInfo, err := models.Record.FindOne(recordID)
	if err != nil {
		return nil, err
	}

	recordImages, err := models.RecordImage.FindAllByRecord(recordID)
	if err != nil {
		return nil, err
	}

	var success, failed int64 = 0, 0
	syncingJobs := make([]*models.RecordImageInfo, 0)
	for _, recordImage := range recordImages {
		status := convertStatus(recordImage.Status)
		if status == StatusSuccess {
			success++
		} else if status == StatusFailed {
			failed++
		} else {
			syncingJobs = append(syncingJobs, recordImage)
		}
	}

	record := &types.Record{
		Id:           recordInfo.Id,
		Replication:  nil, // It's too expensive to make replication
		Trigger:      getRecReplicationTrigger(recordInfo),
		SuccessCount: success,
		FailedCount:  failed,
		Status:       "",
		StartTime:    recordInfo.CreationTime,
		EndTime:      recordInfo.LastUpdateTime,
	}

	// If the record status is not syncing, the record is already in final status, success or failed
	// no need to check latest status. We can return directly.
	if recordInfo.Status != StatusSyncing {
		return record, nil
	}

	repInfo, err := models.Replication.FindByName(tenant, replication)
	cli, err := harbor.ClientMgr.GetClient(repInfo.SourceRegistry)
	if err != nil {
		return nil, err
	}

	// Check latest status of all jobs that are in syncing status from Harbor
	isSyncing := false
	newStatus := StatusSuccess
	for _, j := range syncingJobs {
		job, err := cli.GetJob(repInfo.ReplicationPolicyId, j.RepJobId, j.Repository, j.CreationTime)
		if err != nil {
			continue
		}
		status := convertStatus(job.Status)
		if status == StatusSuccess {
			success++
		} else if status == StatusFailed {
			newStatus = StatusFailed
			failed++
		} else {
			isSyncing = true
		}
	}
	record.SuccessCount = success
	record.FailedCount = failed
	if !isSyncing {
		record.Status = newStatus
	}

	return record, nil
}

// ListRecords 函数的主要用途是获取 list replication 下的 records 信息
// 主要步骤：
// 1. 获取 replication 相关的所有信息，包括 mongodb 中存储的 replication 和 registry 中对应的 replication policy；
// 2. 从 mongodb 中获取 replication 涉及的 source registry 和 target registry；
// 3. 通过 listRecordInfos 获取最近一次 recinfos 和 repJobs，相当于是一次预处理；
// 4. 通过 filterAndSave 函数处理 listRecordInfos 返回的 recinfos；
// 5. 更新 repinfo 中的 last_list_records_time 字段；
// 6. 直接从 mongodb 的 record collection 中 findOnePageRetRecords；
// 7. 启动一个 gorutine 去更新 mongodb 中 record_image collecction 中 image document 的状态。
func ListRecords(ctx context.Context, tenant string, param *ListRecordsParams) (int, []*types.Record, error) {
	log.Infof("list replicaiton: %s 's records, status: %s, trigger type: %s, start: %d, limit: %d",
		param.Replication, param.Status, param.TriggerType, param.Start, param.Limit)
	repinfo, hrp, err := getReplicationDetails(tenant, param.Replication)
	if err != nil {
		log.Errorf("getReplicationDetails error: %v", err)
		return 0, nil, err
	}
	log.Infof("CreationTime: %s, LastUpdateTime: %s, LastTriggerTime: %s, LastListRecordsTime: %s",
		repinfo.CreationTime, repinfo.LastUpdateTime, repinfo.LastTriggerTime, repinfo.LastListRecordsTime)
	sreginfo, treginfo, err := getReplicationRegistry(repinfo.SourceRegistry, repinfo.TargetRegistry)
	if err != nil {
		return 0, nil, err
	}
	recinfos, repJobs, err := makeLatestRecords(repinfo, hrp, sreginfo)
	if err != nil {
		log.Error(err)
		return 0, nil, err
	}
	err = filterAndSave(tenant, recinfos, repJobs)
	if err != nil {
		log.Error(err)
		return 0, nil, err
	}

	err = models.Replication.UpdateLastListRecordsTime(tenant, repinfo.Name, time.Now())
	if err != nil {
		log.Error(err)
		return 0, nil, ErrorUnknownInternal.Error(err)
	}

	total, ret, err := findOnePageRetRecords(repinfo, hrp, sreginfo, treginfo, param)
	if err != nil {
		log.Error(err)
		return 0, nil, err
	}

	go func() { log.Errorf("updateSyncingRecordStatus error: %v", updateSyncingRecordStatus(repinfo)) }()

	return total, ret, nil
}

// =================================================================================================

func getReplicationDetails(tenant string, replication string) (*models.ReplicationInfo, *harbor.HarborReplicationPolicy, error) {
	repinfo, err := models.Replication.FindByName(tenant, replication)
	if err != nil {
		log.Error(err)
		return nil, nil, ErrorUnknownInternal.Error(err)
	}

	cli, err := harbor.ClientMgr.GetClient(repinfo.SourceRegistry)
	if err != nil {
		return nil, nil, err
	}
	hrp, err := cli.GetReplicationPolicy(repinfo.ReplicationPolicyId)
	if err != nil {
		log.Errorf("get replication policy: %d from registry: %s error: %v", repinfo.ReplicationPolicyId, repinfo.SourceRegistry, err)
		return nil, nil, ErrorUnknownInternal.Error(err)
	}
	return repinfo, hrp, nil
}

// =================================================================================================

func getReplicationRegistry(source, target string) (sreginfo, treginfo *models.RegistryInfo, err error) {
	sreginfo, err = models.Registry.FindByName(source)
	if err != nil {
		log.Error(err)
		return nil, nil, ErrorUnknownInternal.Error(err)
	}
	treginfo, err = models.Registry.FindByName(target)
	if err != nil {
		log.Error(err)
		return nil, nil, ErrorUnknownInternal.Error(err)
	}
	return sreginfo, treginfo, nil
}

// =================================================================================================

// 传入 recinfos 切片的元素可能存在 nil，因此，需要先 filter 再 save 到数据库
func filterAndSave(tenant string, recinfos []*models.RecordInfo, repJobs []*harbor.HarborRepJob) error {
	saveRecInfos := make([]*models.RecordInfo, 0, len(recinfos))
	for _, recinfo := range recinfos {
		if recinfo != nil {
			saveRecInfos = append(saveRecInfos, recinfo)
		}
	}

	if len(saveRecInfos) != 0 {
		err := models.Record.SaveBatch(saveRecInfos)
		if err != nil {
			log.Errorf("save batch of recordinfos error: %v", err)
			return ErrorUnknownInternal.Error(err)
		}
	}

	if len(saveRecInfos) != 0 {
		recimginfos := getRecordImages(tenant, saveRecInfos, repJobs)
		err := models.RecordImage.SaveBatch(recimginfos)
		if err != nil {
			log.Error(err)
			return ErrorUnknownInternal.Error(err)
		}
	}
	return nil
}

func getRecordImages(tenant string, recinfos []*models.RecordInfo, repJobs []*harbor.HarborRepJob) []*models.RecordImageInfo {
	ret := make([]*models.RecordImageInfo, 0)
	repJobMap := convertHarborRepJobsToMap(repJobs)
	for _, recinfo := range recinfos {
		for _, repJobId := range recinfo.RepJobIds {
			repJob, ok := repJobMap[repJobId]
			if !ok {
				continue
			}
			if len(repJob.TagList) == 0 {
				// TODO(li-ang): 这种情况下，可能需要直接从 registry 中直接 list repository 的所有 tags
			}
			for _, tag := range repJob.TagList {
				ret = append(ret, &models.RecordImageInfo{
					RecordId:       recinfo.Id,
					Tenant:         tenant,
					Registry:       recinfo.Registry,
					Replication:    recinfo.Replication,
					RepJobId:       repJobId,
					Repository:     repJob.Repository,
					Tag:            tag,
					Operation:      repJob.Operation,
					Status:         repJob.Status,
					CreationTime:   repJob.CreationTime,
					LastUpdateTime: repJob.UpdateTime,
				})
			}
		}
	}
	return ret
}

func convertHarborRepJobsToMap(repJobs []*harbor.HarborRepJob) map[int64]*harbor.HarborRepJob {
	ret := make(map[int64]*harbor.HarborRepJob)
	for _, repJob := range repJobs {
		ret[repJob.ID] = repJob
	}
	return ret
}

func convertStatus(repJobStatus string) string {
	status := StatusSuccess

	if repJobStatus == harbor.JobRunning ||
		repJobStatus == harbor.JobPending ||
		repJobStatus == harbor.JobRetrying {
		status = StatusSyncing
		return status
	}
	if repJobStatus == harbor.JobError ||
		repJobStatus == harbor.JobStopped ||
		repJobStatus == harbor.JobCanceled {
		status = StatusFailed
	}
	return status
}

func revertStatus(recstatus string) []string {
	ret := make([]string, 0)
	switch recstatus {
	case StatusSuccess:
		ret = append(ret, harbor.JobFinished)
		return ret
	case StatusFailed:
		ret = append(ret, harbor.JobError, harbor.JobStopped, harbor.JobCanceled)
		return ret
	case StatusSyncing:
		ret = append(ret, harbor.JobRunning, harbor.JobPending)
		return ret
	}
	return ret
}

// =================================================================================================

func updateSyncingRecordStatus(repinfo *models.ReplicationInfo) error {
	recinfos, err := models.Record.FindByStatus(repinfo.Name, StatusSyncing)
	if err != nil {
		return err
	}
	if len(recinfos) == 0 {
		return nil
	}

	for _, recinfo := range recinfos {
		startTime := recinfo.CreationTime
		endTime := recinfo.LastUpdateTime
		repJobs, err := updateRecordImageStatus(repinfo, startTime, endTime)
		if err != nil {
			log.Errorf("updateRecordImageStatus error: %v", err)
			return err
		}

		status := getRecordStatus(repJobs)
		if recinfo.Status != status {
			err = models.Record.UpdateStatus(recinfo.Replication, recinfo.Id, status)
			if err != nil {
				log.Errorf("update status error: %v", err)
			}
		}
	}
	return nil
}

// record 和 record image collection 中的 status 更新策略
// record 的 status 字段实际上是由 record image 的 status 字段决定的，因此要先更新 record image 的 status 字段才行
func updateRecordImageStatus(repinfo *models.ReplicationInfo, startTime, endTime time.Time) ([]*harbor.HarborRepJob, error) {
	cli, err := harbor.ClientMgr.GetClient(repinfo.SourceRegistry)
	if err != nil {
		return make([]*harbor.HarborRepJob, 0), err
	}

	param := &harbor.ListRepoJobsParams{
		PolicyId:   repinfo.ReplicationPolicyId,
		Repository: repinfo.Project,
		StartTime:  startTime,
		EndTime:    endTime,
	}
	repJobs, err := cli.ListRepJobs(param)
	if err != nil {
		log.Errorf("list replication jobs from registry: %s error: %v", repinfo.SourceRegistry, err)
		log.Errorf("list param, PolicyId: %d, Repository: %s, StartTime: %v, EndTime: %v",
			param.PolicyId, param.Repository, param.StartTime, param.EndTime)
		return make([]*harbor.HarborRepJob, 0), err
	}
	if len(repJobs) == 0 {
		log.Info("repjob's length is zero")
		log.Infof("list param, PolicyId: %d, Repository: %s, StartTime: %v, EndTime: %v",
			param.PolicyId, param.Repository, param.StartTime, param.EndTime)
		return make([]*harbor.HarborRepJob, 0), nil
	}
	for _, repJob := range repJobs {
		err = models.RecordImage.UpdateBatchStatus(repJob.ID, repJob.Status)
		if err != nil {
			log.Errorf("update batch of record image 's status error: %v, repJob id: %s, repJob status: %v", err, repJob.ID, repJob.Status)
		}
	}
	return repJobs, nil
}

// =================================================================================================

func findOnePageRetRecords(repinfo *models.ReplicationInfo, hrp *harbor.HarborReplicationPolicy,
	sreginfo, treginfo *models.RegistryInfo, param *ListRecordsParams) (int, []*types.Record, error) {
	total, retrecinfos, err := models.Record.FindOnePage(repinfo.Name, param.TriggerType, param.Status, param.Start, param.Limit)
	if err != nil {
		log.Errorf("mongo error: %v", err)
		return 0, nil, err
	}
	ret := make([]*types.Record, 0, len(retrecinfos))
	recReplication := getRecReplication(repinfo, hrp, sreginfo, treginfo, retrecinfos)
	recimginfoMap, err := getRecordImageInfoMap(repinfo.Name)
	if err != nil {
		return 0, nil, err
	}
	for _, recinfo := range retrecinfos {
		success, failed, _ := getRecordImageCounts(recimginfoMap, recinfo.Id.String())
		lastUpdateTime := recinfo.LastUpdateTime
		if recinfo.Status == StatusSyncing {
			lastUpdateTime = time.Now()
		}
		r := &types.Record{
			Id:           recinfo.Id,
			Replication:  recReplication,
			Trigger:      getRecReplicationTrigger(recinfo),
			SuccessCount: success,
			FailedCount:  failed,
			Status:       recinfo.Status,
			// Reason:
			StartTime: recinfo.CreationTime,
			EndTime:   lastUpdateTime,
		}
		ret = append(ret, r)
	}
	return total, ret, nil
}

func getRecReplication(repinfo *models.ReplicationInfo, hrp *harbor.HarborReplicationPolicy,
	sreginfo, treginfo *models.RegistryInfo, recinfos []*models.RecordInfo) *types.Replication {
	isSyncing := repIsSyncing(recinfos)
	return &types.Replication{
		Metadata: &types.ReplicationMetadata{
			Name:           repinfo.Name,
			Alias:          repinfo.Alias,
			CreationTime:   repinfo.CreationTime,
			LastUpdateTime: repinfo.LastUpdateTime,
		},
		Spec: &types.ReplicationSpec{
			Project:           repinfo.Project,
			ReplicateNow:      hrp.ReplicateExistingImageNow,
			ReplicateDeletion: hrp.ReplicateDeletion,
			Source: &types.ReplicationSource{
				Name:   sreginfo.Name,
				Alias:  sreginfo.Alias,
				Domain: sreginfo.Domain,
			},
			Target: &types.ReplicationTarget{
				Name:   treginfo.Name,
				Alias:  treginfo.Alias,
				Domain: treginfo.Domain,
			},
			Trigger: getReplicationTrigger(hrp.Trigger),
		},
		Status: &types.ReplicationStatus{
			IsSyncing:           isSyncing,
			LastReplicationTime: repinfo.LastTriggerTime,
		},
	}
}

func repIsSyncing(recinfos []*models.RecordInfo) bool {
	ret := false
	for _, recinfo := range recinfos {
		if recinfo.Status == StatusSyncing {
			ret = true
		}
	}
	return ret
}

func getRecReplicationTrigger(recinfo *models.RecordInfo) *types.ReplicationTrigger {
	ret := &types.ReplicationTrigger{Kind: recinfo.Trigger.Kind}
	if recinfo.Trigger.ScheduleParam != nil {
		ret.ScheduleParam = &types.ScheduleParam{
			Type:    recinfo.Trigger.ScheduleParam.Type,
			Weekday: recinfo.Trigger.ScheduleParam.Weekday,
			Offtime: recinfo.Trigger.ScheduleParam.Offtime,
		}
	}
	return ret
}

func getRecordImageInfoMap(replication string) (map[string][]*models.RecordImageInfo, error) {
	ret := make(map[string][]*models.RecordImageInfo)
	recimginfos, err := models.RecordImage.FindAllByReplication(replication)
	if err != nil {
		return nil, err
	}

	for _, recimginfo := range recimginfos {
		_, ok := ret[recimginfo.RecordId.String()]
		if !ok {
			ret[recimginfo.RecordId.String()] = make([]*models.RecordImageInfo, 0)
		}
		ret[recimginfo.RecordId.String()] = append(ret[recimginfo.RecordId.String()], recimginfo)
	}
	return ret, nil
}

func getRecordImageCounts(recimginfoMap map[string][]*models.RecordImageInfo, recodId string) (success, failed, sycning int64) {
	recimginfos, ok := recimginfoMap[recodId]
	if !ok {
		return 0, 0, 0
	}
	for _, recimginfo := range recimginfos {
		switch convertStatus(recimginfo.Status) {
		case StatusSuccess:
			success += 1
		case StatusFailed:
			failed += 1
		case StatusSyncing:
			sycning += 1
		default:
			continue
		}
	}
	return success, failed, sycning
}
