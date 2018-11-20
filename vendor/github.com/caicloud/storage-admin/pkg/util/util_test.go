package util

import (
	"fmt"
	"testing"
)

func TestGetRequestPageStartAndLimit(t *testing.T) {
	type testCase struct {
		strings [2]string
		results [2]int
		isError bool
	}
	testCaseInfo := func(c *testCase) string {
		if c.isError {
			return fmt.Sprintf("case input:[start=%s][limit=%s] output:[err=%v]",
				c.strings[0], c.strings[1], c.isError)
		} else {
			return fmt.Sprintf("case input:[start=%s][limit=%s] output:[start=%d][limit=%d]",
				c.strings[0], c.strings[1], c.results[0], c.results[1])
		}
	}
	testCaseTest := func(c *testCase) {
		start, limit, fe := getRequestPageStartAndLimit(c.strings[0], c.strings[1])
		switch {
		case c.isError && fe != nil:
			t.Logf("test %s done", testCaseInfo(c))
		case c.isError && fe == nil:
			t.Fatalf("test %s expected error but get nil", testCaseInfo(c))
		case !c.isError && fe != nil:
			t.Fatalf("test %s unexpected error %v", testCaseInfo(c), fe.Error())
		case !c.isError && fe == nil && (start != c.results[0] || limit != c.results[1]):
			t.Fatalf("test %s result [start=%d][limit=%d] unexpected", testCaseInfo(c), start, limit)
		case !c.isError && fe == nil && (start == c.results[0] && limit == c.results[1]):
			t.Logf("test %s done", testCaseInfo(c))
		default:
			t.Fatalf("test %s unexpected case!!!", testCaseInfo(c))
		}
	}
	caseList := []testCase{
		// case 1
		{strings: [2]string{"", ""}, results: [2]int{0, 0}, isError: false},
		// case 2
		{strings: [2]string{"", "a"}, results: [2]int{0, 0}, isError: true},
		{strings: [2]string{"", "0"}, results: [2]int{0, 0}, isError: true},
		{strings: [2]string{"", "1"}, results: [2]int{0, 1}, isError: false},
		// case 3
		{strings: [2]string{"a", ""}, results: [2]int{0, 0}, isError: true},
		{strings: [2]string{"-1", ""}, results: [2]int{0, 0}, isError: true},
		{strings: [2]string{"1", ""}, results: [2]int{1, 0}, isError: false},
		// case 4
		{strings: [2]string{"a", "1"}, results: [2]int{0, 0}, isError: true},
		{strings: [2]string{"1", "a"}, results: [2]int{0, 0}, isError: true},
		{strings: [2]string{"a", "a"}, results: [2]int{0, 0}, isError: true},
		{strings: [2]string{"-1", "1"}, results: [2]int{0, 0}, isError: true},
		{strings: [2]string{"1", "0"}, results: [2]int{0, 0}, isError: true},
		{strings: [2]string{"-1", "0"}, results: [2]int{0, 0}, isError: true},
		{strings: [2]string{"0", "1"}, results: [2]int{0, 1}, isError: false},
		{strings: [2]string{"1", "1"}, results: [2]int{1, 1}, isError: false},
	}
	for i := range caseList {
		testCaseTest(&caseList[i])
	}
}

func TestGetStartLimitEnd(t *testing.T) {
	type testCase struct {
		start, limit, arrayLen int
		end                    int
	}
	testCaseInfo := func(c *testCase) string {
		return fmt.Sprintf("case input:[start=%d][limit=%d][len=%d] output:[end=%d]",
			c.start, c.limit, c.arrayLen, c.end)
	}
	testCaseTest := func(c *testCase) {
		end := GetStartLimitEnd(c.start, c.limit, c.arrayLen)
		if end == c.end {
			t.Logf("test %s done", testCaseInfo(c))
		} else {
			t.Fatalf("test %s result:[end=%d] unexpected", testCaseInfo(c), end)
		}
	}
	const (
		lenFull = 10

		startSet  = 0
		startHalf = lenFull/2 - 1
		startFull = lenFull - 1

		limitSet  = 0
		limitMin  = 1
		limitHalf = lenFull/2 - 1 // startHalf+limitHalf<lenFull
		limitMax  = lenFull
	)
	// start < arrayLen guaranteed by upper
	caseList := []testCase{
		{start: startSet, limit: limitSet, arrayLen: lenFull, end: lenFull},
		{start: startSet, limit: limitMin, arrayLen: lenFull, end: startSet + 1},
		{start: startSet, limit: limitHalf, arrayLen: lenFull, end: startSet + limitHalf},
		{start: startSet, limit: limitMax, arrayLen: lenFull, end: lenFull},

		{start: startHalf, limit: limitSet, arrayLen: lenFull, end: lenFull},
		{start: startHalf, limit: limitMin, arrayLen: lenFull, end: startHalf + 1},
		{start: startHalf, limit: limitHalf, arrayLen: lenFull, end: startHalf + limitHalf},
		{start: startHalf, limit: limitMax, arrayLen: lenFull, end: lenFull},
		{start: startHalf, limit: lenFull - startHalf, arrayLen: lenFull, end: lenFull},

		{start: startFull, limit: limitSet, arrayLen: lenFull, end: lenFull},
		{start: startFull, limit: limitMin, arrayLen: lenFull, end: lenFull},
		{start: startFull, limit: limitHalf, arrayLen: lenFull, end: lenFull},
		{start: startFull, limit: limitMax, arrayLen: lenFull, end: lenFull},
	}
	for i := range caseList {
		testCaseTest(&caseList[i])
	}
}

func Test_checkOptionalMapParameters(t *testing.T) {
	type testCase struct {
		describe string
		input    map[string]string
		optional map[string]string
		isError  bool
	}
	testCaseTest := func(c *testCase) {
		// all input keys should exist in optional
		fe := checkOptionalMapParameters(c.input, c.optional)
		switch {
		case c.isError && fe == nil:
			t.Fatalf("test case <%s> should get an error", c.describe)
		case !c.isError && fe != nil:
			t.Fatalf("test case <%s> got unexpected error, %v", c.describe, fe.Error())
		default:
			t.Logf("test case <%s> done", c.describe)
		}
	}
	caseList := []testCase{
		{
			describe: "all match",
			input:    map[string]string{"ka": "va", "kb": "vb"},
			optional: map[string]string{"ka": "ta", "kb": "tb"},
			isError:  false,
		},
		{
			describe: "all empty",
			input:    map[string]string{},
			optional: map[string]string{},
			isError:  false,
		},
		{
			describe: "one more",
			input:    map[string]string{"ka": "va", "kb": "vb", "kc": "vc"},
			optional: map[string]string{"ka": "ta", "kb": "tb"},
			isError:  true,
		},
		{
			describe: "one less",
			input:    map[string]string{"ka": "va"},
			optional: map[string]string{"ka": "ta", "kb": "tb"},
			isError:  false,
		},
		{
			describe: "one diff",
			input:    map[string]string{"ka": "va", "kc": "vc"},
			optional: map[string]string{"ka": "ta", "kb": "tb"},
			isError:  true,
		},
	}
	for i := range caseList {
		testCaseTest(&caseList[i])
	}
}

func TestSyncStorageClassWithTypeAndService(t *testing.T) {
	type testCase struct {
		describe   string
		scAll      map[string]string
		ssReq      map[string]string
		tpReq      map[string]string
		tpOpt      map[string]string
		needUpdate bool
	}
	testCaseTest := func(c *testCase) {
		// all input keys should exist in optional
		updated := SyncStorageClassWithTypeAndService(c.scAll, c.ssReq, c.tpReq, c.tpOpt)
		switch {
		case c.needUpdate && updated == nil:
			t.Fatalf("test case <%s> should get an update", c.describe)
		case !c.needUpdate && updated != nil:
			t.Fatalf("test case <%s> got unexpected update, %v", c.describe, updated)
		default:
			t.Logf("test case <%s> done", c.describe)
		}
	}
	caseList := []testCase{
		{
			describe:   "all match",
			scAll:      map[string]string{"ka": "ta", "kb": "vb"},
			ssReq:      map[string]string{"ka": "ta"},
			tpReq:      map[string]string{"ka": "va"},
			tpOpt:      map[string]string{"kb": "tb"},
			needUpdate: false,
		}, // TODO more test
	}
	for i := range caseList {
		testCaseTest(&caseList[i])
	}
}
