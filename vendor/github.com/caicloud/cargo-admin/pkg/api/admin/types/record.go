package types

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

type Record struct {
	Id           bson.ObjectId       `json:"id"`
	Replication  *Replication        `json:"replication"`
	Trigger      *ReplicationTrigger `json:"trigger"`
	SuccessCount int64               `json:"successCount"`
	FailedCount  int64               `json:"failedCount"`
	Status       string              `json:"status"`
	Reason       string              `json:"reason"`
	StartTime    time.Time           `json:"startTime"`
	EndTime      time.Time           `json:"endTime"`
}

// =================================================================================================

type RecordImage struct {
	Image     string `json:"image"`
	Operation string `json:"operation"`
	Status    string `json:"status"`
}
