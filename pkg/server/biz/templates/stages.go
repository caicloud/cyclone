package templates

import (
	"io/ioutil"
	"os"

	"github.com/caicloud/nirvana/log"
	"gopkg.in/yaml.v2"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/biz/utils"
)

// StageTemplatesLoader loads stage templates from manifest files.
type StageTemplatesLoader struct {
	// TemplatesDir is path of the folder holding templates manifest files
	TemplatesDir string
}

// LoadStageTemplates loads templates from the manifest files.
// scene - Scene of the Workflow, for example, cicd. If provided, only templates
// for that scene would be loaded, otherwise all templates will be loaded.
func (l *StageTemplatesLoader) LoadStageTemplates(scene string) ([]*v1alpha1.Stage, error) {
	templatesPath := l.TemplatesDir
	if l.TemplatesDir == "" {
		templatesPath = "/root/templates"
	}
	if scene != "" {
		templatesPath = templatesPath + string(os.PathSeparator) + scene
	}

	files, err := utils.ListAllFiles(templatesPath)
	if err != nil {
		return nil, err
	}

	var results []*v1alpha1.Stage
	for _, f := range files {
		data, err := ioutil.ReadFile(f)
		if err != nil {
			log.Errorf("Load templates %s failed: error", f, err)
			continue
		}

		templates := make([]*v1alpha1.Stage, 0)
		err = yaml.Unmarshal(data, &templates)
		if err != nil {
			log.Errorf("Unmarshal yaml template file %s error: %v", f, err)
			continue
		}

		results = append(results, templates...)
	}

	return results, nil
}
