package v1alpha1

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/caicloud/nirvana/log"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/meta"
	api "github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
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
	wfr.Labels[meta.LabelWorkflowName] = wf
	return nil
}

// MergeMap merges map 'in' into map 'out', if there is an element in both 'in' and 'out',
// we will use the value of 'in'.
func MergeMap(in, out map[string]string) map[string]string {
	if in != nil {
		if out == nil {
			out = make(map[string]string)
		}

		for k, v := range in {
			out[k] = v
		}
	}

	return out
}

// CreateSCMWebhook creates webhook for SCM repo.
func CreateSCMWebhook(scmSource *api.SCMSource, tenant, secret, repo string) error {
	sp, err := scm.GetSCMProvider(scmSource)
	if err != nil {
		log.Errorf("Fail to get SCM provider for %s", scmSource.Server)
		return err
	}

	webhook := &scm.Webhook{
		URL: generateWebhookURL(tenant, secret),
		Events: []scm.EventType{
			scm.PushEventType,
			scm.TagReleaseEventType,
			scm.PullRequestEventType,
			scm.PullRequestCommentEventType,
		},
	}

	return sp.CreateWebhook(repo, webhook)
}

// DeleteSCMWebhook deletes webhook from SCM repo.
func DeleteSCMWebhook(scmSource *api.SCMSource, tenant, secret, repo string) error {
	sp, err := scm.GetSCMProvider(scmSource)
	if err != nil {
		log.Errorf("Fail to get SCM provider for %s", scmSource.Server)
		return err
	}

	return sp.DeleteWebhook(repo, generateWebhookURL(tenant, secret))
}

func generateWebhookURL(tenant, secret string) string {
	webhookURL := strings.TrimPrefix(config.Config.WebhookURL, "/")
	// Construct webhook URL, refer to cyclone/pkg/server/apis/v1alpha1/descriptors/webhook.go
	return fmt.Sprintf("%s/tenants/%s/integrations/%s/webhook", webhookURL, tenant, secret)
}

func getReposFromSecret(tenant, secretName string) (map[string][]string, error) {
	repos := map[string][]string{}
	secret, err := handler.K8sClient.CoreV1().Secrets(common.TenantNamespace(tenant)).Get(
		common.IntegrationSecret(secretName), meta_v1.GetOptions{})
	if err != nil {
		return repos, cerr.ConvertK8sError(err)
	}

	if d, ok := secret.Data[common.SecretKeyRepos]; ok {
		if err = json.Unmarshal(d, &repos); err != nil {
			log.Errorf("Failed to unmarshal repos from secret")
			return repos, err
		}
	}

	return repos, nil
}
