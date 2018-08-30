package types

import (
	"time"
)

type Replication struct {
	Metadata *ReplicationMetadata `json:"metadata"`
	Spec     *ReplicationSpec     `json:"spec"`
	Status   *ReplicationStatus   `json:"status"`
}

type ReplicationMetadata struct {
	Name           string    `json:"name"`
	Alias          string    `json:"alias"`
	CreationTime   time.Time `json:"creationTime"`
	LastUpdateTime time.Time `json:"lastUpdateTime"`
}

type ReplicationSpec struct {
	Project           string              `json:"project"`
	ReplicateNow      bool                `json:"replicateNow"`
	ReplicateDeletion bool                `json:"replicateDeletion"`
	Source            *ReplicationSource  `json:"source"`
	Target            *ReplicationTarget  `json:"target"`
	Trigger           *ReplicationTrigger `json:"trigger"`
}

type ReplicationSource struct {
	Name   string `json:"name"`
	Alias  string `json:"alias"`
	Domain string `json:"domain"`
}

type ReplicationTarget struct {
	Name   string `json:"name"`
	Alias  string `json:"alias"`
	Domain string `json:"domain"`
}

type ReplicationTrigger struct {
	Kind          string         `json:"type"` // 注意：这个地方的 json 注释和 变量名不一样
	ScheduleParam *ScheduleParam `json:"scheduleParam"`
}

// ScheduleParam defines the parameters used by schedule trigger
type ScheduleParam struct {
	Type    string `json:"type"`    //daily or weekly
	Weekday int8   `json:"weekday"` //Optional, only used when type is 'weekly'
	Offtime int64  `json:"offtime"` //The time offset with the UTC 00:00 in seconds
}

type ReplicationStatus struct {
	IsSyncing           bool      `json:"isSyncing"`
	LastReplicationTime time.Time `json:"lastReplicationTime"`
}

// =================================================================================================

type CreateReplicationReq struct {
	Metadata *ReplicationMetadata `json:"metadata"`
	Spec     *ReplicationSpec     `json:"spec"`
}

// =================================================================================================

type UpdateReplicationReq struct {
	Metadata *ReplicationMetadata `json:"metadata"`
	Spec     *ReplicationSpec     `json:"spec"`
}

// =================================================================================================

type TriggerReplicationReq struct {
	Action string `json:"action"`
}
