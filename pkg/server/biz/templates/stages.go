package templates

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/caicloud/nirvana/log"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/biz/utils"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const (
	// TemplatesPathEnvName is templates path environment name
	TemplatesPathEnvName = "TEMPLATES_PATH"

	// DefaultTemplatesPath is default path of stage templates
	DefaultTemplatesPath = "/root/templates"
)

// StageTemplatesLoader loads stage templates from manifest files.
type StageTemplatesLoader struct {
	// TemplatesDir is path of the folder holding templates manifest files.
	// If not set, default value DefaultTemplatesPath will be used.
	TemplatesDir string
}

// LoadStageTemplates loads templates from the manifest files.
// scene - Scene of the Workflow, for example, cicd. If provided, only templates
// for that scene would be loaded, otherwise all templates will be loaded.
func (l *StageTemplatesLoader) LoadStageTemplates(scene string) ([]*v1alpha1.Stage, error) {
	templatesPath := l.TemplatesDir
	if l.TemplatesDir == "" {
		templatesPath = DefaultTemplatesPath
	}
	if scene != "" {
		templatesPath = templatesPath + string(os.PathSeparator) + scene
	}
	log.Infof("Load stage templates from '%s'", templatesPath)

	files, err := utils.ListAllFiles(templatesPath)
	if err != nil {
		return nil, err
	}

	var results []*v1alpha1.Stage
	for _, f := range files {
		data, err := ioutil.ReadFile(f)
		if err != nil {
			log.Errorf("Load templates %s failed: %v", f, err)
			continue
		}

		templates := make([]*v1alpha1.Stage, 0)
		jsonData, err := yaml.ToJSON(data)
		if err != nil {
			log.Errorf("Convert template %s from YAML to JSON error: %v", f, err)
			continue
		}
		json.Unmarshal(jsonData, &templates)

		results = append(results, templates...)
	}

	return results, nil
}
