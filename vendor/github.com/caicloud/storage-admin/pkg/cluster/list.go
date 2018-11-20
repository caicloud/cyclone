package cluster

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"

	"github.com/golang/glog"

	cdsv1 "github.com/caicloud/storage-admin/pkg/cluster/apis/v1alpha1"
	"github.com/caicloud/storage-admin/pkg/kubernetes"
)

func GetKubeClientsFromClusterAdmin(host string) (map[string]kubernetes.Interface, error) {
	addr := cleanUpListAddress(host)
	resp, e := http.Get(addr)
	if e != nil {
		return nil, e
	}

	clr, e := parseClusterListResponse(resp)
	if e != nil {
		return nil, e
	}

	return getClusterMap(clr)
}

func cleanUpListAddress(clusterAdmAddr string) string {
	scheme := "http"
	addrBase := clusterAdmAddr
	schemes := []string{"http", "https"}
	for i := range schemes {
		prefix := schemes[i] + "://"
		if strings.HasPrefix(addrBase, prefix) {
			scheme = schemes[i]
			addrBase = addrBase[len(prefix):]
			break
		}
	}
	substr := path.Join(RootPath, PathListAllClusters)
	if strings.Contains(addrBase, substr) {
		return scheme + "://" + path.Clean(addrBase)
	}
	return scheme + "://" + path.Join(addrBase, substr)
}

func parseClusterListResponse(resp *http.Response) (*cdsv1.ClusterListResponse, error) {
	if resp == nil {
		return nil, fmt.Errorf("empty response")
	}

	defer resp.Body.Close()
	b, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return nil, fmt.Errorf("[httpcode:%d]read response body failed, %v", resp.StatusCode, e)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected http code %d, body:%v", resp.StatusCode, string(b))
	}
	glog.V(2).Info(string(b))

	clr := new(cdsv1.ClusterListResponse)
	if e = json.Unmarshal(b, clr); e != nil {
		return nil, fmt.Errorf("parse json failed, %v, body:%v", e, string(b))
	}
	if len(clr.ErrorMessage) > 0 {
		return nil, fmt.Errorf("ErrorMessage: %v", clr.ErrorMessage)
	}
	return clr, nil
}

func getClusterMap(clr *cdsv1.ClusterListResponse) (map[string]kubernetes.Interface, error) {
	if clr == nil {
		return nil, fmt.Errorf("empty ClusterListResponse")
	}
	m := make(map[string]kubernetes.Interface, len(clr.Clusters))

	for i := range clr.Clusters {
		cluster := &clr.Clusters[i]
		name := cluster.ClusterId
		kc, e := getClusterInfoKubeClt(cluster)
		if e != nil {
			return m, e
		}
		m[name] = kc
	}
	return m, nil
}
