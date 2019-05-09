package meta

import (
	"reflect"
	"testing"
)

func TestPodKindString(t *testing.T) {
	testCases := map[string]struct {
		kind     PodKind
		expected string
	}{
		"gc": {
			PodKindGC,
			"gc",
		},
	}

	var result string
	for d, tc := range testCases {
		result = tc.kind.String()
		if result != tc.expected {
			t.Errorf("Test case %s failed: expected %s, but got %s", d, tc.expected, result)
		}
	}
}

func TestLabelExistsSelector(t *testing.T) {
	testCases := map[string]struct {
		label    string
		expected string
	}{
		"tenant exists": {
			LabelTenantName,
			LabelTenantName,
		},
		"project exists": {
			LabelProjectName,
			LabelProjectName,
		},
	}

	var result string
	for d, tc := range testCases {
		result = LabelExistsSelector(tc.label)
		if result != tc.expected {
			t.Errorf("Test case %s failed: expected %s, but got %s", d, tc.expected, result)
		}
	}
}

func TestLabelExists(t *testing.T) {
	testCases := map[string]struct {
		labels         map[string]string
		expectedLabel  string
		expectedResult bool
	}{
		"empty labels": {
			nil,
			LabelProjectName,
			false,
		},
		"label exists": {
			map[string]string{
				LabelProjectName:  "devops",
				LabelWorkflowName: "test",
			},
			LabelProjectName,
			true,
		},
		"label does not exists": {
			map[string]string{
				LabelWorkflowName: "test",
			},
			LabelProjectName,
			false,
		},
	}

	var result bool
	for d, tc := range testCases {
		result = LabelExists(tc.labels, tc.expectedLabel)
		if result != tc.expectedResult {
			t.Errorf("Test case %s failed: expected %t, but got %t", d, tc.expectedResult, result)
		}
	}
}

func TestAddStageTemplateLabel(t *testing.T) {
	testCases := map[string]struct {
		labels   map[string]string
		expected map[string]string
	}{
		"nil labels": {
			nil,
			map[string]string{
				LabelStageTemplate: LabelValueTrue,
			},
		},
		"empty labels": {
			map[string]string{},
			map[string]string{
				LabelStageTemplate: LabelValueTrue,
			},
		},
		"non-empty labels": {
			map[string]string{
				LabelTenantName: "admin",
			},
			map[string]string{
				LabelTenantName:    "admin",
				LabelStageTemplate: LabelValueTrue,
			},
		},
	}

	var result map[string]string
	for d, tc := range testCases {
		result = AddStageTemplateLabel(tc.labels)
		if !reflect.DeepEqual(result, tc.expected) {
			t.Errorf("Test case %s failed: expected %s, but got %s", d, tc.expected, result)
		}
	}
}
