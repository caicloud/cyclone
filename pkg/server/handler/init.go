package handler

import (
	"encoding/json"
	"time"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	// K8sClient is used to operate k8s resources
	K8sClient clientset.Interface
)

// Init initializes the server resources handlers.
func Init(c clientset.Interface) {
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

// BuildWfrStatusPatch builds patch for updating status of workflowrun
func BuildWfrStatusPatch(statusPhase v1alpha1.StatusPhase) ([]byte, error) {
	now := metav1.Time{Time: time.Now()}

	p := map[string]string{
		"/status/overall/phase":              string(statusPhase),
		"/status/overall/lastTransitionTime": now.UTC().Format(time.RFC3339),
	}
	return BuildPatch(p)
}
