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

const (
	DefaultTargetType     = 0
	DefaultTargetInsecure = false
)

type TargetClienter interface {
	CreateTarget(name, endpoint, username, password string) (int64, error)
	ListTargets() ([]*HarborRepTarget, error)
	GetTarget(tid int64) (*HarborRepTarget, error)
	UpdateTarget(tid int64, name, url, username, password string) error
	DeleteTarget(tid int64) error
}

func (c *Client) CreateTarget(name, endpoint, username, password string) (int64, error) {
	path := TargetsPath()
	hctReq := &HarborCreateTargetReq{
		URL:      endpoint,
		Name:     name,
		Username: username,
		Password: password,
		Type:     DefaultTargetType,
		Insecure: DefaultTargetInsecure,
	}
	b, err := json.Marshal(hctReq)
	if err != nil {
		return 0, ErrorUnknownInternal.Error(err)
	}

	log.Infof("%s %s", http.MethodPost, path)
	log.Infof("create target: %s", spew.Sdump(hctReq))
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
		log.Infof("create target success, redirect url in resp header: %s", redirect)
		tidStr := strings.TrimPrefix(redirect, path+"/")
		tid, err := strconv.ParseInt(tidStr, 10, 64)
		log.Infof("tidStr: %s, tid: %d", tidStr, tid)
		if err != nil {
			log.Errorf("the returned targetId can not be ParseInt: %v", err)
			return 0, ErrorUnknownInternal.Error(err)
		}
		log.Infof("create target success, targetId: %d", tid)
		return tid, nil
	}

	if resp.StatusCode == 409 {
		log.Errorf("statusCode 409, harbor already the target, reponse body: %s, request body: %s", body, spew.Sdump(hctReq))
		return 0, ErrorAlreadyExist.Error(fmt.Sprintf("target %v", name))
	}
	log.Errorf("harbor return unexcepted statusCode: %d, resp body: %s", resp.StatusCode, body)

	return 0, ErrorUnknownInternal.Error(fmt.Sprintf("%s", body))
}

func (c *Client) ListTargets() ([]*HarborRepTarget, error) {
	path := TargetsPath()

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
		ret := make([]*HarborRepTarget, 0)
		err := json.Unmarshal(body, &ret)
		if err != nil {
			return nil, ErrorUnknownInternal.Error(err)
		}
		return ret, nil
	}
	log.Errorf("harbor return unexpected statusCode: %d, resp body: %s", resp.StatusCode, body)

	return nil, ErrorUnknownInternal.Error(body)
}

func (c *Client) GetTarget(tid int64) (*HarborRepTarget, error) {
	path := TargetPath(tid)

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
		ret := &HarborRepTarget{}
		err := json.Unmarshal(body, ret)
		if err != nil {
			return nil, ErrorUnknownInternal.Error(err)
		}
		return ret, nil
	}
	if resp.StatusCode == 404 {
		log.Errorf("not found target: %d in harbor", tid)
		return nil, ErrorContentNotFound.Error(fmt.Sprintf("target: %d", tid))
	}
	log.Errorf("harbor return unexcepted statusCode: %d, resp body: %s", resp.StatusCode, body)

	return nil, ErrorUnknownInternal.Error(fmt.Sprintf("%s", body))
}

func (c *Client) UpdateTarget(targetID int64, name, url, username, password string) error {
	path := TargetPath(targetID)
	hutReq := &HarborUpdateTargetReq{
		Name:     name,
		Username: username,
		Password: password,
		Endpoint: url,
		Insecure: DefaultTargetInsecure,
	}
	b, err := json.Marshal(hutReq)
	if err != nil {
		return ErrorUnknownInternal.Error(err)
	}

	log.Infof("%s %s", http.MethodPost, path)
	log.Infof("update target: %s", spew.Sdump(hutReq))
	resp, err := c.do(http.MethodPost, path, bytes.NewReader(b))
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
		log.Infof("update target: %d successfully", targetID)
		return nil
	}
	if resp.StatusCode == 404 {
		log.Errorf("not found target: %d in harbor", targetID)
		return ErrorContentNotFound.Error(fmt.Sprintf("target: %d", targetID))
	}
	log.Errorf("harbor return unexpected statusCode: %d, resp body: %s", resp.StatusCode, body)

	return ErrorUnknownInternal.Error(fmt.Sprintf("%s", body))
}

func (c *Client) DeleteTarget(tid int64) error {
	path := TargetPath(tid)

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
		log.Infof("delete target: %d successfully", tid)
		return nil
	}
	if resp.StatusCode == 404 {
		log.Errorf("not found target: %d in harbor", tid)
		return ErrorContentNotFound.Error(fmt.Sprintf("target: %d", tid))
	}
	log.Errorf("harbor return unexcepted statusCode: %d, resp body: %s", resp.StatusCode, body)

	return ErrorUnknownInternal.Error(fmt.Sprintf("%s", body))
}
