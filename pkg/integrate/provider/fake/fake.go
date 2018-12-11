package fake

import (
	"github.com/caicloud/nirvana/log"

	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/integrate"
)

type FakeProvider struct {
}

func init() {
	if err := integrate.Register(api.IntegrationTypeSonar, new(FakeProvider)); err != nil {
		log.Error(err)
	}
}

// CodeScan execute code analysis.
func (f *FakeProvider) CodeScan(url, token string, config *integrate.CodeScanConfig) (string, error) {
	return "", nil
}

// SetCodeScanStatus sets status for CodeScanStageStatus.
func (f *FakeProvider) SetCodeScanStatus(url, token string, pid string, s *api.CodeScanStageStatus) error {
	return nil
}

// CreateProject create a project.
func (f *FakeProvider) CreateProject(url, token string, projectKey, projectName string) error {
	return nil
}

// SetQualityGate sets the project's quality gate.
func (f *FakeProvider) SetQualityGate(url, token string, projectKey string, gateId int) error {
	return nil
}

// DeleteProject delete a project.
func (f *FakeProvider) DeleteProject(url, token string, projectKey string) error {
	return nil
}

// Validate validate the token.
func (f *FakeProvider) Validate(url, token string) (bool, error) {
	return true, nil
}
