package v1alpha1

import (
	"fmt"
	"os"
	"strings"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	api "github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/server/config"
	"github.com/caicloud/cyclone/pkg/server/handler"
	"github.com/caicloud/cyclone/pkg/util/cerr"
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

func getIntegrationMetadata(tenant, name string) (meta_v1.ObjectMeta, error) {
	resource, err := handler.K8sClient.CoreV1().Secrets(common.TenantNamespace(tenant)).Get(common.IntegrationSecret(name), meta_v1.GetOptions{})
	if err != nil {
		return meta_v1.ObjectMeta{}, err
	}
	return resource.ObjectMeta, nil
}

func getProjectMetadata(tenant, name string) (meta_v1.ObjectMeta, error) {
	resource, err := handler.K8sClient.CycloneV1alpha1().Projects(common.TenantNamespace(tenant)).Get(name, meta_v1.GetOptions{})
	if err != nil {
		return meta_v1.ObjectMeta{}, err
	}
	return resource.ObjectMeta, nil
}

func getTenantMetadata(tenant, name string) (meta_v1.ObjectMeta, error) {
	resource, err := handler.K8sClient.CoreV1().Namespaces().Get(common.TenantNamespace(name), meta_v1.GetOptions{})
	if err != nil {
		return meta_v1.ObjectMeta{}, err
	}
	return resource.ObjectMeta, nil
}

// CreationModifier is used in creating cyclone resources. It's used to modify cyclone resources.
type CreationModifier func(tenant, project, wf string, object interface{}) error

// GenerateNameModifier is a modifier of create cyclone CRD resources.
// It will give the resource a name if it is empty.
func GenerateNameModifier(tenant, project, wf string, object interface{}) error {
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
	case *api.Tenant:
		meta = &obj.ObjectMeta
		getMetadata = getTenantMetadata
		resource = "tenants"
	case *api.Integration:
		meta = &obj.ObjectMeta
		getMetadata = getIntegrationMetadata
		resource = "integrations"
	case *v1alpha1.Project:
		meta = &obj.ObjectMeta
		getMetadata = getProjectMetadata
		resource = "projects"
	default:
		return fmt.Errorf("resource type not support")
	}

	if meta.Name == "" && (meta.Annotations == nil || meta.Annotations[common.AnnotationAlias] == "") {
		return fmt.Errorf("name and metadata.annotations[cyclone.io/alias] can not both be empty")
	}

	// Get name and alias, if alias not set, use name as alias
	name := meta.Name
	alias := ""
	if meta.Annotations != nil {
		alias = meta.Annotations[common.AnnotationAlias]
	}
	if alias == "" {
		alias = name
	}

	// Add alias annotation if not set
	if meta.Annotations == nil {
		meta.Annotations = make(map[string]string)
	}
	meta.Annotations[common.AnnotationAlias] = alias

	// If resource name set, check whether name conflict exists.
	if name != "" {
		_, err := getMetadata(tenant, name)
		if err == nil {
			return cerr.ErrorAlreadyExist.Error(fmt.Sprintf("%s %s", resource, name))
		}
		return nil
	}

	// If resource name not set, generate one from alias.
	if project != "" {
		name = slugify.Slugify(project+"-"+alias, false, -1)
	} else {
		name = slugify.Slugify(alias, false, -1)
	}

	_, err := getMetadata(tenant, name)
	if err == nil {
		name = slugify.Slugify(name, true, -1)
	}
	meta.Name = name

	return nil
}

// TenantModifier is a modifier of create cyclone tenant.
// This modifier will give some default value for tenant if it is nil.
func TenantModifier(tenant, project, wf string, object interface{}) error {
	t, ok := object.(*api.Tenant)
	if !ok {
		return fmt.Errorf("resource type not support")
	}

	if t.Spec.PersistentVolumeClaim.Size == "" {
		t.Spec.PersistentVolumeClaim.Size = config.Config.DefaultPVCConfig.Size
	}

	if t.Spec.ResourceQuota == nil {
		t.Spec.ResourceQuota = config.Config.WorkerNamespaceQuota
	}

	return nil
}

// InjectProjectLabelModifier is a modifier of create cyclone CRD resources.
// It will add project labels for the resource.
func InjectProjectLabelModifier(tenant, project, wf string, object interface{}) error {
	var meta *meta_v1.ObjectMeta
	switch obj := object.(type) {
	case *v1alpha1.Resource:
		meta = &obj.ObjectMeta
	case *v1alpha1.Stage:
		meta = &obj.ObjectMeta
	case *v1alpha1.Workflow:
		meta = &obj.ObjectMeta
	case *v1alpha1.WorkflowRun:
		meta = &obj.ObjectMeta
	case *v1alpha1.WorkflowTrigger:
		meta = &obj.ObjectMeta
	default:
		return fmt.Errorf("resource type not support")
	}

	// Add project label
	if meta.Labels == nil {
		meta.Labels = make(map[string]string)
	}
	meta.Labels[common.LabelProjectName] = project
	return nil
}

// WorkflowRunModifier is a modifier of create cyclone workflowrun resources.
// It will add workflow labels for workflowrun.
func WorkflowRunModifier(tenant, project, wf string, object interface{}) error {
	wfr, ok := object.(*v1alpha1.WorkflowRun)
	if !ok {
		return fmt.Errorf("resource type not support")
	}

	// Add workflow label
	if wfr.Labels == nil {
		wfr.Labels = make(map[string]string)
	}
	wfr.Labels[common.LabelWorkflowName] = wf
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

// LabelCustomizedTemplate gives a label to indicate that this is a customized template
func LabelCustomizedTemplate(stage *v1alpha1.Stage) {
	if stage.Labels == nil {
		stage.Labels = make(map[string]string)
	}
	stage.Labels[common.LabelBuiltin] = common.LabelFalseValue
	return
}

// LabelStageTemplate gives a label to indicate that this stage is a stage template
func LabelStageTemplate(stage *v1alpha1.Stage) {
	if stage.Labels == nil {
		stage.Labels = make(map[string]string)
	}
	stage.Labels[common.LabelStageTemplate] = common.LabelTrueValue
	return
}
