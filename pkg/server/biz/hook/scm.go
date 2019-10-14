package hook

import (
	"fmt"
	"strings"
	"sync"

	"github.com/caicloud/nirvana/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/meta"
	api "github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	in "github.com/caicloud/cyclone/pkg/server/biz/integration"
	"github.com/caicloud/cyclone/pkg/server/biz/scm"
	"github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/server/config"
	"github.com/caicloud/cyclone/pkg/server/handler"
	"github.com/caicloud/cyclone/pkg/util/cerr"
)

var once sync.Once
var scmManager *SCMManager

// SCMManager represents manager of SCM instances.
type SCMManager struct {
	mutex sync.Mutex
}

// getScmManager returns the scm manager instance
func getScmManager() *SCMManager {
	once.Do(func() {
		scmManager = &SCMManager{
			mutex: sync.Mutex{},
		}
	})

	return scmManager
}

// Register registers SCM webhook if if has not been registered.
func (*SCMManager) Register(tenant string, wft v1alpha1.WorkflowTrigger) error {
	var wftName, secretName, repo = wft.Name, wft.Spec.SCM.Secret, wft.Spec.SCM.Repo

	log.Infof("start to register webhook for %s/%s , %s", secretName, repo, wftName)
	scmManager.mutex.Lock()
	defer scmManager.mutex.Unlock()

	wfts, err := ListSCMWfts(tenant, repo, secretName)
	if err != nil {
		log.Infof("list wfts by %s/%s error %s", secretName, repo, err)
		return err
	}

	if len(wfts.Items) > 0 {
		log.Infof("webhook of %s/%s has already registered, using wfts length: %d", repo, secretName, len(wfts.Items))
		return nil
	}

	log.Infof("start to create scm webhook %s/%s for wft %s", secretName, repo, wftName)
	secret, err := handler.K8sClient.CoreV1().Secrets(common.TenantNamespace(tenant)).Get(secretName, metav1.GetOptions{})
	if err != nil {
		return cerr.ConvertK8sError(err)
	}

	integration, err := in.FromSecret(secret)
	if err != nil {
		log.Error(err)
		return err
	}

	err = createSCMWebhook(integration.Spec.SCM, tenant, secretName, repo)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

// createSCMWebhook creates webhook for SCM repo.
func createSCMWebhook(scmSource *api.SCMSource, tenant, secret, repo string) error {
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

	err = sp.CreateWebhook(repo, webhook)
	if cerr.ErrorExternalAuthenticationFailed.Derived(err) || cerr.ErrorExternalAuthorizationFailed.Derived(err) ||
		// GitHub sends a 404 error when your client isn't properly authenticated.
		// More info: https://developer.github.com/v3/troubleshooting/
		(cerr.ErrorExternalNotFound.Derived(err) && scmSource.Type == api.GitHub) {
		return cerr.ErrorCreateWebhookPermissionDenied.Error()
	}
	return err
}

func generateWebhookURL(tenant, secret string) string {
	urlPrefix := strings.TrimPrefix(config.GetWebhookURLPrefix(), "/")
	// Construct webhook URL, refer to cyclone/pkg/server/apis/v1alpha1/descriptors/webhook.go
	return fmt.Sprintf("%s/tenants/%s/webhook?sourceType=SCM&integration=%s", urlPrefix, tenant, secret)
}

// Unregister unregisters SCM webhook if if has no other wft using.
func (o *SCMManager) Unregister(tenant string, wft v1alpha1.WorkflowTrigger) error {
	var wftName, secretName, repo = wft.Name, wft.Spec.SCM.Secret, wft.Spec.SCM.Repo

	log.Infof("start to unregister webhook for %s/%s , %s", secretName, repo, wftName)
	o.mutex.Lock()
	defer o.mutex.Unlock()

	wfts, err := ListSCMWfts(tenant, repo, secretName)
	if err != nil {
		return err
	}

	if len(wfts.Items) > 1 {
		log.Infof("there are other wfts using the webhook %s/%s, skip deleting, using wfts length: %d", repo, secretName, len(wfts.Items))
		return nil
	}

	log.Infof("start to delete scm webhook %s/%s for wft %s", secretName, repo, wftName)
	secret, err := handler.K8sClient.CoreV1().Secrets(common.TenantNamespace(tenant)).Get(
		secretName, metav1.GetOptions{})
	if err != nil {
		return cerr.ConvertK8sError(err)
	}

	log.Infof("Delete webhook for repo %s", repo)
	integration, err := in.FromSecret(secret)
	if err != nil {
		log.Error(err)
		return err
	}

	err = deleteSCMWebhook(integration.Spec.SCM, tenant, secretName, repo)
	if err != nil {
		log.Error("delete webhook error:%v", err)
	}

	return nil
}

// deleteSCMWebhook deletes webhook from SCM repo.
func deleteSCMWebhook(scmSource *api.SCMSource, tenant, secret, repo string) error {
	sp, err := scm.GetSCMProvider(scmSource)
	if err != nil {
		log.Errorf("Fail to get SCM provider for %s", scmSource.Server)
		return err
	}

	return sp.DeleteWebhook(repo, generateWebhookURL(tenant, secret))
}

// LabelSCMTrigger add labels about scm trigger
func LabelSCMTrigger(wft *v1alpha1.WorkflowTrigger) {
	if wft.Labels == nil {
		wft.Labels = make(map[string]string)
	}
	wft.Labels[meta.LabelWftEventSource] = wft.Spec.SCM.Secret
}

// ListSCMWfts list all related SCM type workflow triggers
func ListSCMWfts(tenant, repo, integration string) (*v1alpha1.WorkflowTriggerList, error) {
	labelMap := make(map[string]string)

	// Did not use repo as a label since repo length may greater than the label length limit 64.
	if integration != "" {
		labelMap[meta.LabelWftEventSource] = integration
	}

	listOption := metav1.ListOptions{
		LabelSelector: labels.Set(labelMap).String(),
	}

	originWfts, err := handler.K8sClient.CycloneV1alpha1().WorkflowTriggers(common.TenantNamespace(tenant)).List(listOption)
	if err != nil {
		return nil, err
	}

	wfts := originWfts.DeepCopy()
	wfts.Items = make([]v1alpha1.WorkflowTrigger, 0)
	for _, wft := range originWfts.Items {
		if wft.Spec.SCM.Repo == repo {
			wfts.Items = append(wfts.Items, wft)
		}
	}
	return wfts, nil
}
