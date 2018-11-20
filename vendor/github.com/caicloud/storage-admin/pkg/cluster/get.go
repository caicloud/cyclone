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

func GetKubeClientFromClusterAdmin(host, clusterName string) (kubernetes.Interface, error) {
	addr := cleanUpGetAddress(host, clusterName)
	resp, e := http.Get(addr)
	if e != nil {
		return nil, e
	}
	return parseClusterGetResponse(resp)
}

func cleanUpGetAddress(clusterAdmAddr, clusterName string) string {
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
	substr := path.Join(RootPath, PathGetCluster, clusterName)
	if strings.Contains(addrBase, substr) {
		return scheme + "://" + path.Clean(addrBase)
	}
	return scheme + "://" + path.Join(addrBase, substr)
}

func parseClusterGetResponse(resp *http.Response) (kubernetes.Interface, error) {
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

	cir := new(cdsv1.ClusterInfoResponse)
	if e = json.Unmarshal(b, cir); e != nil {
		return nil, fmt.Errorf("parse json failed, %v, body:%v", e, string(b))
	}
	if len(cir.ErrorMessage) > 0 {
		return nil, fmt.Errorf("ErrorMessage: %v", cir.ErrorMessage)
	}

	return getClusterInfoKubeClt(&cir.Cluster)
}

func getClusterInfoKubeClt(cluster *cdsv1.ClusterInfo) (kubernetes.Interface, error) {
	apiServerAddr := "https://" + cluster.K8s.EndpointIp
	if len(cluster.K8s.EndpointPort) > 0 {
		apiServerAddr = apiServerAddr + ":" + cluster.K8s.EndpointPort
	}
	return kubernetes.NewClientFromUser(apiServerAddr, cluster.K8s.User, cluster.K8s.Password)
}
