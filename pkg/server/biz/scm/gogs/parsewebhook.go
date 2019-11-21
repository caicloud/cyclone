package gogs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/caicloud/nirvana/log"
	gogs "github.com/gogs/go-gogs-client"

	"github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/biz/scm"
)

const EventTypeHeader = "X-Gogs-Event"

func ParseEvent(scmCfg *v1alpha1.SCMSource, request *http.Request) (eventData *scm.EventData) {
	var err error

	var payload []byte
	if payload, err = ioutil.ReadAll(request.Body); err != nil {
		log.Errorln(err)
		err = nil
		return
	}

	eventKey := request.Header.Get(EventTypeHeader)
	switch eventKey {
	case string(GogsEventCreate):
		var createPayload = new(gogs.CreatePayload)
		if err = json.Unmarshal(payload, createPayload); err != nil {
			log.Errorln(err)
			err = nil
			return
		}
		if createPayload.Repo == nil {
			log.Errorln("Gogs webhook event 'create' cannot get the repo name")
			err = nil
			return
		}
		eventData = &scm.EventData{
			Type: scm.PullRequestEventType,
			Repo: createPayload.Repo.FullName,
			Ref:  createPayload.Ref,
		}
		return
	case string(GogsEventPR):
		var prPayload = new(gogs.PullRequestPayload)
		if err = json.Unmarshal(payload, prPayload); err != nil {
			log.Errorln(err)
			err = nil
			return
		}
		if prPayload.PullRequest == nil {
			log.Errorln("Gogs webhook event 'pull_request' cannot get the pull request repo info")
			err = nil
			return
		}
		eventData = &scm.EventData{
			Type: scm.PullRequestEventType,
			Repo: prPayload.Repository.FullName,
			Ref:  fmt.Sprintf("refs/pull/%d/head", prPayload.PullRequest.Index),
		}
		return
	case string(GogsEventPush):
		var pushPayload = new(gogs.PushPayload)
		if err = json.Unmarshal(payload, pushPayload); err != nil {
			log.Errorln(err)
			err = nil
			return
		}
		if pushPayload.Repo == nil {
			log.Errorln("Gogs web hook event 'create' cannot get the repo name")
			err = nil
			return
		}
		eventData = &scm.EventData{
			Type: scm.PushEventType,
			Repo: pushPayload.Repo.FullName,
			Ref:  pushPayload.Ref,
		}
		return
	default:
		log.Warningln("Skip unsupported Github event")
	}
	return
}
