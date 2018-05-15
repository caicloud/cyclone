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
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/zoumo/logdog"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/caicloud/cyclone/cmd/worker/options"
	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/event"
	"github.com/caicloud/cyclone/pkg/osutil"
	"github.com/caicloud/cyclone/pkg/scm"
	"github.com/caicloud/cyclone/pkg/store"
	httperror "github.com/caicloud/cyclone/pkg/util/http/errors"
)

// PipelineManager represents the interface to manage pipeline.
type PipelineManager interface {
	CreatePipeline(projectName string, pipeline *api.Pipeline) (*api.Pipeline, error)
	GetPipeline(projectName string, pipelineName string) (*api.Pipeline, error)
	GetPipelineByID(id string) (*api.Pipeline, error)
	ListPipelines(projectName string, queryParams api.QueryParams, recentCount, recentSuccessCount, recentFailedCount int) ([]api.Pipeline, int, error)
	UpdatePipeline(projectName string, pipelineName string, newPipeline *api.Pipeline) (*api.Pipeline, error)
	DeletePipeline(projectName string, pipelineName string) error
	ClearPipelinesOfProject(projectName string) error
	GetStatistics(projectName, pipelineName string, start, end time.Time) (*api.PipelineStatusStats, error)
}

// pipelineManager represents the manager for pipeline.
type pipelineManager struct {
	dataStore             *store.DataStore
	pipelineRecordManager PipelineRecordManager

	// TODO (robin) Move event manager to pipeline record manager.
	eventManager event.EventManager

	// callbackURL represents callback url for SCM webhooks.
	// It's mainly for SCM webhooks to trigger pipelines when SCM server can not directly connect Cyclone server.
	// callbackURL string
}

// NewPipelineManager creates a pipeline manager.
func NewPipelineManager(dataStore *store.DataStore, pipelineRecordManager PipelineRecordManager) (PipelineManager, error) {
	if dataStore == nil {
		return nil, fmt.Errorf("Fail to new pipeline manager as data store is nil")
	}

	if pipelineRecordManager == nil {
		return nil, fmt.Errorf("Fail to new pipeline manager as pipeline record is nil")
	}

	eventManager := event.NewEventManager(dataStore)

	return &pipelineManager{dataStore, pipelineRecordManager, eventManager}, nil
}

// CreatePipeline creates a pipeline.
func (m *pipelineManager) CreatePipeline(projectName string, pipeline *api.Pipeline) (*api.Pipeline, error) {
	// Check the existence of the project and pipeline.
	if _, err := m.GetPipeline(projectName, pipeline.Name); err == nil {
		return nil, httperror.ErrorAlreadyExist.Format(pipeline.Name)
	}

	scmConfig, err := m.GetSCMConfigFromProject(projectName)
	if err != nil {
		return nil, err
	}

	provider, err := scm.GetSCMProvider(scmConfig.Type)
	if err != nil {
		return nil, httperror.ErrorInternalTypeError.Format("Can not get the SCM provider")
	}

	// Create SCM webhook if enable SCM trigger.
	var webHook *scm.WebHook
	var gitSource *api.GitSource
	if pipeline.AutoTrigger != nil && pipeline.AutoTrigger.SCMTrigger != nil {
		gitSource, err = api.GetGitSource(pipeline.Build.Stages.CodeCheckout.MainRepo)
		if err != nil {
			return nil, err
		}

		pipeline.ID = bson.NewObjectId().Hex()
		webHook = &scm.WebHook{
			Url:    generateWebhookURL(scmConfig.Type, pipeline.ID),
			Events: collectSCMEvents(pipeline.AutoTrigger.SCMTrigger),
		}
		if err := provider.CreateWebHook(scmConfig, gitSource.Url, webHook); err != nil {
			return nil, err
		}
		pipeline.AutoTrigger.SCMTrigger.Webhook = webHook.Url
	}

	// Remove the webhook if there is error.
	defer func() {
		if err != nil && gitSource != nil && webHook != nil {
			if err = provider.DeleteWebHook(scmConfig, gitSource.Url, webHook.Url); err != nil {
				logdog.Errorf("Fail to delete the pipeline %s", pipeline.Name)
			}
		}
	}()

	createdPipeline, err := m.dataStore.CreatePipeline(pipeline)
	if err != nil {
		return nil, err
	}

	return createdPipeline, nil
}

// GetPipeline gets the pipeline by name in one project.
func (m *pipelineManager) GetPipeline(projectName string, pipelineName string) (*api.Pipeline, error) {
	project, err := m.dataStore.FindProjectByName(projectName)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, httperror.ErrorContentNotFound.Format(projectName)
		}

		return nil, err
	}

	pipeline, err := m.dataStore.FindPipelineByName(project.ID, pipelineName)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, httperror.ErrorContentNotFound.Format(pipelineName)
		}

		return nil, err
	}

	return pipeline, nil
}

// GetPipelineByID gets the pipeline by id.
func (m *pipelineManager) GetPipelineByID(id string) (*api.Pipeline, error) {
	pipeline, err := m.dataStore.FindPipelineByID(id)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, httperror.ErrorContentNotFound.Format("pipeline with id %s", id)
		}

		return nil, err
	}

	return pipeline, nil
}

// ListPipelines lists all pipelines in one project.
func (m *pipelineManager) ListPipelines(projectName string, queryParams api.QueryParams,
	recentCount, recentSuccessCount, recentFailedCount int) ([]api.Pipeline, int, error) {
	ds := m.dataStore

	project, err := ds.FindProjectByName(projectName)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, 0, httperror.ErrorContentNotFound.Format(projectName)
		}
		return nil, 0, err
	}

	pipelines, total, err := ds.FindPipelinesByProjectID(project.ID, queryParams)
	if err != nil {
		return nil, 0, err
	}

	if recentCount <= 0 && recentSuccessCount <= 0 && recentFailedCount <= 0 {
		return pipelines, total, nil
	}

	wg := sync.WaitGroup{}
	for i, _ := range pipelines {
		wg.Add(1)

		go func(pipeline *api.Pipeline) {
			defer wg.Done()

			if recentCount > 0 {
				recentRecords, _, err := ds.FindRecentRecordsByPipelineID(pipeline.ID, nil, recentCount)
				if err != nil {
					logdog.Error(err)
				} else {
					pipeline.RecentRecords = recentRecords
				}
			}

			if recentSuccessCount > 0 {
				filter := map[string]interface{}{
					"status": api.Success,
				}
				recentSuccessRecords, _, err := ds.FindRecentRecordsByPipelineID(pipeline.ID, filter, recentSuccessCount)
				if err != nil {
					logdog.Error(err)
				} else {
					pipeline.RecentSuccessRecords = recentSuccessRecords
				}
			}

			if recentFailedCount > 0 {
				filter := map[string]interface{}{
					"status": api.Failed,
				}
				recentFailedRecords, _, err := ds.FindRecentRecordsByPipelineID(pipeline.ID, filter, recentFailedCount)
				if err != nil {
					logdog.Error(err)
				} else {
					pipeline.RecentFailedRecords = recentFailedRecords
				}
			}
		}(&pipelines[i])
	}
	wg.Wait()

	return pipelines, total, nil
}

// UpdatePipeline updates the pipeline by name in one project.
func (m *pipelineManager) UpdatePipeline(projectName string, pipelineName string, newPipeline *api.Pipeline) (*api.Pipeline, error) {
	pipeline, err := m.GetPipeline(projectName, pipelineName)
	if err != nil {
		return nil, err
	}

	scmConfig, err := m.GetSCMConfigFromProject(projectName)
	if err != nil {
		return nil, httperror.ErrorInternalTypeError.Format("Can not get the SCM config")
	}

	provider, err := scm.GetSCMProvider(scmConfig.Type)
	if err != nil {
		return nil, httperror.ErrorInternalTypeError.Format("Can not get the SCM provider")
	}

	// Remove the old webhook if exists.
	if pipeline.AutoTrigger != nil && pipeline.AutoTrigger.SCMTrigger != nil {
		scmTrigger := pipeline.AutoTrigger.SCMTrigger
		if scmTrigger.Webhook != "" {
			gitSource, err := api.GetGitSource(pipeline.Build.Stages.CodeCheckout.MainRepo)
			if err != nil {
				return nil, err
			}

			if err := provider.DeleteWebHook(scmConfig, gitSource.Url, scmTrigger.Webhook); err != nil {
				return nil, err
			}
		}
	}

	// Create the new webhook if necessary.
	if newPipeline.AutoTrigger != nil && newPipeline.AutoTrigger.SCMTrigger != nil {
		scmTrigger := newPipeline.AutoTrigger.SCMTrigger
		gitSource, err := api.GetGitSource(newPipeline.Build.Stages.CodeCheckout.MainRepo)
		if err != nil {
			return nil, err
		}

		webHook := &scm.WebHook{
			Url:    generateWebhookURL(scmConfig.Type, pipeline.ID),
			Events: collectSCMEvents(scmTrigger),
		}
		if err := provider.CreateWebHook(scmConfig, gitSource.Url, webHook); err != nil {
			logdog.Errorf("create webhook failed: %v", err)
			return nil, httperror.ErrorInternalTypeError.Format("Can not create webhook")
		}

		newPipeline.AutoTrigger.SCMTrigger.Webhook = webHook.Url
	}

	pipeline.AutoTrigger = newPipeline.AutoTrigger

	// Update the properties of the pipeline.
	// TODO (robin) Whether need a method for this merge?
	if len(newPipeline.Name) > 0 {
		pipeline.Name = newPipeline.Name
	}

	if len(newPipeline.Alias) > 0 {
		pipeline.Alias = newPipeline.Alias
	}

	if len(newPipeline.Description) > 0 {
		pipeline.Description = newPipeline.Description
	}

	if len(newPipeline.Owner) > 0 {
		pipeline.Owner = newPipeline.Owner
	}

	if newPipeline.Build != nil {
		pipeline.Build = newPipeline.Build
	}

	if err = m.dataStore.UpdatePipeline(pipeline); err != nil {
		return nil, err
	}

	return pipeline, nil
}

// DeletePipeline deletes the pipeline by name in one project.
func (m *pipelineManager) DeletePipeline(projectName string, pipelineName string) error {
	pipeline, err := m.GetPipeline(projectName, pipelineName)
	if err != nil {
		return err
	}

	scmConfig, err := m.GetSCMConfigFromProject(projectName)
	if err != nil {
		return err
	}

	return m.deletePipeline(scmConfig, pipeline)
}

// ClearPipelinesOfProject deletes all pipelines in one project.
func (m *pipelineManager) ClearPipelinesOfProject(projectName string) error {
	pipelines, _, err := m.ListPipelines(projectName, api.QueryParams{}, 0, 0, 0)
	if err != nil {
		return nil
	}

	scmConfig, err := m.GetSCMConfigFromProject(projectName)
	if err != nil {
		return err
	}
	for _, pipeline := range pipelines {
		if err := m.deletePipeline(scmConfig, &pipeline); err != nil {
			return err
		}
	}

	return nil
}

// deletePipeline deletes the pipeline.
func (m *pipelineManager) deletePipeline(scmConfig *api.SCMConfig, pipeline *api.Pipeline) error {
	ds := m.dataStore
	var err error

	// Delete the pipeline records of this pipeline.
	if err = m.pipelineRecordManager.ClearPipelineRecordsOfPipeline(pipeline.ID); err != nil {
		return fmt.Errorf("Fail to delete all pipeline records for pipeline %s as %s", pipeline.Name, err.Error())
	}

	if err = ds.DeletePipelineByID(pipeline.ID); err != nil {
		return fmt.Errorf("Fail to delete the pipeline %s as %s", pipeline.Name, err.Error())
	}

	if pipeline.AutoTrigger != nil && pipeline.AutoTrigger.SCMTrigger != nil {
		gitSource, err := api.GetGitSource(pipeline.Build.Stages.CodeCheckout.MainRepo)
		if err != nil {
			return err
		}

		provider, err := scm.GetSCMProvider(scmConfig.Type)
		if err != nil {
			return fmt.Errorf("Can not get provider for SCM %s", scmConfig.Type)
		}

		if err := provider.DeleteWebHook(scmConfig, gitSource.Url, pipeline.AutoTrigger.SCMTrigger.Webhook); err != nil {
			logdog.Errorf("Fail to delete webhook for pipeline %s", pipeline.Name)
		}
	}

	return nil
}

func generateWebhookURL(scmType api.SCMType, pipelineID string) string {
	callbackURL := osutil.GetStringEnv(options.CallbackURL, "http://127.0.0.1:7099/v1/pipelines")
	callbackURL = strings.TrimSuffix(callbackURL, "/")
	return fmt.Sprintf("%s/%s/%swebhook", callbackURL, pipelineID, strings.ToLower(string(scmType)))
}

func collectSCMEvents(scmTrigger *api.SCMTrigger) []scm.EventType {
	var events []scm.EventType
	if scmTrigger == nil {
		return events
	}

	if scmTrigger.PullRequest != nil {
		events = append(events, scm.PullRequestEventType)
	}
	if scmTrigger.PullRequestComment != nil {
		events = append(events, scm.PullRequestCommentEventType)
	}
	if scmTrigger.TagRelease != nil {
		events = append(events, scm.TagReleaseEventType)
	}
	if scmTrigger.Push != nil {
		events = append(events, scm.PushEventType)
	}

	return events
}

// GetSCMConfigFromProject
func (m *pipelineManager) GetSCMConfigFromProject(projectName string) (*api.SCMConfig, error) {
	// Get the SCM config from project.
	project, err := m.dataStore.FindProjectByName(projectName)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, httperror.ErrorContentNotFound.Format(projectName)
		}

		return nil, err
	}

	return project.SCM, nil
}

/// GetStatistics gets the statistic by pipeline name.
func (m *pipelineManager) GetStatistics(projectName, pipelineName string, start, end time.Time) (*api.PipelineStatusStats, error) {
	pipeline, err := m.GetPipeline(projectName, pipelineName)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, httperror.ErrorContentNotFound.Format(projectName)
		}

		return nil, err
	}

	// find all records ( start<={records}.startTime<end && {records}.pipelineID=pipeline.ID )
	records, _, err := m.dataStore.FindPipelineRecordsByStartTime(pipeline.ID, start, end)
	if err != nil {
		return nil, err
	}

	return transRecordsToStats(records, start, end)
}

func transRecordsToStats(records []api.PipelineRecord, start, end time.Time) (*api.PipelineStatusStats, error) {
	statistics := &api.PipelineStatusStats{
		Overview: api.StatsOverview{
			Total:        len(records),
			SuccessRatio: "0.00%",
		},
		Details: []*api.StatsDetail{},
	}

	initStatsDetails(statistics, start, end)

	for _, record := range records {
		for _, detail := range statistics.Details {
			if detail.Timestamp == formatTimeToDay(record.StartTime) {
				// set details status
				detail.StatsStatus = statsStatus(detail.StatsStatus, record.Status)
			}

		}

		// set overview status
		statistics.Overview.StatsStatus = statsStatus(statistics.Overview.StatsStatus, record.Status)
	}

	if statistics.Overview.Total != 0 {
		statistics.Overview.SuccessRatio = fmt.Sprintf("%.2f%%",
			float64(statistics.Overview.Success)/float64(statistics.Overview.Total)*100)
	}
	return statistics, nil
}

func formatTimeToDay(t time.Time) int64 {
	timestamp := t.Unix()
	return timestamp - (timestamp % 86400)
}

func statsStatus(s api.StatsStatus, recordStatus api.Status) api.StatsStatus {
	switch recordStatus {
	case api.Success:
		s.Success++
	case api.Failed:
		s.Failed++
	case api.Aborted:
		s.Aborted++
	default:
	}

	return s
}

func initStatsDetails(statistics *api.PipelineStatusStats, start, end time.Time) {
	for ; !start.After(end); start = start.Add(24 * time.Hour) {
		detail := &api.StatsDetail{
			Timestamp: formatTimeToDay(start),
		}
		statistics.Details = append(statistics.Details, detail)
	}

	// if last day not equal end day, append end day.
	endDay := formatTimeToDay(end)
	length := len(statistics.Details)
	if length > 0 {
		if statistics.Details[length-1].Timestamp != endDay {
			detail := &api.StatsDetail{
				Timestamp: endDay,
			}
			statistics.Details = append(statistics.Details, detail)
		}
	}
}
