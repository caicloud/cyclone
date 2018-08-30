package resource

// ListReplicationRecords 的实现比较复杂，主要体现在以下几个方面：

// =================================================================================================

/*
一、ListReplicationRecords 函数应该在何时被调用

1. 请求 List /replications/{replication}/records API 时，一定会调用此函数，属于直接调用；
2. update replication 前，可能会调用此函数，属于间接调用；
3. trigger replication 前，可能会调用此函数，属于间接调用。
*/

// =================================================================================================

/*
二、调用 ListReplicationRecords 函数的前一步操作，可能会影响 ListReplicationRecords 函数返回的结果，
那么，到底是哪几种操作会影响 ListReplicationRecords 函数返回的结果

1. 在此次调用 ListReplicationRecords 前，对 replication 的前一次操作为：ListReplicationRecords；
2. 在此次调用 ListReplicationRecords 前，对 replication 的前一次操作为：update replication；
3. 在此次调用 ListReplicationRecords 前，对 replication 的前一次操作为：trigger replication；
4. 第一次调用 ListReplicationRecords。
*/

// =================================================================================================

/*
三、ListReplicationRecords 的实现逻辑：

首先，以下几个变量的值非常重要：

1. 在 replication 中查询上次 update replication 的时间：LastUpdateTime
2. 在 replication 中查询上次 trigger replication 的时间：LastTriggerTime
3. 在 replication 中查询上次 ListReplicationRecords 的时间：LastListRecordsTime

然后，根据以上三个时间变量，判断调用 ListReplicationRecords 函数前，对 replication 的前一次操作是什么：

在此次调用 ListReplicationRecords 前：
switch {
	case replication.LastTriggerTime > replication.LastUpdateTime: {
		对 replication 的前一次操作，肯定不是 update replication，只可能是 trigger replication 和 ListReplicationRecords 之一
    	if replication.LastTriggerTime > replication.LastListRecordsTime {
    		对 replication 的前一次操作为：trigger replication
    	} else {
    		对 replication 的前一次操作为：ListReplicationRecords
    	}
	}
	case replication.LastTriggerTime < replication.LastUpdateTime: {
		对 replication 的前一次操作，肯定不是 trigger replication
    	if replication.LastUpdateTime > replication.LastListRecordsTime {
    		对 replication 的前一次操作为：update replication
    	} else {
    		对 replication 的前一次操作为：ListReplicationRecords
    	}
    }
    case replication.LastTriggerTime == replication.LastUpdateTime: {
		if replication.LastTriggerTime == replication.LastListRecordsTime {
			肯定是第一次 ListReplicationRecords
			先删除 replication 下的所有record，然后再调用 ListReplicationRecords 函数
		} else {
			对 replication 的前一次操作为：ListReplicationRecords
		}
    }
    defaut: error
}

再然后，根据以上的判断，ListReplicationRecords 会有不同的处理逻辑：

repJobs 归类方案：相邻时间的两个 repoJobs 的时间差小于 2s，那么被认为是同一次 replication 触发，被归类为一组 records

ListReplicationRecords 处理逻辑：

if 第一次调用 ListReplicationRecords {
	repJobs, err := cli.ListRepJobs(), start_time = replication.LastUpdateTime, end_time = time.Now()
	records := convertRepJobsToRecord(repJobs)
	if replication.replicateNow == true && record.StartTime <=> replication.LastUpdateTime {
		这一组都是 Manual 触发
	}
	if replication.Trigger.Type == Scheduled {
		if record.StartTime%time.Hour * 24 <=> replication.Trgger.ScheduleParam.Offtime {
			这一组都是 Scheduled 触发
		} else {
			丢弃，不存在这种情况
		}
	}
	if replication.Trigger.Type == OnPush {
		这一组都是 OnPush 触发
	}
	if replication.Trigger.Type == Manual {
		丢弃，不存在这种请求
	}
}

if 对 replication 的前一次操作，是普通的一次 ListReplicationRecords {
	repJobs, err := cli.ListRepJobs(), start_time = replication.LastListRecordsTime, end_time = time.Now()
	records := convertRepJobsToRecord(repJobs)
	if replication.Trigger.Type == Scheduled {
		if replication.Trgger.ScheduleParam.Offtime <=> record.StartTime%time.Hour * 24 {
			这一组都是 Scheduled 触发
		} else {
			丢弃，不存在这种情况
		}
	}
	if replication.Trigger.Type == OnPush {
		这一组都是 OnPush 触发
	}
	if replication.Trigger.Type == Manual {
		丢弃，不存在这种请求
	}
}

if 对 replication 的前一次操作，是 trigger replication {
	repoJobs, err := cli.ListRepJobs(), start_time = repolication.LastTriggerTime, end_time = time.Now()
	records := convertRepJobsToRecord(repJobs)
	if  replication.Trigger.Type == Scheduled {
		if replication.Trgger.ScheduleParam.Offtime <=> record.StartTime%time.Hour * 24 {
			这一组都是 Scheduled 触发
		} else if replication.LastTriggerTime <=> record.StartTime {
			这一组都是 Manual 触发
		} else {
			丢弃，不存在这种情况
		}
	}
	if replication.Trigger.Type == OnPush {
		if replication.LastTriggerTime <=> record.StartTime {
			这一组都是 Manual 触发
		} else {
			这一组都是 OnPush 触发
		}
	}
	if replication.Trigger.Type == Manual {
		这一组都是 Manual 触发
	}
}

if 对 replication 的前一次操作，是 update replication {
	repJobs, err := cli.ListRepJobs(), start_time = repolication.LastUpdateTime, end_time = time.Now()
	records := convertRepJobsToRecord(repJobs)
	if oldReplication.replicateNow == true && record.StartTime <=> replication.LastUpdateTime {
		这一组都是 Manual 触发
	}
	if oldReplication.Trigger.Type == Scheduled {
		if replication.Trgger.ScheduleParam.Offtime <=> record.StartTime%time.Hour * 24 {
			这一组都是 Scheduled 触发
		} else {
			丢弃，不存在这种情况
		}
	}
	if replication.Trigger.Type == OnPush {
		这一组都是 OnPush 触发
	}
	if replication.Trigger.Type == Manual {
		丢弃，不存在这种情况
	}
}

最后，要更新 replication 的 LastListRecordsTime，把 record 写入数据库
*/
