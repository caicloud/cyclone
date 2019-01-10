package handler

import (
	"github.com/caicloud/cyclone/pkg/server/manager"
	"github.com/caicloud/cyclone/pkg/store"
)

var (
	pipelineRecordManager manager.PipelineRecordManager
	pipelineManager       manager.PipelineManager
	eventManager          manager.EventManager
	cloudManager          manager.CloudManager
	projectManager        manager.ProjectManager
	integrationManager    manager.IntegrationManager

	ds *store.DataStore
)

func InitHandler(dataStore *store.DataStore, recordRotationThreshold int) (err error) {
	ds = dataStore
	// New pipeline record manager
	pipelineRecordManager, err = manager.NewPipelineRecordManager(dataStore, recordRotationThreshold)
	if err != nil {
		return err
	}

	// New pipeline manager
	pipelineManager, err = manager.NewPipelineManager(dataStore, pipelineRecordManager)
	if err != nil {
		return err
	}

	// New project manager
	projectManager, err = manager.NewProjectManager(dataStore, pipelineManager)
	if err != nil {
		return err
	}

	// New event manager
	eventManager, err = manager.NewEventManager(dataStore)
	if err != nil {
		return err
	}

	// New cloud manager
	cloudManager, err = manager.NewCloudManager(dataStore)
	if err != nil {
		return err
	}

	// New integration manager
	integrationManager, err = manager.NewIntegrationManager(dataStore)
	if err != nil {
		return err
	}

	return nil
}
