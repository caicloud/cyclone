package v1alpha1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"text/template"

	"github.com/caicloud/nirvana/log"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/meta"
	api "github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/biz/integration"
	"github.com/caicloud/cyclone/pkg/server/biz/scm"
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

	// queryFilterName represents filter by name.
	queryFilterName = "name"
	// queryFilterAlias represents filter by alias.
	queryFilterAlias = "alias"
	// queryFilterType represents filter by type.
	queryFilterType = "type"
)

func getLogFilePath(tenant, project, workflow, workflowrun, stage, container string) (string, error) {
	rf, err := getLogFolder(tenant, project, workflow, workflowrun)
	if err != nil {
		return "", err
	}
	return strings.Join([]string{rf, fmt.Sprintf("%s_%s", stage, container)}, string(os.PathSeparator)), nil
}

func getLogFolder(tenant, project, workflow, workflowrun string) (string, error) {
	if tenant == "" || project == "" || workflow == "" || workflowrun == "" {
		return "", fmt.Errorf("tenant/project/workflow/workflowrun can not be empty")
	}
	return strings.Join([]string{cycloneHome, tenant, project, workflow, workflowrun, logsFolderName}, string(os.PathSeparator)), nil
}

// deleteCollections deletes collections in the sub paths of a tenant in the pvc, collections including:
// - logs
// - artifacts (WIP)
// subpaths could be:
// - project
// - workflow
// - workflowrun
func deleteCollections(tenant string, subpaths ...string) error {
	if tenant == "" {
		return fmt.Errorf("tenant can not be empty")
	}

	// get collection folder
	folder := getCollectionFolder(tenant, subpaths...)

	// remove collection folder
	if err := os.RemoveAll(folder); err != nil {
		log.Errorf("remove folder %s error:%v", folder, err)
		return err
	}

	return nil
}

func getCollectionFolder(tenant string, subpaths ...string) string {
	paths := []string{cycloneHome, tenant}
	paths = append(paths, subpaths...)
	return strings.Join(paths, string(os.PathSeparator))
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
	resource, err := handler.K8sClient.CoreV1().Secrets(common.TenantNamespace(tenant)).Get(integration.GetSecretName(name), meta_v1.GetOptions{})
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

	// if tenant exist, use it directly.
	return resource.ObjectMeta, cerr.ErrorAlreadyExist.Error(fmt.Sprintf("%s %s", "namespaces", name))
}

// CreationModifier is used in creating cyclone resources. It's used to modify cyclone resources.
type CreationModifier func(tenant, project, wf string, object interface{}) error

// GenerateNameModifier is a modifier of create cyclone CRD resources.
// It will give the resource a name if it is empty.
func GenerateNameModifier(tenant, project, wf string, object interface{}) error {
	var getMetadata GetMetadata
	var objectMeta *meta_v1.ObjectMeta
	var resource string
	switch obj := object.(type) {
	case *v1alpha1.Resource:
		objectMeta = &obj.ObjectMeta
		getMetadata = getResourceMetadata
		resource = "resources"
	case *v1alpha1.Stage:
		objectMeta = &obj.ObjectMeta
		getMetadata = getStageMetadata
		resource = "stages"
	case *v1alpha1.Workflow:
		objectMeta = &obj.ObjectMeta
		getMetadata = getWfMetadata
		resource = "workflows"
	case *v1alpha1.WorkflowRun:
		objectMeta = &obj.ObjectMeta
		getMetadata = getWfrMetadata
		resource = "workflowruns"
	case *v1alpha1.WorkflowTrigger:
		objectMeta = &obj.ObjectMeta
		getMetadata = getWftMetadata
		resource = "workflowtriggers"
	case *api.Tenant:
		objectMeta = &obj.ObjectMeta
		getMetadata = getTenantMetadata
		resource = "tenants"
	case *api.Integration:
		objectMeta = &obj.ObjectMeta
		getMetadata = getIntegrationMetadata
		resource = "integrations"
	case *v1alpha1.Project:
		objectMeta = &obj.ObjectMeta
		getMetadata = getProjectMetadata
		resource = "projects"
	default:
		return fmt.Errorf("resource type not support")
	}

	if objectMeta.Name == "" && (objectMeta.Annotations == nil || objectMeta.Annotations[meta.AnnotationAlias] == "") {
		return fmt.Errorf("name and metadata.annotations[cyclone.dev/alias] can not both be empty")
	}

	// Get name and alias, if alias not set, use name as alias
	name := objectMeta.Name
	alias := ""
	if objectMeta.Annotations != nil {
		alias = objectMeta.Annotations[meta.AnnotationAlias]
	}
	if alias == "" {
		alias = name
	}

	// Add alias annotation if not set
	if objectMeta.Annotations == nil {
		objectMeta.Annotations = make(map[string]string)
	}
	objectMeta.Annotations[meta.AnnotationAlias] = alias

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
	objectMeta.Name = name

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
	var objectMeta *meta_v1.ObjectMeta
	switch obj := object.(type) {
	case *v1alpha1.Resource:
		objectMeta = &obj.ObjectMeta
	case *v1alpha1.Stage:
		objectMeta = &obj.ObjectMeta
	case *v1alpha1.Workflow:
		objectMeta = &obj.ObjectMeta
	case *v1alpha1.WorkflowRun:
		objectMeta = &obj.ObjectMeta
	case *v1alpha1.WorkflowTrigger:
		objectMeta = &obj.ObjectMeta
	default:
		return fmt.Errorf("resource type not support")
	}

	// Add project label
	if objectMeta.Labels == nil {
		objectMeta.Labels = make(map[string]string)
	}
	objectMeta.Labels[meta.LabelProjectName] = project
	return nil
}

// InjectProjectOwnerRefModifier is a modifier of creating cyclone CRD resources.
// It will add project owner reference for the resource.
func InjectProjectOwnerRefModifier(tenant, project, wf string, object interface{}) error {
	var objectMeta *meta_v1.ObjectMeta
	switch obj := object.(type) {
	case *v1alpha1.Resource:
		objectMeta = &obj.ObjectMeta
	case *v1alpha1.Stage:
		objectMeta = &obj.ObjectMeta
	case *v1alpha1.Workflow:
		objectMeta = &obj.ObjectMeta
	case *v1alpha1.WorkflowTrigger:
		objectMeta = &obj.ObjectMeta
	default:
		return fmt.Errorf("resource type not support")
	}

	projectMeta, err := getProjectMetadata(tenant, project)
	if err != nil {
		return err
	}
	// Add project OwnerReferences
	objectMeta.OwnerReferences = append(objectMeta.OwnerReferences, meta_v1.OwnerReference{
		APIVersion: v1alpha1.APIVersion,
		Kind:       reflect.TypeOf(v1alpha1.Project{}).Name(),
		Name:       projectMeta.Name,
		UID:        projectMeta.UID,
	})
	return nil
}

// InjectWorkflowOwnerRefModifier is a modifier of creating cyclone CRD resources.
// It will add workflow owner reference for the resource.
func InjectWorkflowOwnerRefModifier(tenant, project, wf string, object interface{}) error {
	var objectMeta *meta_v1.ObjectMeta
	switch obj := object.(type) {
	case *v1alpha1.WorkflowRun:
		objectMeta = &obj.ObjectMeta
	case *v1alpha1.WorkflowTrigger:
		objectMeta = &obj.ObjectMeta
	default:
		return fmt.Errorf("resource type not support")
	}

	wfMeta, err := getWfMetadata(tenant, wf)
	if err != nil {
		return err
	}
	// Add project OwnerReferences
	objectMeta.OwnerReferences = append(objectMeta.OwnerReferences, meta_v1.OwnerReference{
		APIVersion: v1alpha1.APIVersion,
		Kind:       reflect.TypeOf(v1alpha1.Workflow{}).Name(),
		Name:       wfMeta.Name,
		UID:        wfMeta.UID,
	})
	return nil
}

// InjectWorkflowLabelModifier is a modifier of create cyclone CRD resources.
// It will add workflow labels for the resource.
func InjectWorkflowLabelModifier(tenant, project, wf string, object interface{}) error {
	var objectMeta *meta_v1.ObjectMeta
	switch obj := object.(type) {
	case *v1alpha1.WorkflowRun:
		objectMeta = &obj.ObjectMeta
	case *v1alpha1.WorkflowTrigger:
		objectMeta = &obj.ObjectMeta
	default:
		return fmt.Errorf("resource type not support")
	}

	// Add workflow label
	if objectMeta.Labels == nil {
		objectMeta.Labels = make(map[string]string)
	}
	objectMeta.Labels[meta.LabelWorkflowName] = wf
	return nil
}

func listSCMRepos(scmSource *api.SCMSource) ([]scm.Repository, error) {
	sp, err := scm.GetSCMProvider(scmSource)
	if err != nil {
		log.Errorf("Fail to get SCM provider for %s", scmSource.Server)
		return nil, err
	}

	return sp.ListRepos()
}

func listSCMBranches(scmSource *api.SCMSource, repo string) ([]string, error) {
	sp, err := scm.GetSCMProvider(scmSource)
	if err != nil {
		log.Errorf("Fail to get SCM provider for %s", scmSource.Server)
		return nil, err
	}

	return sp.ListBranches(repo)
}

func listSCMTags(scmSource *api.SCMSource, repo string) ([]string, error) {
	sp, err := scm.GetSCMProvider(scmSource)
	if err != nil {
		log.Errorf("Fail to get SCM provider for %s", scmSource.Server)
		return nil, err
	}

	return sp.ListTags(repo)
}

func listSCMPullRequests(scmSource *api.SCMSource, repo, state string) ([]scm.PullRequest, error) {
	sp, err := scm.GetSCMProvider(scmSource)
	if err != nil {
		log.Errorf("Fail to get SCM provider for %s", scmSource.Server)
		return nil, err
	}

	return sp.ListPullRequests(repo, state)
}

func listSCMDockerfiles(scmSource *api.SCMSource, repo string) ([]string, error) {
	sp, err := scm.GetSCMProvider(scmSource)
	if err != nil {
		log.Errorf("Fail to get SCM provider for %s", scmSource.Server)
		return nil, err
	}

	return sp.ListDockerfiles(repo)
}

func createSCMStatus(scmSource *api.SCMSource, status v1alpha1.StatusPhase, recordURL string, event *scm.EventData) error {
	sp, err := scm.GetSCMProvider(scmSource)
	if err != nil {
		log.Errorf("Fail to get SCM provider for %s", scmSource.Server)
		return err
	}

	return sp.CreateStatus(status, recordURL, event.Repo, event.CommitSHA)
}

func generateRecordURL(tenant, project, wfName, wfrName string) (string, error) {
	type urlData struct {
		Tenant          string
		ProjectName     string
		WorkflowName    string
		WorkflowRunName string
	}

	data := urlData{
		tenant,
		project,
		wfName,
		wfrName,
	}
	tmpl, err := template.New("recordURL").Parse(config.GetRecordWebURLTemplate())
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	return buf.String(), err
}

func setSCMEventData(annos map[string]string, event *scm.EventData) (map[string]string, error) {
	bs, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	if annos == nil {
		annos = make(map[string]string)
	}

	annos[meta.AnnotationWorkflowRunSCMEvent] = string(bs)
	return annos, err
}

func getSCMEventData(annos map[string]string) (*scm.EventData, error) {
	if annos == nil {
		return nil, fmt.Errorf("Failed to parse SCM event as annotations are nil")
	}

	eventStr, ok := annos[meta.AnnotationWorkflowRunSCMEvent]
	if !ok {
		return nil, fmt.Errorf("Failed to parse SCM event as annotations do not have key %s", meta.AnnotationWorkflowRunSCMEvent)
	}

	event := &scm.EventData{}
	if err := json.Unmarshal([]byte(eventStr), event); err != nil {
		return nil, fmt.Errorf("Failed to parse SCM event as %v", err)
	}

	return event, nil
}
