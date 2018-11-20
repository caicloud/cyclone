package admission

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	DefaultAppName               = "storage-admission"
	DefaultAppNamespace          = "kube-system"
	DefaultCertName              = "extension-apiserver-authentication"
	DefaultCertNamespace         = "kube-system"
	DefaultServerSecretName      = "apiserver-tls-secret"
	DefaultServerSecretNamespace = "kube-system"

	DefaultExternalAdmissionHookNameReign = ".caicloud.io"

	EnvAppNamePath          = "ENV_APP_NAME_PATH"
	EnvCertNamePath         = "ENV_CERT_NAME_PATH"
	EnvServerSecretNamePath = "ENV_SECRET_NAME_PATH"
	EnvIgnoredNamespaces    = "ENV_IGNORED_NAMESPACES"
	EnvIgnoredClasses       = "ENV_IGNORED_CLASSES"

	DefaultAppNamePath          = DefaultAppNamespace + "/" + DefaultAppName
	DefaultCertNamePath         = DefaultCertNamespace + "/" + DefaultCertName
	DefaultServerSecretNamePath = DefaultServerSecretNamespace + "/" + DefaultServerSecretName
	DefaultIgnoredNamespaces    = metav1.NamespaceDefault + "," + metav1.NamespaceSystem
	DefaultIgnoredClasses       = ""
)
