/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package interceptor

import "testing"

func TestIsWorkspaceOperation(T *testing.T) {
	testData := map[string]bool{
		"/api/v1/workspaces":                                   true,
		"/api/v1/workspaces/":                                  true,
		"/api/v1/workspaces/abc":                               true,
		"/api/v1/workspaces/123":                               true,
		"/api/v1/workspaces/a*c1?&@":                           true,
		"/api/v1/workspaces/abc123/":                           true,
		"/api/v1/workspaces/abc123?start=0&limit=10&query=xyz": true,
		"/api/v2/workspaces":                                   false,
		"api/v1/workspaces":                                    false,
		"/api/v1/workspace":                                    false,
		"/api/v1/workspace/abc":                                false,
		"/api/v1/workspaces/abc123/pipelines":                  false,
		"/api/v1/workspaces/abc/123/pipelines":                 false,
	}

	for path, expected := range testData {
		result := IsWorkspaceOperation(path)
		if result != expected {
			T.Errorf("The result for path %s is %t; expected %t", path, result, expected)
		}
	}
}

func TestIsSpecifiedWorkspaceOperation(T *testing.T) {
	testData := map[string]bool{
		"/api/v1/workspaces/abc":                               true,
		"/api/v1/workspaces/123":                               true,
		"/api/v1/workspaces/a*c1?&@":                           true,
		"/api/v1/workspaces/abc123/":                           true,
		"/api/v1/workspaces/abc123?start=0&limit=10&query=xyz": true,
		"/api/v1/workspaces":                                   false,
		"/api/v1/workspaces/":                                  false,
		"/api/v2/workspaces":                                   false,
		"api/v1/workspaces":                                    false,
		"/api/v1/workspace":                                    false,
		"/api/v1/workspace/abc":                                false,
		"/api/v1/workspaces/abc123/pipelines":                  false,
		"/api/v1/workspaces/abc/123/pipelines":                 false,
	}

	for path, expected := range testData {
		result := IsSpecifiedWorkspaceOperation(path)
		if result != expected {
			T.Errorf("The result for path %s is %t; expected %t", path, result, expected)
		}
	}
}

func TestIsPipelineOperation(T *testing.T) {
	testData := map[string]bool{
		"/api/v1/workspaces/abc123/pipelines":                                      true,
		"/api/v1/workspaces/123/pipelines/":                                        true,
		"/api/v1/workspaces/abc123/pipelines?start=0&limit=10":                     true,
		"/api/v1/workspaces/a*c1?&@/pipelines/abc123":                              true,
		"/api/v1/workspaces/a*c1?&@/pipelines/c1*?%/":                              true,
		"/api/v1/workspaces/Abc/pipelines/abc/records?start=0&limit=10":            true,
		"/api/v1/workspaces/123Abc/pipelines/aBc012/records/":                      true,
		"/api/v1/workspaces/123Abc/pipelines/aBc012/records/1234":                  true,
		"/api/v1/workspaces/123Abc/pipelines/aBc012/records/1234/status":           true,
		"/api/v1/workspaces/123Abc/pipelines/aBc012/records/1234/Status":           false,
		"/api/v1/workspaces/123/Abc/pipelines/aBc012/records/1234":                 false,
		"/api/v1/workspaces/123Abc/pipelines/aBc/012/records/1234":                 false,
		"/api/v1/workspaces/123Abc/pipeline/aBc012/records/1234":                   false,
		"/api/v1/workspaces/123Abc/Pipeline/aBc012/records/1234":                   false,
		"/api/v1/workspaces/123Abc/pipelines/aBc012/record/1234":                   false,
		"/api/v1/workspaces/123Abc/pipelines/aBc012/Records/1234":                  false,
		"/api/v1/workspaces/abc/123/pipelines":                                     false,
		"/api/v1/workspaces/abc":                                                   false,
		"/api/v1/workspaces/123":                                                   false,
		"/api/v1/workspaces/a*c1?&@":                                               false,
		"/api/v1/workspaces/abc123/":                                               false,
		"/api/v1/workspaces/abc123?start=0&limit=10&query=xyz":                     false,
		"/api/v1/workspaces":                                                       false,
		"/api/v1/workspaces/":                                                      false,
		"/api/v2/workspaces":                                                       false,
		"api/v1/workspaces":                                                        false,
		"/api/v1/workspace":                                                        false,
		"/api/v1/workspace/abc":                                                    false,
	}

	for path, expected := range testData {
		result := IsPipelineOperation(path)
		if result != expected {
			T.Errorf("The result for path %s is %t; expected %t", path, result, expected)
		}
	}
}
