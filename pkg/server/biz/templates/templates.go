package templates

import (
	"os"

	"github.com/caicloud/nirvana/log"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/caicloud/cyclone/pkg/k8s/clientset"
)

// InitStageTemplates loads and creates stage templates for the given scene.
// scene - Workflow scene, for example, 'cicd', empty value indicates all scenes.
func InitStageTemplates(client clientset.Interface, systemNamespace, scene string) {
	// Load all stage templates. Template files path is given by environment variable
	// TEMPLATES_PATH, if not set, use default one "/root/templates"
	loader := &StageTemplatesLoader{TemplatesDir: os.Getenv(TemplatesPathEnvName)}
	stages, err := loader.LoadStageTemplates(scene)
	if err != nil {
		log.Errorf("Load stage templates error: %v", err)
		return
	}
	log.Infof("%d stage templates loaded.", len(stages))

	// Create all stage templates
	for _, stg := range stages {
		_, err := client.CycloneV1alpha1().Stages(systemNamespace).Create(stg)
		if err != nil {
			if errors.IsAlreadyExists(err) {
				log.Infof("Stage template '%s' already exist, skip it.", stg.Name)
			} else {
				log.Errorf("Create stage template '%s' error: %v", stg.Name, err)
			}
		}
	}
}
