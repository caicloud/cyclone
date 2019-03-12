package cerr

import (
	"fmt"
	"reflect"
	"testing"

	v1beta1 "k8s.io/api/extensions/v1beta1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
)

func TestConvertK8sError(t *testing.T) {
	testCases := map[string]struct {
		err         *k8serr.StatusError
		expectedErr error
	}{
		"Not found": {
			k8serr.NewNotFound(v1beta1.Resource("workflow"), "wf"),
			ErrorContentNotFound.Error(fmt.Sprintf("%s %s", "workflow", "wf")),
		},
		"Conflict": {
			k8serr.NewConflict(v1beta1.Resource("workflow"), "wf", nil),
			ErrorAlreadyExist.Error(fmt.Sprintf("%s %s", "workflow", "wf")),
		},
		"Already exist": {
			k8serr.NewAlreadyExists(v1beta1.Resource("workflow"), "wf"),
			ErrorAlreadyExist.Error(fmt.Sprintf("%s %s", "workflow", "wf")),
		},
	}

	for d, tc := range testCases {
		err := ConvertK8sError(tc.err)

		if !reflect.DeepEqual(err, tc.expectedErr) {
			t.Errorf("%s failed: expected to convert", d)
		}
	}
}
