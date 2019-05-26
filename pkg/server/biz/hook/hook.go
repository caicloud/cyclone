package hook

import (
	"fmt"
	"strings"
	"sync"

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
	"github.com/caicloud/nirvana/log"
)

var o *Operator

// Operator
type Operator struct {
	mutex sync.Mutex
}

// GetInstance returns the Operator instance
func GetInstance() *Operator {
	if o == nil {
		o = &Operator{
			mutex: sync.Mutex{},
		}
	}

	return o
}

// ListWfts list all relative workflow triggers
func (*Operator) ListWfts(tenant, repo, integration string) (*v1alpha1.WorkflowTriggerList, error) {
	labelMap := make(map[string]string)
	if repo != "" {
		labelMap[meta.LabelRepoOfWft] = encodeRepoValue(repo)
	}
	if integration != "" {
		labelMap[meta.LabelIntegrationOfWft] = integration
	}

	listOption := metav1.ListOptions{
		LabelSelector: labels.Set(labelMap).String(),
	}

	wfts, err := handler.K8sClient.CycloneV1alpha1().WorkflowTriggers(common.TenantNamespace(tenant)).List(listOption)
	if err != nil {
		return nil, err
	}

	return wfts, nil
}

// RegisterSCMWebhook registers SCM webhook if if has not been registered.
func (*Operator) RegisterSCMWebhook(tenant, wftName, secretName, repo string) error {
	log.Infof("start to register webhook for %s/%s , %s", secretName, repo, wftName)
	o.mutex.Lock()
	defer o.mutex.Unlock()

	wfts, err := o.ListWfts(tenant, repo, secretName)
	if err != nil {
		log.Infof("list wfts by %s/%s error %s", secretName, repo, err)
		return err
	}

	if len(wfts.Items) > 0 {
		log.Infof("webhook of %s/%s has already registered, using wfts length:%s", repo, secretName, len(wfts.Items))
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

	return sp.CreateWebhook(repo, webhook)
}

func generateWebhookURL(tenant, secret string) string {
	webhookURL := strings.TrimPrefix(config.Config.WebhookURL, "/")
	// Construct webhook URL, refer to cyclone/pkg/server/apis/v1alpha1/descriptors/webhook.go
	return fmt.Sprintf("%s/tenants/%s/webhook?eventType=SCM&integration=%s", webhookURL, tenant, secret)
}

// UnregisterSCMWebhook unregisters SCM webhook if if has no other wft using.
func (o *Operator) UnregisterSCMWebhook(tenant, wftName, secretName, repo string) error {
	log.Infof("start to unregister webhook for %s/%s , %s", secretName, repo, wftName)
	o.mutex.Lock()
	defer o.mutex.Unlock()

	wfts, err := o.ListWfts(tenant, repo, secretName)
	if err != nil {
		return err
	}

	if len(wfts.Items) > 1 {
		log.Infof("there are other wfts using the webhook %s/%s, skip deleting, using wfts length:", repo, secretName, len(wfts.Items))
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
	wft.Labels[meta.LabelIntegrationOfWft] = wft.Spec.SCM.Secret
	wft.Labels[meta.LabelRepoOfWft] = encodeRepoValue(wft.Spec.SCM.Repo)
}

func encodeRepoValue(repo string) string {
	return strings.Replace(repo, "/", ".", -1)
}
