package v1alpha1

import (
	"fmt"
	"os"
	"strings"

	"github.com/caicloud/nirvana/log"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/server/handler"
	"github.com/caicloud/cyclone/pkg/util/slugify"
)

const (
	// cycloneHome is the home folder for Cyclone.
	cycloneHome = "/var/lib/cyclone"

	// logsFolderName is the folder name for logs files.
	logsFolderName = "logs"
)

func getLogFilePath(workflowrun, stage, container, namespace string) (string, error) {
	if workflowrun == "" || stage == "" || container == "" {
		return "", fmt.Errorf("workflowrun/stage/container/namespace can not be empty")
	}

	rf, _ := getLogFolder(workflowrun, stage, namespace)
	return strings.Join([]string{rf, container}, string(os.PathSeparator)), nil
}

func getLogFolder(workflowrun, stage, namespace string) (string, error) {
	if workflowrun == "" || stage == "" || namespace == "" {
		return "", fmt.Errorf("workflowrun/stage/namespace can not be empty")
	}
	return strings.Join([]string{cycloneHome, namespace, workflowrun, stage, logsFolderName}, string(os.PathSeparator)), nil
}

// GetMetadata gets metadata of a type of k8s resources
type GetMetadata func(string, string) (meta_v1.ObjectMeta, error)

func getResourceMetadata(tenant, name string) (meta_v1.ObjectMeta, error) {
	resource, err := handler.K8sClient.CycloneV1alpha1().Resources(common.TenantNamespace(tenant)).Get(name, meta_v1.GetOptions{})
	if err != nil {
		return meta_v1.ObjectMeta{}, err
	}
	return resource.ObjectMeta, nil
}

func getStageMetadata(tenant, name string) (meta_v1.ObjectMeta, error) {
	resource, err := handler.K8sClient.CycloneV1alpha1().Stages(common.TenantNamespace(tenant)).Get(name, meta_v1.GetOptions{})
	if err != nil {
		return meta_v1.ObjectMeta{}, err
	}
	return resource.ObjectMeta, nil
}

func getWfMetadata(tenant, name string) (meta_v1.ObjectMeta, error) {
	resource, err := handler.K8sClient.CycloneV1alpha1().Workflows(common.TenantNamespace(tenant)).Get(name, meta_v1.GetOptions{})
	if err != nil {
		return meta_v1.ObjectMeta{}, err
	}
	return resource.ObjectMeta, nil
}

func getWfrMetadata(tenant, name string) (meta_v1.ObjectMeta, error) {
	resource, err := handler.K8sClient.CycloneV1alpha1().WorkflowRuns(common.TenantNamespace(tenant)).Get(name, meta_v1.GetOptions{})
	if err != nil {
		return meta_v1.ObjectMeta{}, err
	}
	return resource.ObjectMeta, nil
}

func getWftMetadata(tenant, name string) (meta_v1.ObjectMeta, error) {
	resource, err := handler.K8sClient.CycloneV1alpha1().WorkflowTriggers(common.TenantNamespace(tenant)).Get(name, meta_v1.GetOptions{})
	if err != nil {
		return meta_v1.ObjectMeta{}, err
	}
	return resource.ObjectMeta, nil
}

// CreatePrelude is the prelude of create cyclone CRD resources.
// It will give the resource a name if it is empty. and will add project labels for the project.
func CreatePrelude(project, tenant string, object interface{}) error {
	var getMetadata GetMetadata
	var meta *meta_v1.ObjectMeta
	var resource string
	switch obj := object.(type) {
	case *v1alpha1.Resource:
		meta = &obj.ObjectMeta
		getMetadata = getResourceMetadata
		resource = "resources"
	case *v1alpha1.Stage:
		meta = &obj.ObjectMeta
		getMetadata = getResourceMetadata
		resource = "stages"
	case *v1alpha1.Workflow:
		meta = &obj.ObjectMeta
		getMetadata = getWfMetadata
		resource = "workflows"
	case *v1alpha1.WorkflowRun:
		meta = &obj.ObjectMeta
		getMetadata = getWfrMetadata
		resource = "workflowruns"
	case *v1alpha1.WorkflowTrigger:
		meta = &obj.ObjectMeta
		getMetadata = getWftMetadata
		resource = "workflowtriggers"
	default:
		return fmt.Errorf("resource type not support")
	}

	if meta.Name == "" && (meta.Annotations == nil || meta.Annotations[common.AnnotationAlias] == "") {
		return fmt.Errorf("name and metadata.annotations[cyclone.io/alias] can not both be empty")
	}
	name := meta.Name
	alias := ""
	if meta.Annotations != nil {
		alias = meta.Annotations[common.AnnotationAlias]
	}

	nameEmpty := false
	if name == "" && alias != "" {
		name = slugify.Slugify(project+"-"+alias, false, -1)
		nameEmpty = true
	}

	if old, err := getMetadata(tenant, name); err == nil {
		log.Errorf("name %s conflict, alias:%s, exist alias:%s",
			name, alias, old.Annotations[common.AnnotationAlias])
		if nameEmpty {
			name = slugify.Slugify(name, true, -1)
		} else {
			return errors.NewAlreadyExists(schema.GroupResource{Group: v1alpha1.APIVersion, Resource: resource}, name)
		}
	}

	meta.Name = name
	if meta.Annotations == nil {
		meta.Annotations = make(map[string]string)
	}
	if alias == "" {
		meta.Annotations[common.AnnotationAlias] = name
	} else {
		meta.Annotations[common.AnnotationAlias] = alias
	}

	// add project label
	if meta.Labels == nil {
		meta.Labels = make(map[string]string)
	}
	meta.Labels[common.LabelProject] = project
	return nil
}

// UpdateAnnotations updates alias and description annotations
func UpdateAnnotations(oldm, newm map[string]string) map[string]string {
	if oldm != nil {
		if newm == nil {
			newm = make(map[string]string)
		}
		newm[common.AnnotationAlias] = oldm[common.AnnotationAlias]
		newm[common.AnnotationDescription] = oldm[common.AnnotationDescription]
	}

	return newm
}
