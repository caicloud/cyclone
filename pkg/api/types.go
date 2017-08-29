/*
Copyright 2017 caicloud authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package api

import "time"

// Project represents a group to manage a set of related applications. It maybe a real project, which contains several or many applications.
type Project struct {
	ID          string    `bson:"_id,omitempty" json:"id,omitempty" description:"id of the project"`
	Name        string    `bson:"name,omitempty" json:"name,omitempty" description:"name of the project, should be unique"`
	Description string    `bson:"description,omitempty" json:"description,omitempty" description:"description of the project"`
	Owner       string    `bson:"owner,omitempty" json:"owner,omitempty" description:"owner of the project"`
	CreatedTime time.Time `bson:"createdTime,omitempty" json:"createdTime,omitempty" description:"created time of the project"`
	UpdatedTime time.Time `bson:"updatedTime,omitempty" json:"updatedTime,omitempty" description:"updated time of the project"`
}

// PipelineRecord is all information about some pipeline building history.
type PipelineRecord struct {
	ID          string       `bson:"_id,omitempty" json:"id,omitempty" description:"id of the pipelineRecord"`
	PipelineID  string       `bson:"pipelineID,omitempty" json:"pipelineID,omitempty" description:"id of the related pipeline which the pipelineRecord belongs to"`
	VersionID   string       `bson:"versionID,omitempty" json:"versionID,omitempty" description:"id of the related version which the pipelineRecord belongs to"`
	Trigger     string       `bson:"trigger,omitempty" json:"trigger,omitempty" description:"trigger of the pipelineRecord"`
	StageStatus *StageStatus `bson:"stageStatus,omitempty" json:"stageStatus,omitempty" description:"status of the latest stage"`
	Status      Status       `bson:"status,omitempty" json:"status,omitempty" description:"status of the pipelineRecord"`
	StartTime   time.Time    `bson:"startTime,omitempty" json:"startTime,omitempty" description:"startTime of the pipelineRecord"`
	EndTime     time.Time    `bson:"endTime,omitempty" json:"endTime,omitempty" description:"endTime of the pipelineRecord"`
}

// Status can be the status of some pipelineRecord or some stage
type Status string

const (
	// Pending means Cyclone is preparing to run this stage.
	Pending Status = "Pending"
	// Running means Cyclone is running with this stage.
	Running Status = "Running"
	// Success means Cyclone had run with this stage and result in success.
	Success Status = "Success"
	// Failure means Cyclone had run with this stage and result in failure.
	Failure Status = "Failed"
	// Abort means the stage was aborted by some reason, and we can get the reason from the log.
	Abort Status = "Aborted"
)

// TODO The status of every stage may be different.
// StageStatus shows the information of every stage.
type StageStatus struct {
	CodeCheckout    *GeneralStageStatus `bson:"codeCheckout,omitempty" json:"codeCheckout,omitempty" description:"code checkout stage"`
	UnitTest        *GeneralStageStatus `bson:"unitTest,omitempty" json:"unitTest,omitempty" description:"unit test stage"`
	CodeScan        *GeneralStageStatus `bson:"codeScan,omitempty" json:"codeScan,omitempty" description:"code scan stage"`
	Package         *GeneralStageStatus `bson:"package,omitempty" json:"package,omitempty" description:"package stage"`
	ImageBuild      *GeneralStageStatus `bson:"imageBuild,omitempty" json:"imageBuild,omitempty" description:"image build stage"`
	IntegrationTest *GeneralStageStatus `bson:"integrationTest,omitempty" json:"integrationTest,omitempty" description:"integration test stage"`
	ImageRelease    *GeneralStageStatus `bson:"imageRelease,omitempty" json:"imageRelease,omitempty" description:"image release stage"`
}

// GeneralStageStatus show the information of some stage.
type GeneralStageStatus struct {
	Status    Status    `bson:"status,omitempty" json:"status,omitempty" description:"status of the stage"`
	StartTime time.Time `bson:"startTime,omitempty" json:"startTime,omitempty" description:"startTime of the stage"`
	EndTime   time.Time `bson:"endTime,omitempty" json:"endTime,omitempty" description:"endTime of the stage"`
}
