package util

import (
	"fmt"
	"strings"
	"testing"

	resv1b1 "github.com/caicloud/clientset/pkg/apis/resource/v1beta1"
	storagev1 "k8s.io/api/storage/v1"
)

func TestStorageServiceTypeAndNameFilter(t *testing.T) {
	// init
	ssNum := 5
	names := []string{"05", "04", "03", "02", "01"}
	alias := []string{"o5", "e4", "O3", "E2", "o1"}
	types := []string{"t5", "T3", "t3", "t1", "t1"}
	in := make([]resv1b1.StorageService, 5)
	for i := 0; i < ssNum; i++ {
		in[i].Name = names[i]
		in[i].TypeName = types[i]
		SetObjectAlias(&in[i], alias[i])
	}
	// struct
	type testCase struct {
		alias string
		tp    string
		count int
	}
	testCaseInfo := func(c *testCase) string {
		return fmt.Sprintf("[tp:%s][al:%s][num:%d]", c.tp, c.alias, c.count)
	}
	testCaseTest := func(c *testCase) {
		describe := testCaseInfo(c)
		out := StorageServiceTypeAndNameFilter(in, c.tp, c.alias)
		if len(out) != c.count {
			t.Fatalf("%s test count not match, want %d got %d", describe, c.count, len(out))
		}
		// test sort
		for i := 1; i < len(out); i++ {
			if out[i].Name <= out[i-1].Name {
				names := ""
				for i := range out {
					names += out[i].Name + ", "
				}
				t.Fatalf("%s test sort failed, %v", describe, names)
			}
		}
		t.Logf("%s test sort done", describe)
		// test tp
		if len(c.tp) > 0 {
			for i := range out {
				if out[i].TypeName != c.tp {
					t.Fatalf("%s test tp failed, want %v got %v", describe, c.tp, out[i].TypeName)
				}
			}
			t.Logf("%s test tp done", describe)
		}
		// test alias
		if len(c.alias) > 0 {
			for i := range out {
				aliasLower := strings.ToLower(GetObjectAlias(&out[i]))
				if !strings.Contains(aliasLower, strings.ToLower(c.alias)) {
					t.Fatalf("%s test alias failed, want %v got %v", describe, c.alias, GetObjectAlias(&out[i]))
				}
			}
			t.Logf("%s test alias done", describe)
		}
	}
	caseList := []testCase{
		{ // sort
			alias: "",
			tp:    "",
			count: 5,
		},
		{ // alias low
			alias: "e",
			tp:    "",
			count: 2,
		},
		{ // alias upper
			alias: "E",
			tp:    "",
			count: 2,
		},
		{ // alias other
			alias: "o",
			tp:    "",
			count: 3,
		},
		{ // alias none
			alias: "xxx",
			tp:    "",
			count: 0,
		},
		{ // tp
			alias: "",
			tp:    "t1",
			count: 2,
		},
		{ // tp upper
			alias: "",
			tp:    "t3",
			count: 1,
		},
		{ // tp upper
			alias: "",
			tp:    "T3",
			count: 1,
		},
		{ // mix
			alias: "o",
			tp:    "t1",
			count: 1,
		},
	}
	for i := range caseList {
		testCaseTest(&caseList[i])
	}
}

func TestStorageClassTypeAndNameFilter(t *testing.T) {
	// init
	ssNum := 5
	names := []string{"05", "04", "03", "02", "01"}
	alias := []string{"o5", "e4", "O3", "E2", "o1"}
	types := []string{"t5", "T3", "t3", "t1", "t1"}
	in := make([]storagev1.StorageClass, 5)
	for i := 0; i < ssNum; i++ {
		in[i].Name = names[i]
		SetClassType(&in[i], types[i])
		SetObjectAlias(&in[i], alias[i])
	}
	// struct
	type testCase struct {
		alias string
		tp    string
		count int
	}
	testCaseInfo := func(c *testCase) string {
		return fmt.Sprintf("[tp:%s][al:%s][num:%d]", c.tp, c.alias, c.count)
	}
	testCaseTest := func(c *testCase) {
		describe := testCaseInfo(c)
		out := StorageClassTypeAndNameFilter(in, c.tp, c.alias)
		if len(out) != c.count {
			t.Fatalf("%s test count not match, want %d got %d", describe, c.count, len(out))
		}
		// test sort
		for i := 1; i < len(out); i++ {
			if out[i].Name <= out[i-1].Name {
				names := ""
				for i := range out {
					names += out[i].Name + ", "
				}
				t.Fatalf("%s test sort failed, %v", describe, names)
			}
		}
		t.Logf("%s test sort done", describe)
		// test tp
		if len(c.tp) > 0 {
			for i := range out {
				if GetClassType(&out[i]) != c.tp {
					t.Fatalf("%s test tp failed, want %v got %v", describe, c.tp, GetClassType(&out[i]))
				}
			}
			t.Logf("%s test tp done", describe)
		}
		// test alias
		if len(c.alias) > 0 {
			for i := range out {
				aliasLower := strings.ToLower(GetObjectAlias(&out[i]))
				if !strings.Contains(aliasLower, strings.ToLower(c.alias)) {
					t.Fatalf("%s test alias failed, want %v got %v", describe, c.alias, GetObjectAlias(&out[i]))
				}
			}
			t.Logf("%s test alias done", describe)
		}
	}
	caseList := []testCase{
		{ // sort
			alias: "",
			tp:    "",
			count: 5,
		},
		{ // alias low
			alias: "e",
			tp:    "",
			count: 2,
		},
		{ // alias upper
			alias: "E",
			tp:    "",
			count: 2,
		},
		{ // alias other
			alias: "o",
			tp:    "",
			count: 3,
		},
		{ // alias none
			alias: "xxx",
			tp:    "",
			count: 0,
		},
		{ // tp
			alias: "",
			tp:    "t1",
			count: 2,
		},
		{ // tp upper
			alias: "",
			tp:    "t3",
			count: 1,
		},
		{ // tp upper
			alias: "",
			tp:    "T3",
			count: 1,
		},
		{ // mix
			alias: "o",
			tp:    "t1",
			count: 1,
		},
	}
	for i := range caseList {
		testCaseTest(&caseList[i])
	}
}
