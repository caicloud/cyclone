package sonar

import (
	"testing"
)

func Test_extractCeTaskId(t *testing.T) {

	testCases := map[string]struct {
		path string
		id   string
	}{
		"success": {
			path: "testdata/report-task.txt",
			id:   "AWd9EPiuomqmtD7hc2iT",
		},
	}

	for d, tc := range testCases {

		id, err := extractCeTaskId(tc.path)
		if err != nil || id != tc.id {
			t.Errorf("fail to extract ceTaskId by %s : expect %v, but got %v , error: %v", d, tc.id, id, err)
		}

	}
}
