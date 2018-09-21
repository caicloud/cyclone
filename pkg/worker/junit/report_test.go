package junit

import (
	"os"
	"strings"
	"testing"

	osutil "github.com/caicloud/cyclone/pkg/util/os"
)

func TestConvertPerformStageSet(t *testing.T) {

	testCases := map[string]struct {
		base    string
		files   map[string]string
		results map[string]interface{}
	}{
		"success": {
			base: "/tmp/test-results/",
			files: map[string]string{
				"test.xml": `<testsuite name="com.example.demo.DemoApplicationTests" tests="1" skipped="0" failures="0" errors="0" timestamp="2018-09-17T13:32:25" hostname="15421a702ece" time="0.247">
</testsuite>`,
				"test.go": "xxxxx",
				"test1.xml": `<testsuite name="com.example.demo.DemoApplicationTests" tests="1" skipped="0" failures="0" errors="0" timestamp="2018-09-17T13:32:25" hostname="15421a702ece" time="0.247">
</testsuite>`,
				"test2.xml": `<testsuite name="com.example.demo.DemoApplicationTests"`,
			},
			results: map[string]interface{}{"/tmp/test-results/test.xml": nil, "/tmp/test-results/test1.xml": nil},
		},
	}

	for d, tc := range testCases {
		// Prepare the test files.
		os.RemoveAll(tc.base)
		os.MkdirAll(tc.base, os.ModePerm)
		for k, v := range tc.files {
			osutil.ReplaceFile(tc.base+k, strings.NewReader(v))
		}

		report := NewReport(tc.base)
		results := report.FindReportFiles()

		if len(results) != len(tc.results) {
			t.Errorf("fail to find report files for %s : expect %v, but got %v", d, len(tc.results), len(results))
		}
		for _, r := range results {
			_, ok := tc.results[r]
			if !ok {
				t.Errorf("fail to find report files for %s : %v not found in expect results", d, r)
			}
		}

		// clean the test files.
		os.RemoveAll(tc.base)
	}
}
