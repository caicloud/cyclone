package integrate

import (
	"fmt"

	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/util/http/errors"
)

// ITGProvider is an interface for integration.
type ITGProvider interface {
	// CodeScan execute code analysis.
	CodeScan(url, token string, config *CodeScanConfig) (string, error)

	// SetCodeScanStatus sets status for CodeScanStageStatus.
	SetCodeScanStatus(url, token string, pid string, s *api.CodeScanStageStatus) error

	// CreateProject create a project.
	CreateProject(url, token string, projectKey, projectName string) error

	// SetQualityGate sets the project's quality gate.
	SetQualityGate(url, token string, projectKey string, gateId int) error

	// DeleteProject delete a project.
	DeleteProject(url, token string, projectKey string) error

	// Validate validate the token.
	Validate(url, token string) (bool, error)
}

// itgProviders represents the set of integration providers.
var itgProviders map[api.IntegrationType]ITGProvider

func init() {
	itgProviders = make(map[api.IntegrationType]ITGProvider)
}

// Register registers integration providers.
func Register(itype api.IntegrationType, p ITGProvider) error {
	if _, ok := itgProviders[itype]; ok {
		return fmt.Errorf("provider %s already exists.", itype)
	}

	itgProviders[itype] = p
	return nil
}

// GetProvider gets the integration provider by the type.
func GetProvider(itype api.IntegrationType) (ITGProvider, error) {
	provider, ok := itgProviders[itype]
	if !ok {
		return nil, fmt.Errorf("unsupported integration type %s", itype)
	}

	return provider, nil
}

// ScanSonarQubeConfig represents config of sonarqube-type code scan.
type CodeScanConfig struct {
	SourcePath    string   `bson:"sourcePath,omitempty" json:"sourcePath,omitempty"`
	EncodingStyle string   `bson:"encodingStyle,omitempty" json:"encodingStyle,omitempty"`
	Language      string   `bson:"language,omitempty" json:"language,omitempty"`
	Threshold     int      `bson:"threshold,omitempty" json:"threshold,omitempty"`
	ExtensionAgrs []string `bson:"extensionArgs,omitempty" json:"extensionArgs,omitempty"`
	ProjectName   string   `bson:"projectName,omitempty" json:"projectName,omitempty"`
	ProjectKey    string   `bson:"projectKey,omitempty" json:"projectKey,omitempty"`
}

// ScanCode execute code analysis.
func ScanCode(itype api.IntegrationType, url, token string, config *CodeScanConfig) (string, error) {
	p, err := GetProvider(itype)
	if err != nil {
		return "", err
	}

	return p.CodeScan(url, token, config)
}

// SetCodeScanStatus set status for CodeScanStageStatus, and the CodeScanStageStatus 's' must not be nil.
func SetCodeScanStatus(itype api.IntegrationType, url, token string, projectID string, s *api.CodeScanStageStatus) error {
	p, err := GetProvider(itype)
	if err != nil {
		return err
	}

	if s == nil {
		return fmt.Errorf("CodeScanStageStatus 's' can not be nil.")
	}

	return p.SetCodeScanStatus(url, token, projectID, s)
}

// CreateProject create a project.
func CreateProject(itype api.IntegrationType, url, token string, projectKey, projectName string) error {
	p, err := GetProvider(itype)
	if err != nil {
		return err
	}
	return p.CreateProject(url, token, projectKey, projectName)
}

// SetQualityGate sets the project's quality gate.
func SetQualityGate(itype api.IntegrationType, url, token string, projectKey string, gateId int) error {
	p, err := GetProvider(itype)
	if err != nil {
		return err
	}
	return p.SetQualityGate(url, token, projectKey, gateId)
}

// DeleteProject delete a project.
func DeleteProject(itype api.IntegrationType, url, token string, projectKey string) error {
	p, err := GetProvider(itype)
	if err != nil {
		return err
	}
	return p.DeleteProject(url, token, projectKey)
}

// Validate validate the token.
func Validate(it *api.Integration) (bool, error) {
	p, err := GetProvider(it.Type)
	if err != nil {
		return false, err
	}

	var url, token string

	switch it.Type {
	case api.IntegrationTypeSonar:
		if it.SonarQube == nil {
			return false, fmt.Errorf("integration type is SonarQube, so SonarQube info can not be empty")
		}
		url = it.SonarQube.Address
		token = it.SonarQube.Token

	default:
		return false, errors.ErrorNotImplemented.Error(
			fmt.Sprintf("Validate token for %s type integration", it.Type))
	}

	return p.Validate(url, token)
}
