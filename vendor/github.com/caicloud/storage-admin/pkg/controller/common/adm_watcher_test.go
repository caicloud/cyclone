package common

import (
	"fmt"
	"testing"

	"github.com/caicloud/storage-admin/pkg/constants"
	"github.com/caicloud/storage-admin/pkg/kubernetes"
	"github.com/caicloud/storage-admin/pkg/kubernetes/fake"
	tt "github.com/caicloud/storage-admin/pkg/util/testtools"
)

var (
	clusters      = []string{"aaa", "bbb", "ccc"}
	masters       = []string{"https://1.1.1.1:443", "https://2.2.2.2:443", "https://3.3.3.3:443"}
	tokens        = []string{"t1111111", "t2222222", "t3333333"}
	users         = []string{"u1111111", "u2222222", "u3333333"}
	passwords     = []string{"p1111111", "p2222222", "p3333333"}
	fakeTokenCfgs = []kubernetes.Config{
		{Host: masters[0], BearerToken: tokens[0]},
		{Host: masters[1], BearerToken: tokens[1]},
		{Host: masters[2], BearerToken: tokens[2]},
	}
	fakeUserPwdCfgs = []kubernetes.Config{
		{Host: masters[0], Username: users[0], Password: passwords[0]},
		{Host: masters[1], Username: users[1], Password: passwords[1]},
		{Host: masters[2], Username: users[2], Password: passwords[2]},
	}
)

func getFakeClients(cfgs []kubernetes.Config) []kubernetes.Interface {
	ks := make([]kubernetes.Interface, len(cfgs))
	for i := range ks {
		ks[i] = fake.NewSimpleClientset().SetRestConfig(&cfgs[i])
	}
	return ks
}

func newFakeTokenCltMap() map[string]kubernetes.Interface {
	m := make(map[string]kubernetes.Interface, len(clusters))
	ks := getFakeClients(fakeTokenCfgs)
	for i := range clusters {
		m[clusters[i]] = ks[i]
	}
	return m
}
func newFakeUserPwdCltMap() map[string]kubernetes.Interface {
	m := make(map[string]kubernetes.Interface, len(clusters))
	ks := getFakeClients(fakeUserPwdCfgs)
	for i := range clusters {
		m[clusters[i]] = ks[i]
	}
	return m
}

func newLocalWatcher(m map[string]kubernetes.Interface) *Watcher {
	w := &Watcher{
		ccg:                kubernetes.NewClusterClientsetGetter(m),
		clusterAdmAddr:     "",
		clusterAdmWatchSec: constants.DefaultWatchIntervalSecond,
	}
	w.SetProcessors(w.ProcessUpdateListSimple, w.ProcessAddListSimple, w.ProcessDelListSimple)
	return w
}

func TestGetClusterClientsetMapDiff(t *testing.T) {
	type testCase struct {
		describe   string
		prev, next map[string]kubernetes.Interface
		addList    []string
		delList    []string
		updateList []string
	}
	defaultTkMap := newFakeTokenCltMap()
	defaultUpMap := newFakeUserPwdCltMap()
	testCaseTest := func(tc *testCase) {
		logPath := fmt.Sprintf("test case <%s>", tc.describe)
		addList, delList, updateList := GetClusterClientsetMapDiff(tc.prev, tc.next)
		if !tt.IsStringsSame(tc.addList, addList) {
			t.Fatalf("%s addList not same: %v != %v", logPath, tc.addList, addList)
		}
		if !tt.IsStringsSame(tc.delList, delList) {
			t.Fatalf("%s delList not same: %v != %v", logPath, tc.delList, delList)
		}
		if !tt.IsStringsSame(tc.updateList, updateList) {
			t.Fatalf("%s updateList not same: %v != %v", logPath, tc.updateList, updateList)
		}
		t.Logf("%s done", logPath)
	}
	testCases := []testCase{
		{
			describe: "token all same",
			prev:     defaultTkMap,
			next:     defaultTkMap,
		}, {
			describe: "user/pwd all same",
			prev:     defaultUpMap,
			next:     defaultUpMap,
		}, {
			describe: "add 1",
			prev:     copyMapAndDelKey(defaultTkMap, clusters[0]),
			next:     defaultTkMap,
			addList:  []string{clusters[0]},
		}, {
			describe: "del 1",
			prev:     defaultUpMap,
			next:     copyMapAndDelKey(defaultUpMap, clusters[0]),
			delList:  []string{clusters[0]},
		}, {
			describe:   "token update 1",
			prev:       defaultTkMap,
			next:       copyMapAndUpdateKey(defaultTkMap, clusters[0]),
			updateList: []string{clusters[0]},
		}, {
			describe:   "user/pwd update 1",
			prev:       defaultUpMap,
			next:       copyMapAndUpdateKey(defaultUpMap, clusters[0]),
			updateList: []string{clusters[0]},
		},
	}
	for i := range testCases {
		testCaseTest(&testCases[i])
	}
}

func Test_updateClusterInfo(t *testing.T) {
	type testCase struct {
		describe   string
		prev, next map[string]kubernetes.Interface
	}
	testCaseTest := func(tc *testCase) {
		logPath := fmt.Sprintf("test case <%s>", tc.describe)
		w := newLocalWatcher(tc.prev)
		w.updateClusterInfo(tc.next)

		var newList []string
		w.ccg.Range(func(cluster string, kc kubernetes.Interface) error {
			v, ok := tc.next[cluster]
			if !ok || v == nil {
				t.Fatalf("%s delete failed for cluster: %v", logPath, cluster)
			}
			if !kubernetes.IsConfigSame(kc.RestConfig(), v.RestConfig()) {
				// maybe pointer change for outer interface
				t.Fatalf("%s update failed for cluster: %v", logPath, cluster)
			}
			newList = append(newList, cluster)
			return nil
		})

		for _, cluster := range newList {
			v, ok := tc.next[cluster]
			if !ok || v == nil {
				t.Fatalf("%s add failed for cluster: %v", logPath, cluster)
			}
		}
		t.Logf("%s done", logPath)
	}
	defaultTkMap := newFakeTokenCltMap()
	defaultUpMap := newFakeUserPwdCltMap()
	testCases := []testCase{
		{
			describe: "token all same",
			prev:     defaultTkMap,
			next:     defaultTkMap,
		}, {
			describe: "user/pwd all same",
			prev:     defaultUpMap,
			next:     defaultUpMap,
		}, {
			describe: "add 1",
			prev:     copyMapAndDelKey(defaultTkMap, clusters[0]),
			next:     defaultTkMap,
		}, {
			describe: "del 1",
			prev:     defaultUpMap,
			next:     copyMapAndDelKey(defaultUpMap, clusters[0]),
		}, {
			describe: "token update 1",
			prev:     defaultTkMap,
			next:     copyMapAndUpdateKey(defaultTkMap, clusters[0]),
		}, {
			describe: "user/pwd update 1",
			prev:     defaultUpMap,
			next:     copyMapAndUpdateKey(defaultUpMap, clusters[0]),
		}, {
			describe: "update all by mode",
			prev:     defaultUpMap,
			next:     defaultTkMap,
		},
	}
	for i := range testCases {
		testCaseTest(&testCases[i])
	}
}

func copyMap(m map[string]kubernetes.Interface) map[string]kubernetes.Interface {
	nm := make(map[string]kubernetes.Interface, len(m))
	for k, v := range m {
		ncfg := *v.RestConfig()
		nm[k] = fake.NewSimpleClientset().SetRestConfig(&ncfg)
	}
	return nm
}
func copyMapAndDelKey(m map[string]kubernetes.Interface, key string) map[string]kubernetes.Interface {
	nm := copyMap(m)
	delete(nm, key)
	return nm
}
func copyMapAndUpdateKey(m map[string]kubernetes.Interface, key string) map[string]kubernetes.Interface {
	nm := copyMap(m)
	cfg := nm[key].RestConfig()
	if len(cfg.BearerToken) > 0 {
		cfg.BearerToken += "-xxx"
	}
	if len(cfg.Username) > 0 {
		cfg.Username += "-xxx"
	}
	if len(cfg.Password) > 0 {
		cfg.Password += "-xxx"
	}
	return nm
}
