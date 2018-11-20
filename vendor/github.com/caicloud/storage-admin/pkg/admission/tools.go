package admission

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/golang/glog"
	admsv1 "k8s.io/api/admission/v1beta1"
	admrv1 "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/storage-admin/pkg/kubernetes"
)

// multi admission checker about

var (
	pvcResource = &metav1.GroupVersionResource{
		Group:    corev1.SchemeGroupVersion.Group,
		Version:  corev1.SchemeGroupVersion.Version,
		Resource: "persistentvolumeclaims"}
)

type RuleMap struct {
	m map[metav1.GroupVersionResource]*RuleWithChecker
}

func (m *RuleMap) Register(gvr *metav1.GroupVersionResource, f CheckAdmissionReviewFunc,
	ops ...admrv1.OperationType) {
	m.m[*gvr] = NewRuleWithChecker(gvr, f, ops...)
}
func (m *RuleMap) GetChecker(ar *admsv1.AdmissionReview) (CheckAdmissionReviewFunc, error) {
	re := m.m[ar.Request.Resource]
	if re == nil {
		return nil, fmt.Errorf("resource type not registed")
	}
	return re.CheckFunc, nil
}
func (m *RuleMap) GetRuleWithOperations() []admrv1.RuleWithOperations {
	re := make([]admrv1.RuleWithOperations, 0, len(m.m))
	for _, v := range m.m {
		re = append(re, v.RuleWithOperations)
	}
	return re
}

type CheckAdmissionReviewFunc func(kc kubernetes.Interface, ar *admsv1.AdmissionReview) error

type RuleWithChecker struct {
	admrv1.RuleWithOperations
	CheckFunc CheckAdmissionReviewFunc
}

func NewRuleWithChecker(gvr *metav1.GroupVersionResource, f CheckAdmissionReviewFunc,
	ops ...admrv1.OperationType) *RuleWithChecker {
	return &RuleWithChecker{
		RuleWithOperations: admrv1.RuleWithOperations{
			Operations: ops,
			Rule: admrv1.Rule{
				APIGroups:   []string{gvr.Group},
				APIVersions: []string{gvr.Version},
				Resources:   []string{gvr.Resource},
			},
		},
		CheckFunc: f,
	}
}

// secrets about

func ParseNamePath(namePath, defaultNamespace, defaultName string) (namespace, name string, e error) {
	vec := strings.Split(namePath, "/")
	switch len(vec) {
	case 1:
		namespace = defaultNamespace
		if name = vec[0]; len(name) == 0 { // include namePath == ""
			name = defaultName
		}
		return
	case 2:
		if namespace = vec[0]; len(namespace) == 0 {
			namespace = defaultNamespace
		}
		if name = vec[1]; len(name) == 0 {
			name = defaultName
		}
		return
	default:
		return defaultNamespace, defaultName, fmt.Errorf("bad name path %s", namePath)
	}
}

func getAPIServerCert(certNamePath string, clientset kubernetes.Interface) ([]byte, error) {
	namespace, name, e := ParseNamePath(certNamePath, DefaultCertNamespace, DefaultCertName)
	if e != nil {
		return nil, e
	}
	c, e := clientset.CoreV1().ConfigMaps(namespace).Get(name, metav1.GetOptions{})
	if e != nil {
		return nil, e
	}

	pem, ok := c.Data["requestheader-client-ca-file"]
	if !ok {
		return nil, fmt.Errorf("cannot find the ca.crt in the configmap, configMap.Data is %#v", c.Data)
	}
	return []byte(pem), nil
}

func getServerSecrets(serverSecretNamePath string, clientset kubernetes.Interface) (serverCrt, serverKey []byte, e error) {
	namespace, name, e := ParseNamePath(serverSecretNamePath, DefaultServerSecretNamespace, DefaultServerSecretName)
	if e != nil {
		return nil, nil, e
	}
	secret, e := clientset.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
	if e != nil {
		return nil, nil, e
	}

	if secret.Type != corev1.SecretTypeTLS {
		return nil, nil, fmt.Errorf("%s/%s type must be %s", namespace, name, corev1.SecretTypeTLS)
	}

	serverCrt = secret.Data[corev1.TLSCertKey]
	serverKey = secret.Data[corev1.TLSPrivateKeyKey]
	return serverCrt, serverKey, nil
}

func packTLSConfig(caCert, serverCrt, serverKey []byte) (*tls.Config, error) {
	apiserverCA := x509.NewCertPool()
	apiserverCA.AppendCertsFromPEM(caCert)

	sCert, e := tls.X509KeyPair(serverCrt, serverKey)
	if e != nil {
		return nil, e
	}
	return &tls.Config{
		Certificates: []tls.Certificate{sCert},
		ClientCAs:    apiserverCA,
		// ClientAuth:   tls.NoClientCert,
	}, nil
}

// webhook self registration about

func initRegistrationObjects(svNamespace, name string, caCert []byte,
	rules []admrv1.RuleWithOperations) *admrv1.ValidatingWebhookConfiguration {
	// admission
	failurePolicy := admrv1.Ignore
	ec := &admrv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Webhooks: []admrv1.Webhook{
			{
				Name:  name + DefaultExternalAdmissionHookNameReign,
				Rules: rules,
				ClientConfig: admrv1.WebhookClientConfig{
					Service: &admrv1.ServiceReference{
						Name:      name,
						Namespace: svNamespace,
					},
					CABundle: caCert,
				},
				FailurePolicy: &failurePolicy,
			},
		},
	}
	return ec
}

func selfRegistration(ec *admrv1.ValidatingWebhookConfiguration, kc kubernetes.Interface) error {
	glog.V(2).Info(toJson(ec))

	prevEC, e := kc.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Get(ec.Name, metav1.GetOptions{})
	if e == nil {
		prevEC.Webhooks = ec.Webhooks
		glog.Infof("config update")
		_, e = kc.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Update(prevEC)
	} else if kubernetes.IsNotFound(e) {
		glog.Infof("config create")
		_, e = kc.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Create(ec)
	}
	if e != nil {
		return e
	}
	return nil
}

// logs about

func admissionReviewToString(ar *admsv1.AdmissionReview) string {
	type infoObj struct {
		metav1.ObjectMeta `json:"metadata"`
	}
	var obj infoObj
	json.Unmarshal(ar.Request.Object.Raw, &obj)
	return fmt.Sprintf("[op:%s][path:%s/%s/%s][obj:%s/%s]",
		ar.Request.Operation, ar.Request.Resource.Group, ar.Request.Resource.Version, ar.Request.Resource.Resource,
		obj.Namespace, obj.Name)
}

func toJson(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
