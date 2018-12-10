package sonar

import (
	"os"
	"path/filepath"
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

func Test_findJavaBinaryFiles(t *testing.T) {
	testCases := map[string]struct {
		base    string
		files   map[string][]string
		results map[string]interface{}
	}{
		"success": {
			base: "/tmp/test-java-binary/",
			files: map[string][]string{
				"/":   []string{"a.war", "b.txt"},
				"/cc": []string{"c.jar", "d.go"},
			},
			results: map[string]interface{}{"/tmp/test-java-binary/a.war": nil, "/tmp/test-java-binary/cc/c.jar": nil},
		},
	}

	for d, tc := range testCases {
		// Prepare the test files.
		os.RemoveAll(tc.base)
		os.MkdirAll(tc.base, os.ModePerm)
		for k, v := range tc.files {
			os.MkdirAll(filepath.Join(tc.base, k), os.ModePerm)
			for _, file := range v {
				os.Create(filepath.Join(tc.base, k, file))
			}
		}

		results := findJavaBinaryFiles(tc.base)

		if len(results) != len(tc.results) {
			t.Errorf("fail to find java binary files for %s : expect %v, but got %v", d, len(tc.results), len(results))
		}
		for _, r := range results {
			_, ok := tc.results[r]
			if !ok {
				t.Errorf("fail to find java binary files for %s : %v not found in expect results", d, r)
			}
		}

		// clean the test files.
		os.RemoveAll(tc.base)
	}
}
