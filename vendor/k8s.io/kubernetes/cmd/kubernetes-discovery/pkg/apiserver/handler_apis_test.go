/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package apiserver

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"testing"

	"k8s.io/kubernetes/pkg/api"
	metav1 "k8s.io/kubernetes/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/client/cache"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/diff"

	"k8s.io/kubernetes/cmd/kubernetes-discovery/pkg/apis/apiregistration"
	listers "k8s.io/kubernetes/cmd/kubernetes-discovery/pkg/client/listers/apiregistration/internalversion"
)

type delegationHTTPHandler struct {
	called bool
}

func (d *delegationHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	d.called = true
	w.WriteHeader(http.StatusOK)
}

func TestAPIsDelegation(t *testing.T) {
	indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	delegate := &delegationHTTPHandler{}
	handler := &apisHandler{
		lister:   listers.NewAPIServiceLister(indexer),
		delegate: delegate,
	}

	server := httptest.NewServer(handler)
	defer server.Close()

	pathToDelegation := map[string]bool{
		"/":      true,
		"/apis":  false,
		"/apis/": false,
		"/apis/" + apiregistration.GroupName:                     true,
		"/apis/" + apiregistration.GroupName + "/":               true,
		"/apis/" + apiregistration.GroupName + "/anything":       true,
		"/apis/" + apiregistration.GroupName + "/anything/again": true,
		"/apis/something":                                        true,
		"/apis/something/nested":                                 true,
		"/apis/something/nested/deeper":                          true,
		"/api":     true,
		"/api/v1":  true,
		"/version": true,
	}

	for path, expectedDelegation := range pathToDelegation {
		delegate.called = false

		resp, err := http.Get(server.URL + path)
		if err != nil {
			t.Errorf("%s: %v", path, err)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			httputil.DumpResponse(resp, true)
			t.Errorf("%s: %v", path, err)
			continue
		}
		if e, a := expectedDelegation, delegate.called; e != a {
			t.Errorf("%s: expected %v, got %v", path, e, a)
			continue
		}
	}
}

func TestAPIs(t *testing.T) {
	tests := []struct {
		name        string
		apiservices []*apiregistration.APIService
		expected    *metav1.APIGroupList
	}{
		{
			name:        "empty",
			apiservices: []*apiregistration.APIService{},
			expected: &metav1.APIGroupList{
				TypeMeta: metav1.TypeMeta{Kind: "APIGroupList", APIVersion: "v1"},
				Groups: []metav1.APIGroup{
					discoveryGroup,
				},
			},
		},
		{
			name: "simple add",
			apiservices: []*apiregistration.APIService{
				{
					ObjectMeta: api.ObjectMeta{Name: "v1.foo"},
					Spec: apiregistration.APIServiceSpec{
						Group:    "foo",
						Version:  "v1",
						Priority: 10,
					},
				},
				{
					ObjectMeta: api.ObjectMeta{Name: "v1.bar"},
					Spec: apiregistration.APIServiceSpec{
						Group:    "bar",
						Version:  "v1",
						Priority: 11,
					},
				},
			},
			expected: &metav1.APIGroupList{
				TypeMeta: metav1.TypeMeta{Kind: "APIGroupList", APIVersion: "v1"},
				Groups: []metav1.APIGroup{
					discoveryGroup,
					{
						Name: "foo",
						Versions: []metav1.GroupVersionForDiscovery{
							{
								GroupVersion: "foo/v1",
								Version:      "v1",
							},
						},
						PreferredVersion: metav1.GroupVersionForDiscovery{
							GroupVersion: "foo/v1",
							Version:      "v1",
						},
					},
					{
						Name: "bar",
						Versions: []metav1.GroupVersionForDiscovery{
							{
								GroupVersion: "bar/v1",
								Version:      "v1",
							},
						},
						PreferredVersion: metav1.GroupVersionForDiscovery{
							GroupVersion: "bar/v1",
							Version:      "v1",
						},
					},
				},
			},
		},
		{
			name: "sorting",
			apiservices: []*apiregistration.APIService{
				{
					ObjectMeta: api.ObjectMeta{Name: "v1.foo"},
					Spec: apiregistration.APIServiceSpec{
						Group:    "foo",
						Version:  "v1",
						Priority: 20,
					},
				},
				{
					ObjectMeta: api.ObjectMeta{Name: "v2.bar"},
					Spec: apiregistration.APIServiceSpec{
						Group:    "bar",
						Version:  "v2",
						Priority: 11,
					},
				},
				{
					ObjectMeta: api.ObjectMeta{Name: "v2.foo"},
					Spec: apiregistration.APIServiceSpec{
						Group:    "foo",
						Version:  "v2",
						Priority: 1,
					},
				},
				{
					ObjectMeta: api.ObjectMeta{Name: "v1.bar"},
					Spec: apiregistration.APIServiceSpec{
						Group:    "bar",
						Version:  "v1",
						Priority: 11,
					},
				},
			},
			expected: &metav1.APIGroupList{
				TypeMeta: metav1.TypeMeta{Kind: "APIGroupList", APIVersion: "v1"},
				Groups: []metav1.APIGroup{
					discoveryGroup,
					{
						Name: "foo",
						Versions: []metav1.GroupVersionForDiscovery{
							{
								GroupVersion: "foo/v2",
								Version:      "v2",
							},
							{
								GroupVersion: "foo/v1",
								Version:      "v1",
							},
						},
						PreferredVersion: metav1.GroupVersionForDiscovery{
							GroupVersion: "foo/v2",
							Version:      "v2",
						},
					},
					{
						Name: "bar",
						Versions: []metav1.GroupVersionForDiscovery{
							{
								GroupVersion: "bar/v1",
								Version:      "v1",
							},
							{
								GroupVersion: "bar/v2",
								Version:      "v2",
							},
						},
						PreferredVersion: metav1.GroupVersionForDiscovery{
							GroupVersion: "bar/v1",
							Version:      "v1",
						},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
		delegate := &delegationHTTPHandler{}
		handler := &apisHandler{
			lister:   listers.NewAPIServiceLister(indexer),
			delegate: delegate,
		}
		for _, o := range tc.apiservices {
			indexer.Add(o)
		}

		server := httptest.NewServer(handler)
		defer server.Close()

		resp, err := http.Get(server.URL + "/apis")
		if err != nil {
			t.Errorf("%s: %v", tc.name, err)
			continue
		}
		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("%s: %v", tc.name, err)
			continue
		}

		actual := &metav1.APIGroupList{}
		if err := runtime.DecodeInto(api.Codecs.UniversalDecoder(), bytes, actual); err != nil {
			t.Errorf("%s: %v", tc.name, err)
			continue
		}
		if !api.Semantic.DeepEqual(tc.expected, actual) {
			t.Errorf("%s: %v", tc.name, diff.ObjectDiff(tc.expected, actual))
			continue
		}
	}
}

func TestAPIGroupMissing(t *testing.T) {
	indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	handler := &apiGroupHandler{
		lister:    listers.NewAPIServiceLister(indexer),
		groupName: "foo",
	}

	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL + "/apis/groupName/foo")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected %v, got %v", resp.StatusCode, http.StatusNotFound)
	}

	// foo still has no api services for it (like it was deleted), it should 404
	resp, err = http.Get(server.URL + "/apis/groupName/")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected %v, got %v", resp.StatusCode, http.StatusNotFound)
	}
}

func TestAPIGroup(t *testing.T) {
	tests := []struct {
		name        string
		group       string
		apiservices []*apiregistration.APIService
		expected    *metav1.APIGroup
	}{
		{
			name:  "sorting",
			group: "foo",
			apiservices: []*apiregistration.APIService{
				{
					ObjectMeta: api.ObjectMeta{Name: "v1.foo"},
					Spec: apiregistration.APIServiceSpec{
						Group:    "foo",
						Version:  "v1",
						Priority: 20,
					},
				},
				{
					ObjectMeta: api.ObjectMeta{Name: "v2.bar"},
					Spec: apiregistration.APIServiceSpec{
						Group:    "bar",
						Version:  "v2",
						Priority: 11,
					},
				},
				{
					ObjectMeta: api.ObjectMeta{Name: "v2.foo"},
					Spec: apiregistration.APIServiceSpec{
						Group:    "foo",
						Version:  "v2",
						Priority: 1,
					},
				},
				{
					ObjectMeta: api.ObjectMeta{Name: "v1.bar"},
					Spec: apiregistration.APIServiceSpec{
						Group:    "bar",
						Version:  "v1",
						Priority: 11,
					},
				},
			},
			expected: &metav1.APIGroup{
				TypeMeta: metav1.TypeMeta{Kind: "APIGroup", APIVersion: "v1"},
				Name:     "foo",
				Versions: []metav1.GroupVersionForDiscovery{
					{
						GroupVersion: "foo/v2",
						Version:      "v2",
					},
					{
						GroupVersion: "foo/v1",
						Version:      "v1",
					},
				},
				PreferredVersion: metav1.GroupVersionForDiscovery{
					GroupVersion: "foo/v2",
					Version:      "v2",
				},
			},
		},
	}

	for _, tc := range tests {
		indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
		handler := &apiGroupHandler{
			lister:    listers.NewAPIServiceLister(indexer),
			groupName: "foo",
		}
		for _, o := range tc.apiservices {
			indexer.Add(o)
		}

		server := httptest.NewServer(handler)
		defer server.Close()

		resp, err := http.Get(server.URL + "/apis/" + tc.group)
		if err != nil {
			t.Errorf("%s: %v", tc.name, err)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			httputil.DumpResponse(resp, true)
			t.Errorf("%s", tc.name)
			continue
		}
		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("%s: %v", tc.name, err)
			continue
		}

		actual := &metav1.APIGroup{}
		if err := runtime.DecodeInto(api.Codecs.UniversalDecoder(), bytes, actual); err != nil {
			t.Errorf("%s: %v", tc.name, err)
			continue
		}
		if !api.Semantic.DeepEqual(tc.expected, actual) {
			t.Errorf("%s: %v", tc.name, diff.ObjectDiff(tc.expected, actual))
			continue
		}
	}
}
