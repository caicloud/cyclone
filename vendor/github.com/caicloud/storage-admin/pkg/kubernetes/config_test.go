package kubernetes

import (
	"fmt"
	"sort"
	"testing"
)

const (
	testMapSize = 3
)

func initTestKeys(n int) (keys []string) {
	keys = make([]string, n)
	for i := range keys {
		keys[i] = fmt.Sprintf("cluster-%03d", i)
	}
	return keys
}

func initTestMap(keys []string) map[string]Interface {
	m := make(map[string]Interface, len(keys))
	for _, key := range keys {
		m[key] = new(Clientset)
	}
	return m
}

func TestClusterClientsetMap_Range(t *testing.T) {
	var (
		clusters []string
		cluster  string
		kc       Interface
	)
	testKeys := initTestKeys(testMapSize)
	testMap := initTestMap(testKeys)
	ccg := NewClusterClientsetGetter(testMap)

	// test Range All and check correct
	if e := ccg.Range(func(k string, v Interface) error {
		if v != testMap[k] {
			t.Fatalf("check Range correct failed, cluster=%s, get %v, want %v", k, v, testMap[k])
		}
		clusters = append(clusters, k)
		return nil
	}); e != nil {
		t.Fatalf("check Range failed, get unexpect error %v", e)
	}
	if len(clusters) != testMapSize {
		t.Fatalf("check Range all failed, get %d, want %d", len(clusters), testMapSize)
	}
	sort.Strings(clusters)
	for i := range clusters {
		if clusters[i] != testKeys[i] {
			t.Fatalf("check Range failed, get %v, want %v", clusters[i], testKeys[i])
		}
	}
	t.Logf("test Range all done")

	// test stop
	stopKey := testKeys[testMapSize-1]
	testErr := fmt.Errorf("test error")
	if e := ccg.Range(func(k string, v Interface) error {
		if k == stopKey {
			cluster, kc = k, v
			return testErr
		}
		return nil
	}); e != testErr {
		t.Fatalf("test Range stop failed, get unexpected error %v", e)
	}
	if cluster != stopKey {
		t.Fatalf("test Range stop failed, get cluster=%s != %s", cluster, stopKey)
	}
	if kc != testMap[stopKey] {
		t.Fatalf("test Range stop failed, get client=%v != %v", kc, testMap[stopKey])
	}
	t.Logf("test Range stop done")
}
