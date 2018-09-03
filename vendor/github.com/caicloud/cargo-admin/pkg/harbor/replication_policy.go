package harbor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	. "github.com/caicloud/cargo-admin/pkg/errors"

	"github.com/caicloud/nirvana/log"

	"github.com/davecgh/go-spew/spew"
)

type ReplicationPolicyCenter interface {
	CreateReplicationPolicy(hcrpReq *HarborCreateRepPolicyReq) (int64, error)
	ListReplicationPolicies() ([]*HarborReplicationPolicy, error)
	GetReplicationPolicy(rpid int64) (*HarborCreateRepPolicyReq, error)
	DeleteReplicationPolicy(rpid int64) error
	TriggerReplicationPolicy(rpid int64) error
}

func (c *Client) CreateReplicationPolicy(hcrpReq *HarborCreateRepPolicyReq) (int64, error) {
	path := ReplicationPoliciesPath()
	b, err := json.Marshal(hcrpReq)
	if err != nil {
		return 0, ErrorUnknownInternal.Error(err)
	}

	log.Infof("%s %s", http.MethodPost, path)
	log.Infof("create replication policy: %s", spew.Sdump(hcrpReq))
	resp, err := c.do(http.MethodPost, path, bytes.NewReader(b))
	if err != nil {
		log.Errorf("send request error: %s", err)
		return 0, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, ErrorUnknownInternal.Error(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 201 {
		redirect := resp.Header.Get("Location")
		log.Infof("create replication plolicy success, redirect url in resp header: %s", redirect)
		rpidStr := strings.TrimPrefix(redirect, path+"/")
		rpid, err := strconv.ParseInt(rpidStr, 10, 64)
		log.Infof("rpidStr: %s, rpid: %d", rpidStr, rpid)
		if err != nil {
			log.Errorf("the returned replication policy Id can not be ParseInt: %v", err)
			return 0, ErrorUnknownInternal.Error(err)
		}
		log.Infof("create replication policy success, replication policy id: %d", rpid)
		return rpid, nil
	}
	log.Errorf("harbor return unexcepted statusCode: %d, resp body: %s", resp.StatusCode, body)

	return 0, ErrorUnknownInternal.Error(fmt.Sprintf("%s", body))
}

func (c *Client) ProjectReplicationPolicies(pid, page, pageSize int64) ([]*HarborReplicationPolicy, error) {
	return c.listReplication(ProjectReplicationPoliciesPath(pid, page, pageSize))
}

func (c *Client) ListReplicationPolicies() ([]*HarborReplicationPolicy, error) {
	return c.listReplication(ReplicationPoliciesPath())
}

func (c *Client) listReplication(path string) ([]*HarborReplicationPolicy, error) {
	log.Infof("%s %s", http.MethodGet, path)
	resp, err := c.do(http.MethodGet, path, nil)
	if err != nil {
		log.Errorf("send request error: %s", err)
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrorUnknownInternal.Error(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 == 2 {
		ret := make([]*HarborReplicationPolicy, 0)
		err := json.Unmarshal(body, &ret)
		if err != nil {
			return nil, ErrorUnknownInternal.Error(err)
		}
		return ret, nil
	}
	log.Errorf("harbor return unexcepted statusCode: %d, resp body: %s", resp.StatusCode, body)

	return nil, ErrorUnknownInternal.Error(body)
}

func (c *Client) GetReplicationPolicy(rpid int64) (*HarborReplicationPolicy, error) {
	path := ReplicationPolicyPath(rpid)

	log.Infof("%s %s", http.MethodGet, path)
	resp, err := c.do(http.MethodGet, path, nil)
	if err != nil {
		log.Errorf("send request error: %s", err)
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrorUnknownInternal.Error(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 == 2 {
		ret := &HarborReplicationPolicy{}
		err := json.Unmarshal(body, ret)
		if err != nil {
			return nil, ErrorUnknownInternal.Error(err)
		}
		return ret, nil
	}
	if resp.StatusCode == 404 {
		log.Errorf("not replication policy: %d in harbor", rpid)
		return nil, ErrorContentNotFound.Error(fmt.Sprintf("replication policy: %d", rpid))
	}
	log.Errorf("harbor return unexcepted statusCode: %d, resp body: %s", resp.StatusCode, body)

	return nil, ErrorUnknownInternal.Error(fmt.Sprintf("%s", body))
}

func (c *Client) UpdateReplicationPolicy(rpid int64, hurpReq *HarborUpdateRepPolicyReq) error {
	path := ReplicationPolicyPath(rpid)
	b, err := json.Marshal(hurpReq)
	if err != nil {
		return ErrorUnknownInternal.Error(err)
	}

	log.Infof("%s %s", http.MethodPut, path)
	log.Infof("update replication policy: %s", spew.Sdump(hurpReq))
	resp, err := c.do(http.MethodPut, path, bytes.NewReader(b))
	if err != nil {
		log.Errorf("send request error: %s", err)
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ErrorUnknownInternal.Error(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 == 2 {
		log.Infof("update replication plolicy: %d success", rpid)
		return nil
	}
	if resp.StatusCode == 404 {
		log.Errorf("not found replication policy: %d in harbor", rpid)
		return ErrorContentNotFound.Error(fmt.Sprintf("replication policy: %d", rpid))
	}
	log.Errorf("harbor return unexcepted statusCode: %d, resp body: %s", resp.StatusCode, body)

	return ErrorUnknownInternal.Error(fmt.Sprintf("%s", body))
}

func (c *Client) DeleteReplicationPolicy(rpid int64) error {
	path := ReplicationPolicyPath(rpid)

	log.Infof("%s %s", http.MethodDelete, path)
	resp, err := c.do(http.MethodDelete, path, nil)
	if err != nil {
		log.Errorf("send request error: %s", err)
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ErrorUnknownInternal.Error(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 == 2 {
		log.Infof("delete replication policy: %d successfully", rpid)
		return nil
	}
	if resp.StatusCode == 404 {
		log.Errorf("not found replication policy: %d in harbor", rpid)
		return ErrorContentNotFound.Error(fmt.Sprintf("replication policy: %d", rpid))
	}
	log.Errorf("harbor return unexcepted statusCode: %d, resp body: %s", resp.StatusCode, body)

	return ErrorUnknownInternal.Error(fmt.Sprintf("%s", body))
}

func (c *Client) TriggerReplicationPolicy(rpid int64) error {
	path := ReplicationsPath()

	log.Infof("%s %s", http.MethodPost, path)
	resp, err := c.do(http.MethodPost, path, strings.NewReader(fmt.Sprintf("{\"policy_id\": %d}", rpid)))
	if err != nil {
		return ErrorUnknownInternal.Error(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ErrorUnknownInternal.Error(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 == 2 {
		log.Infof("trigger replication policy: %d successfully", rpid)
		return nil
	}
	if resp.StatusCode == 404 {
		return ErrorContentNotFound.Error(fmt.Sprintf("replication policy: %d", rpid))
	}
	log.Errorf("harbor return unexpected statusCode: %d, resp body: %s", resp.StatusCode, body)

	return ErrorUnknownInternal.Error(fmt.Sprintf("%s", body))
}

func (c *Client) StopReplicationPolicyJobs(rpid int64) error {
	path := ReplicationPolicyPath(rpid)

	log.Infof("%s %s", http.MethodDelete, path)
	resp, err := c.do(http.MethodDelete, path, nil)
	if err != nil {
		log.Errorf("send request error: %s", err)
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ErrorUnknownInternal.Error(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 == 2 {
		log.Infof("delete replication policy: %d successfully", rpid)
		return nil
	}
	if resp.StatusCode == 404 {
		log.Errorf("not found replication policy: %d in harbor", rpid)
		return ErrorContentNotFound.Error(fmt.Sprintf("replication policy: %d", rpid))
	}
	log.Errorf("harbor return unexcepted statusCode: %d, resp body: %s", resp.StatusCode, body)

	return ErrorUnknownInternal.Error(fmt.Sprintf("%s", body))
}
