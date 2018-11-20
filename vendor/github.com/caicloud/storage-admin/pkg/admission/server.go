package admission

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/golang/glog"
	admrv1 "k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/storage-admin/pkg/kubernetes"
)

type Server struct {
	kc        kubernetes.Interface
	caCert    []byte
	tlsConfig *tls.Config
	ruleMap   *RuleMap
}

func NewServer(kc kubernetes.Interface, certNamePath, serverSecretNamePath string) (*Server, error) {
	if kc == nil {
		return nil, fmt.Errorf("can't get clientset from getter")
	}

	caCert, e := getAPIServerCert(certNamePath, kc)
	if e != nil {
		return nil, e
	}

	serverCrt, serverKey, e := getServerSecrets(serverSecretNamePath, kc)
	if e != nil {
		return nil, e
	}

	tlsConfig, e := packTLSConfig(caCert, serverCrt, serverKey)
	if e != nil {
		return nil, e
	}

	s := &Server{
		kc:        kc,
		caCert:    caCert,
		tlsConfig: tlsConfig,
		ruleMap: &RuleMap{m: map[metav1.GroupVersionResource]*RuleWithChecker{
			*pvcResource: NewRuleWithChecker(pvcResource, admissionReviewCheckerPvc, admrv1.Create),
		}},
	}

	return s, nil
}

func (s *Server) SelfRegistration(appNamePath string) error {
	namespace, name, e := ParseNamePath(appNamePath, DefaultAppName, DefaultAppNamespace)
	if e != nil {
		return e
	}
	ruleWithOperations := s.ruleMap.GetRuleWithOperations()
	ec := initRegistrationObjects(namespace, name, s.caCert, ruleWithOperations)
	return selfRegistration(ec, s.kubeClient())
}

func (s *Server) Run(stopCh chan struct{}, port int) {
	http.HandleFunc("/", s.serve)
	server := &http.Server{
		Addr:      fmt.Sprintf(":%d", port),
		TLSConfig: s.tlsConfig,
	}

	go func() {
		e := server.ListenAndServeTLS("", "")
		if e != nil && e != http.ErrServerClosed {
			glog.Fatalf("ListenAndServe failed, %v", e)
		}
		glog.Info("ListenAndServe closed")
	}()

	<-stopCh
	server.Close()
}

func (s *Server) kubeClient() kubernetes.Interface {
	return s.kc
}
