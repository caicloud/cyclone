package handler

import (
	"encoding/json"

	"github.com/caicloud/cyclone/pkg/k8s/clientset"
)

var (
	// K8sClient is used to operate k8s resources
	K8sClient clientset.Interface
)

// InitHandlers initializes the server resources handlers.
func InitHandlers(c clientset.Interface) {
	K8sClient = c
}

// BuildPatch builds string patch from input p map,
// map key used as patch path, map value used as patch value.
func BuildPatch(p map[string]string) ([]byte, error) {
	patches := []JSONPatch{}
	for p, v := range p {
		patch := JSONPatch{
			Op:    "replace",
			Path:  p,
			Value: v,
		}

		patches = append(patches, patch)

	}

	patchesBytes, err := json.Marshal(patches)
	if err != nil {
		return nil, err
	}
	return patchesBytes, nil
}

// JSONPatch contains information about application/json-patch+json type patch
type JSONPatch struct {
	// Op represents operation type
	Op string `json:"op"`
	// Path represents elements will be operate
	Path string `json:"path"`
	// Value represents the new value of the operation elements
	Value string `json:"value"`
}

// BuildPVCPatch builds patch for pvc
func BuildPVCPatch(storageclass, size string) ([]byte, error) {
	p := map[string]string{
		"/spec/storageClassName":           storageclass,
		"/spec/resources/limits/storage":   size,
		"/spec/resources/requests/storage": size,
	}
	return BuildPatch(p)
}
