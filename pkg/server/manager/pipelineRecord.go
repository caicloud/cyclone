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

package manager

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	log "github.com/golang/glog"
	"github.com/gorilla/websocket"
	"gopkg.in/mgo.v2"

	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/store"
	fileutil "github.com/caicloud/cyclone/pkg/util/file"
	httperror "github.com/caicloud/cyclone/pkg/util/http/errors"
)

const (
	// logFileSuffix is the suffix of log file name.
	logFileSuffix = ".log"

	// logsFolderName is the folder name for logs files.
	logsFolderName = "logs"
)

// cycloneHome is the home folder for Cyclone.
var cycloneHome = "/var/lib/cyclone"

// PipelineRecordManager represents the interface to manage pipeline record.
type PipelineRecordManager interface {
	CreatePipelineRecord(pipelineRecord *api.PipelineRecord) (*api.PipelineRecord, error)
	GetPipelineRecord(pipelineRecordID string) (*api.PipelineRecord, error)
	ListPipelineRecords(projectName string, pipelineName string, queryParams api.QueryParams) ([]api.PipelineRecord, int, error)
	UpdatePipelineRecord(pipelineRecordID string, pipelineRecord *api.PipelineRecord) (*api.PipelineRecord, error)
	DeletePipelineRecord(pipelineRecordID string) error
	ClearPipelineRecordsOfPipeline(pipelineID string) error
	GetPipelineRecordLogs(pipelineRecordID, stage, task string) (string, error)
	GetPipelineRecordLogStream(pipelineRecordID, stage, task string, ws *websocket.Conn) error
	ReceivePipelineRecordLogStream(pipelineRecordID, stage, task string, ws *websocket.Conn) error
}

// pipelineRecordManager represents the manager for pipeline record.
type pipelineRecordManager struct {
	dataStore *store.DataStore
}

// NewPipelineRecordManager creates a pipeline record manager.
func NewPipelineRecordManager(dataStore *store.DataStore) (PipelineRecordManager, error) {
	if dataStore == nil {
		return nil, fmt.Errorf("Fail to new pipeline record manager as data store is nil")
	}

	return &pipelineRecordManager{dataStore}, nil
}

// CreatePipelineRecord triggers the pipeline to create a pipeline record.
func (m *pipelineRecordManager) CreatePipelineRecord(pipelineRecord *api.PipelineRecord) (*api.PipelineRecord, error) {
	pipeline, err := m.dataStore.FindPipelineByID(pipelineRecord.PipelineID)
	if err != nil {
		return nil, err
	}

	project, err := m.dataStore.FindProjectByID(pipeline.ProjectID)
	if err != nil {
		return nil, err
	}

	createdPipelineRecord, err := m.dataStore.CreatePipelineRecord(pipelineRecord)
	if err != nil {
		return nil, err
	}

	// Create the build outputs for pipelie record, such logs, etc.
	logsFolder := strings.Join([]string{cycloneHome, pipeline.ProjectID, pipeline.ID, createdPipelineRecord.ID, logsFolderName}, string(os.PathSeparator))
	if !fileutil.DirExists(logsFolder) {
		if err := os.MkdirAll(logsFolder, os.ModePerm); err != nil {
			log.Errorf("fail to make the folder %s as %s", logsFolder, err.Error())
			return nil, err
		}
	}

	// Create the event for this pipeline record.
	event := &api.Event{
		ID:             pipelineRecord.ID,
		Project:        project,
		Pipeline:       pipeline,
		PipelineRecord: pipelineRecord,
	}

	if _, err := m.dataStore.CreateEvent(event); err != nil {
		log.Errorf("fail to create the event %s as %s", event.ID, err.Error())
		return nil, err
	}

	return createdPipelineRecord, nil
}

// GetPipelineRecord gets the pipeline record by id.
func (m *pipelineRecordManager) GetPipelineRecord(pipelineRecordID string) (*api.PipelineRecord, error) {
	return m.dataStore.FindPipelineRecordByID(pipelineRecordID)
}

// ListPipelineRecords finds the pipeline records by pipeline id.
func (m *pipelineRecordManager) ListPipelineRecords(projectName string, pipelineName string, queryParams api.QueryParams) ([]api.PipelineRecord, int, error) {
	project, err := m.dataStore.FindProjectByName(projectName)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, 0, httperror.ErrorContentNotFound.Format(projectName)
		}
		return nil, 0, err
	}

	pipeline, err := m.dataStore.FindPipelineByName(project.ID, pipelineName)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, 0, httperror.ErrorContentNotFound.Format(pipelineName)
		}
		return nil, 0, err
	}

	return m.dataStore.FindRecordsWithPaginationByPipelineID(pipeline.ID, queryParams.Filter, queryParams.Start, queryParams.Limit)
}

// UpdatePipelineRecord updates pipeline record by id.
func (m *pipelineRecordManager) UpdatePipelineRecord(pipelineRecordID string, newPipelineRecord *api.PipelineRecord) (*api.PipelineRecord, error) {
	pipelineRecord, err := m.dataStore.FindPipelineRecordByID(pipelineRecordID)
	if err != nil {
		return nil, err
	}

	// Update the properties of the pipeline record.
	if newPipelineRecord.Status != "" {
		pipelineRecord.Status = newPipelineRecord.Status
	}
	if newPipelineRecord.StageStatus != nil {
		pipelineRecord.StageStatus = newPipelineRecord.StageStatus
	}

	if err = m.dataStore.UpdatePipelineRecord(pipelineRecord); err != nil {
		return nil, err
	}

	return pipelineRecord, nil
}

// DeletePipelineRecord deletes the pipeline record by id.
func (m *pipelineRecordManager) DeletePipelineRecord(pipelineRecordID string) error {
	return m.dataStore.DeletePipelineRecordByID(pipelineRecordID)
}

// ClearPipelineRecordsOfPipeline deletes all the pipeline records of one pipeline by pipeline id.
func (m *pipelineRecordManager) ClearPipelineRecordsOfPipeline(pipelineID string) error {
	ds := m.dataStore

	pipeline, err := ds.FindPipelineByID(pipelineID)
	if err != nil {
		return err
	}

	// Delete the records related to this pipeline.
	records, _, err := ds.FindPipelineRecordsByPipelineID(pipelineID, api.QueryParams{})
	if err != nil {
		return err
	}

	for _, record := range records {
		if err := ds.DeletePipelineByID(record.ID); err != nil {
			return fmt.Errorf("Fail to delete the record %s for pipeline %s as %s", record.ID, pipeline.Name, err.Error())
		}
	}

	return nil
}

// GetPipelineRecordLogs gets the pipeline record logs by id.
func (m *pipelineRecordManager) GetPipelineRecordLogs(pipelineRecordID, stage, task string) (string, error) {
	// Check the existence of record folder.
	logsFolder, err := m.getRecordFolder(pipelineRecordID)
	if err != nil {
		return "", err
	}

	if !fileutil.DirExists(logsFolder) {
		return "", fmt.Errorf("logs folder %s does not exist", logsFolder)
	}

	logFilePath, err := m.getLogFilePath(pipelineRecordID, stage, task)
	if err != nil {
		return "", err
	}

	// Check the existence of the log file for this stage. If does not exist, return error when pipeline record is success,
	// otherwise directly return the got logs as pipeline record is failed or aborted.
	if !fileutil.FileExists(logFilePath) {
		log.Errorf("log file %s does not exist", logFilePath)
		return "", fmt.Errorf("log file for stage %s does not exist", stage)
	}

	// TODO (robin) Read the whole file, need to consider the memory consumption when the log file is too huge.
	log, err := ioutil.ReadFile(logFilePath)
	if err != nil {
		return "", err
	}

	return string(log), nil
}

// GetPipelineRecordLogStream watches the log files and sends the content to the log stream.
func (m *pipelineRecordManager) GetPipelineRecordLogStream(pipelineRecordID, stage, task string, ws *websocket.Conn) error {
	logFilePath, err := m.getLogFilePath(pipelineRecordID, stage, task)
	if err != nil {
		return err
	}

	if !fileutil.FileExists(logFilePath) {
		return fmt.Errorf("log file %s already exists", logFilePath)
	}

	file, err := os.Open(logFilePath)
	if err != nil {
		log.Errorf("fail to open the log file %s as %s", logFilePath, err.Error())
		return err
	}
	defer file.Close()

	buf := bufio.NewReader(file)
	for {
		line, err := buf.ReadBytes('\n')
		if err == io.EOF {
			continue
		}

		if err != nil {
			ws.WriteMessage(websocket.CloseMessage, []byte("Interval error happens, TERMINATE"))
			break
		}
		ws.WriteMessage(websocket.TextMessage, line)
	}

	return nil
}

// ReceivePipelineRecordLogStream receives the log stream for one stage of the pipeline record, and stores it into log files.
func (m *pipelineRecordManager) ReceivePipelineRecordLogStream(pipelineRecordID, stage, task string, ws *websocket.Conn) error {
	logFilePath, err := m.getLogFilePath(pipelineRecordID, stage, task)
	if err != nil {
		return err
	}

	if fileutil.FileExists(logFilePath) {
		return fmt.Errorf("log file %s already exists", logFilePath)
	}

	file, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Errorf("fail to open the log file %s as %s", logFilePath, err.Error())
		return err
	}
	defer file.Close()

	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			if !websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure) {
				return nil
			}
			return err
		}
		_, err = file.Write(message)
		if err != nil {
			return err
		}
	}
}

// getLogFilePath gets the log file path for one stage of the pipeline record.
func (m *pipelineRecordManager) getLogFilePath(pipelineRecordID, stage, task string) (string, error) {
	if stage == "" {
		return "", fmt.Errorf("the stage can not be empty")
	}

	recordFolder, err := m.getRecordFolder(pipelineRecordID)
	if err != nil {
		return "", err
	}

	logFile := stage + logFileSuffix
	if task != "" {
		logFile = stage + "-" + task + logFileSuffix
	}

	logFilePath := strings.Join([]string{recordFolder, logsFolderName, logFile}, string(os.PathSeparator))
	return logFilePath, nil
}

// getRecordFolder gets the folder path for the pipeline record.
func (m *pipelineRecordManager) getRecordFolder(pipelineRecordID string) (string, error) {
	record, err := m.GetPipelineRecord(pipelineRecordID)
	if err != nil {
		return "", err
	}

	pipeline, err := m.dataStore.FindPipelineByID(record.PipelineID)
	if err != nil {
		return "", err
	}

	recordFolder := strings.Join([]string{cycloneHome, pipeline.ProjectID, pipeline.ID, pipelineRecordID}, string(os.PathSeparator))

	return recordFolder, nil
}
