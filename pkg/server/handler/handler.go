package handler

import "github.com/caicloud/cyclone/pkg/k8s/clientset"

var (
	k8sClient clientset.Interface
)

// InitHandlers initializes the server resources handlers.
func InitHandlers(c clientset.Interface) {
	k8sClient = c
}
