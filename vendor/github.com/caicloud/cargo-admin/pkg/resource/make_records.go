package resource

import (
	"fmt"
	"math"
	"time"

	"github.com/caicloud/cargo-admin/pkg/harbor"
	"github.com/caicloud/cargo-admin/pkg/models"

	"github.com/caicloud/nirvana/log"

	"gopkg.in/mgo.v2/bson"
)

// Get all unprocessed replication jobs for a replication and group them into replication records. We don't process all jobs
// here because this is an incremental operation, old jobs have already been processed and stored in Mongo as replication records.
func makeLatestRecords(repInfo *models.ReplicationInfo, hRep *harbor.HarborReplicationPolicy, sRegInfo *models.RegistryInfo) ([]*models.RecordInfo, []*harbor.HarborRepJob, error) {
	var records []*models.RecordInfo
	var repJobs []*harbor.HarborRepJob
	var err error
	switch {
	case repInfo.LastTriggerTime.After(repInfo.LastUpdateTime):
		if repInfo.LastTriggerTime.After(repInfo.LastListRecordsTime) {
			log.Info("make latest records: previous replication operation is trigger")
			records, repJobs, err = listAfterTrigger(repInfo, sRegInfo, hRep)
		} else {
			log.Info("make latest records: previous replication operation is list records")
			records, repJobs, err = listAfterList(repInfo, sRegInfo, hRep)
		}
	case repInfo.LastUpdateTime.After(repInfo.LastTriggerTime):
		if repInfo.LastUpdateTime.After(repInfo.LastListRecordsTime) {
			log.Info("make latest records: previous replication operation is update")
			records, repJobs, err = listAfterUpdate(repInfo, sRegInfo, hRep)
		} else {
			log.Info("make latest records: previous replication operation is list records")
			records, repJobs, err = listAfterList(repInfo, sRegInfo, hRep)
		}
	case repInfo.LastTriggerTime == repInfo.LastUpdateTime:
		if repInfo.LastTriggerTime == repInfo.LastListRecordsTime {
			log.Info("make latest records: the first time to make records, delete all records first if any")
			err = models.Record.DeleteAllByReplication(repInfo.Name)
			if err != nil {
				log.Error(err)
				return nil, nil, err
			}
			records, repJobs, err = firstList(repInfo, sRegInfo, hRep)
		} else {
			log.Info("make latest records: previous replication operation is list records")
			records, repJobs, err = listAfterList(repInfo, sRegInfo, hRep)
		}
	default:
		err = fmt.Errorf("replicaiton: %s is in unexceptable status, please delete it and re-create", repInfo.Name)
		log.Error(err)
		return nil, nil, err
	}
	return records, repJobs, nil
}

func firstList(repInfo *models.ReplicationInfo, regInfo *models.RegistryInfo, hRep *harbor.HarborReplicationPolicy) ([]*models.RecordInfo, []*harbor.HarborRepJob, error) {
	cli, err := harbor.ClientMgr.GetClient(repInfo.SourceRegistry)
	if err != nil {
		return nil, nil, err
	}
	param := &harbor.ListRepoJobsParams{
		PolicyId:   repInfo.ReplicationPolicyId,
		Repository: repInfo.Project,
		StartTime:  repInfo.LastUpdateTime,
		EndTime:    time.Now(),
	}
	repJobs, err := cli.ListRepJobs(param)
	if err != nil {
		log.Infof("list replicatoin jobs error: %v", err)
		return nil, nil, err
	}
	if len(repJobs) == 0 {
		log.Info("repJobs length is zero")
		return make([]*models.RecordInfo, 0), repJobs, nil
	}

	records := getRecordInfos(regInfo, repInfo.Name, repJobs)
	for i, _ := range records {
		if hRep.ReplicateExistingImageNow && isNearly(records[i].CreationTime, repInfo.LastUpdateTime) {
			records[i].Trigger.Kind = TriggerKindManual
		}
		if hRep.Trigger.Kind == harbor.TriggerKindSchedule {
			if isScheduled(records[i], hRep.Trigger.ScheduleParam) {
				records[i].Trigger = &models.Trigger{
					Kind: hRep.Trigger.Kind,
					ScheduleParam: &models.ScheduleParam{
						Type:    hRep.Trigger.ScheduleParam.Type,
						Weekday: hRep.Trigger.ScheduleParam.Weekday,
						Offtime: hRep.Trigger.ScheduleParam.Offtime,
					},
				}
			} else {
				log.Warningf("firstList: unexpected record observed: %v", records[i])
				records[i] = nil
			}
		}
		if hRep.Trigger.Kind == harbor.TriggerKindImmediate {
			records[i].Trigger.Kind = TriggerKindOnPush
		}

		// This should not happen by design
		if hRep.Trigger.Kind == harbor.TriggerKindManual {
			log.Warningf("firstList: unexpected record observed: %v", records[i])
			records[i] = nil
		}
	}

	return records, repJobs, nil
}

func listAfterList(repInfo *models.ReplicationInfo, regInfo *models.RegistryInfo, hRep *harbor.HarborReplicationPolicy) ([]*models.RecordInfo, []*harbor.HarborRepJob, error) {
	cli, err := harbor.ClientMgr.GetClient(repInfo.SourceRegistry)
	if err != nil {
		return nil, nil, err
	}
	param := &harbor.ListRepoJobsParams{
		PolicyId:   repInfo.ReplicationPolicyId,
		Repository: repInfo.Project,
		StartTime:  repInfo.LastListRecordsTime,
		EndTime:    time.Now(),
	}
	repJobs, err := cli.ListRepJobs(param)
	if err != nil {
		log.Errorf("list replication jobs error: %v", err)
		log.Errorf("list param, PolicyId: %d, Repository: %s, StartTime: %v, EndTime: %v",
			param.PolicyId, param.Repository, param.StartTime, param.EndTime)
		return nil, nil, err
	}
	if len(repJobs) == 0 {
		return make([]*models.RecordInfo, 0), repJobs, nil
	}
	records := getRecordInfos(regInfo, repInfo.Name, repJobs)
	for i, _ := range records {
		if hRep.Trigger.Kind == harbor.TriggerKindSchedule {
			if isScheduled(records[i], hRep.Trigger.ScheduleParam) {
				records[i].Trigger = &models.Trigger{
					Kind: hRep.Trigger.Kind,
					ScheduleParam: &models.ScheduleParam{
						Type:    hRep.Trigger.ScheduleParam.Type,
						Weekday: hRep.Trigger.ScheduleParam.Weekday,
						Offtime: hRep.Trigger.ScheduleParam.Offtime,
					},
				}
			} else {
				log.Warningf("listAfterList: unexpected record observed: %v", records[i])
				records[i] = nil
			}
		}
		if hRep.Trigger.Kind == harbor.TriggerKindImmediate {
			records[i].Trigger.Kind = TriggerKindOnPush
		}

		// This should not happen by design
		if hRep.Trigger.Kind == harbor.TriggerKindManual {
			log.Warningf("listAfterList: unexpected record observed: %v", records[i])
			records[i] = nil
		}
	}
	return records, repJobs, nil
}

func listAfterTrigger(repInfo *models.ReplicationInfo, regInfo *models.RegistryInfo, hRep *harbor.HarborReplicationPolicy) ([]*models.RecordInfo, []*harbor.HarborRepJob, error) {
	cli, err := harbor.ClientMgr.GetClient(repInfo.SourceRegistry)
	if err != nil {
		return nil, nil, err
	}
	param := &harbor.ListRepoJobsParams{
		PolicyId:   repInfo.ReplicationPolicyId,
		Repository: repInfo.Project,
		StartTime:  repInfo.LastTriggerTime,
		EndTime:    time.Now(),
	}
	repJobs, err := cli.ListRepJobs(param)
	if err != nil {
		log.Errorf("list replication jobs error: %v", err)
		log.Errorf("list param, PolicyId: %d, Repository: %s, StartTime: %v, EndTime: %v",
			param.PolicyId, param.Repository, param.StartTime, param.EndTime)
		return nil, nil, err
	}
	if len(repJobs) == 0 {
		log.Info("returned replication jobs is empyty")
		return make([]*models.RecordInfo, 0), repJobs, nil
	}
	records := getRecordInfos(regInfo, repInfo.Name, repJobs)
	for i, _ := range records {
		if hRep.Trigger.Kind == harbor.TriggerKindSchedule {
			if isScheduled(records[i], hRep.Trigger.ScheduleParam) {
				records[i].Trigger = &models.Trigger{
					Kind: hRep.Trigger.Kind,
					ScheduleParam: &models.ScheduleParam{
						Type:    hRep.Trigger.ScheduleParam.Type,
						Weekday: hRep.Trigger.ScheduleParam.Weekday,
						Offtime: hRep.Trigger.ScheduleParam.Offtime,
					},
				}
			} else if isNearly(repInfo.LastTriggerTime, records[i].CreationTime) {
				records[i].Trigger.Kind = TriggerKindManual
			} else {
				log.Warningf("listAfterTrigger: unexpected record observed: %v", records[i])
				records[i] = nil
			}
		}
		if hRep.Trigger.Kind == harbor.TriggerKindImmediate {
			if isNearly(repInfo.LastTriggerTime, records[i].CreationTime) {
				records[i].Trigger.Kind = TriggerKindManual
			} else {
				records[i].Trigger.Kind = TriggerKindOnPush
			}
		}
		if hRep.Trigger.Kind == harbor.TriggerKindManual {
			records[i].Trigger.Kind = TriggerKindManual
		}
	}

	return records, repJobs, nil
}

func listAfterUpdate(repInfo *models.ReplicationInfo, regInfo *models.RegistryInfo, hRep *harbor.HarborReplicationPolicy) ([]*models.RecordInfo, []*harbor.HarborRepJob, error) {
	cli, err := harbor.ClientMgr.GetClient(repInfo.SourceRegistry)
	if err != nil {
		return nil, nil, err
	}
	param := &harbor.ListRepoJobsParams{
		PolicyId:   repInfo.ReplicationPolicyId,
		Repository: repInfo.Project,
		StartTime:  repInfo.LastUpdateTime,
		EndTime:    time.Now(),
	}
	repJobs, err := cli.ListRepJobs(param)
	if err != nil {
		return nil, nil, err
	}
	records := getRecordInfos(regInfo, repInfo.Name, repJobs)
	for i, _ := range records {
		if hRep.Trigger.Kind == harbor.TriggerKindSchedule {
			if isScheduled(records[i], hRep.Trigger.ScheduleParam) {
				records[i].Trigger = &models.Trigger{
					Kind: hRep.Trigger.Kind,
					ScheduleParam: &models.ScheduleParam{
						Type:    hRep.Trigger.ScheduleParam.Type,
						Weekday: hRep.Trigger.ScheduleParam.Weekday,
						Offtime: hRep.Trigger.ScheduleParam.Offtime,
					},
				}
			} else {
				log.Warningf("listAfterUpdate: unexpected record observed: %v", records[i])
				records[i] = nil
			}
		}
		if hRep.Trigger.Kind == harbor.TriggerKindImmediate {
			records[i].Trigger.Kind = TriggerKindOnPush
		}

		// This should not happen by design
		if hRep.Trigger.Kind == harbor.TriggerKindManual {
			log.Warningf("listAfterUpdate: unexpected record observed: %v", records[i])
			records[i] = nil
		}
	}
	return records, repJobs, nil
}

func getRecordInfos(regInfo *models.RegistryInfo, replication string, repJobs []*harbor.HarborRepJob) []*models.RecordInfo {
	records := make([]*models.RecordInfo, 0)
	splits := splitRepJobs(repJobs)
	for _, jobs := range splits {
		record := &models.RecordInfo{
			Id:             bson.NewObjectId(),
			Registry:       regInfo.Name,
			Replication:    replication,
			RepJobIds:      getRepJobsIds(jobs),
			Trigger:        &models.Trigger{}, // Empty right now, will be updated later
			Status:         getRecordStatus(jobs),
			Reason:         "", // Empty right now, will be updated later
			CreationTime:   getRecordCreateTime(jobs),
			LastUpdateTime: getRecordUpdateTime(jobs),
		}
		records = append(records, record)
	}
	return records
}

// Record creation time is the earliest creation time of jobs
func getRecordCreateTime(jobs []*harbor.HarborRepJob) time.Time {
	ret := jobs[0].CreationTime
	for _, job := range jobs {
		if ret.After(job.CreationTime) {
			ret = job.CreationTime
		}
	}
	return ret
}

// Record update time is the latest update time of jobs
func getRecordUpdateTime(jobs []*harbor.HarborRepJob) time.Time {
	ret := jobs[0].UpdateTime
	for _, job := range jobs {
		if ret.Before(job.UpdateTime) {
			ret = job.UpdateTime
		}
	}
	return ret
}

func getRepJobsIds(jobs []*harbor.HarborRepJob) []int64 {
	ret := make([]int64, 0, len(jobs))
	for _, job := range jobs {
		ret = append(ret, job.ID)
	}
	return ret
}

// Get record status from all jobs, the rules are:
// - If any job is JobRunning, JobRetrying or JobPending, record status is StatusSyncing
// - Otherwise if any job is JobError, JobStopped or JobCanceled, record status is StatusFailed
// - Otherwise, record status is StatusSuccess
func getRecordStatus(jobs []*harbor.HarborRepJob) string {
	status := StatusSuccess
	hasSyncingJob := false
	for _, job := range jobs {
		if job.Status == harbor.JobRunning ||
			job.Status == harbor.JobPending ||
			job.Status == harbor.JobRetrying {
			hasSyncingJob = true
			break
		}
		if job.Status == harbor.JobError ||
			job.Status == harbor.JobStopped ||
			job.Status == harbor.JobCanceled {
			status = StatusFailed
		}
	}

	if hasSyncingJob {
		return StatusSyncing
	}

	return status
}

// Group jobs into records based on creation time. If adjacent jobs are created within one second, they
// will be considered in a record.
func splitRepJobs(repJobs []*harbor.HarborRepJob) [][]*harbor.HarborRepJob {
	ret := make([][]*harbor.HarborRepJob, 0)
	var last *harbor.HarborRepJob
	for _, repJob := range repJobs {
		if last == nil || !isNearly(last.CreationTime, repJob.CreationTime) {
			ret = append(ret, make([]*harbor.HarborRepJob, 0))
		}
		ret[len(ret)-1] = append(ret[len(ret)-1], repJob)
		last = repJob
	}
	return ret
}

func isNearly(one, two time.Time) bool {
	return one.Add(-time.Second).Before(two) && one.Add(time.Second).After(two)
}

func isScheduled(record *models.RecordInfo, schedParam *harbor.HarborScheduleParam) bool {
	seconds := float64(record.CreationTime.Hour()*3600+record.CreationTime.Minute()*60+record.CreationTime.Second()) + float64(record.CreationTime.Nanosecond()/1000000000.0)
	return math.Abs(seconds-float64(schedParam.Offtime)) <= 2.0
}
