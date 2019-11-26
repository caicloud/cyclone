package gitea

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	api "code.gitea.io/gitea/modules/structs"
	"github.com/caicloud/nirvana/log"

	"github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/biz/scm"
)

// EventTypeHeader Gitea header event type
const EventTypeHeader = "X-Gitea-Event"

// ParseEvent parse the Gitea web hook message
func ParseEvent(scmCfg *v1alpha1.SCMSource, request *http.Request) (eventData *scm.EventData) {
	var err error

	var payload []byte
	if payload, err = ioutil.ReadAll(request.Body); err != nil {
		log.Errorln(err)
		return
	}

	eventKey := request.Header.Get(EventTypeHeader)
	switch eventKey {
	case string(EventCreate):
		var createPayload = new(api.CreatePayload)
		if err = json.Unmarshal(payload, createPayload); err != nil {
			log.Errorln(err)
			return
		}
		if createPayload.Repo == nil {
			log.Errorln("Gitea webhook event 'create' cannot get the repo name")
			return
		}
		eventData = &scm.EventData{
			Type: scm.PullRequestEventType,
			Repo: createPayload.Repo.FullName,
			Ref:  createPayload.Ref,
		}
		return
	case string(EventPR):
		var prPayload = new(api.PullRequestPayload)
		if err = json.Unmarshal(payload, prPayload); err != nil {
			log.Errorln(err)
			return
		}
		if prPayload.PullRequest == nil {
			log.Errorln("Gitea webhook event 'pull_request' cannot get the pull request repo info")
			return
		}
		eventData = &scm.EventData{
			Type: scm.PullRequestEventType,
			Repo: prPayload.Repository.FullName,
			Ref:  fmt.Sprintf("refs/pull/%d/head", prPayload.PullRequest.Index),
		}
		return
	case string(EventPush):
		var pushPayload = new(api.PushPayload)
		if err = json.Unmarshal(payload, pushPayload); err != nil {
			log.Errorln(err)
			return
		}
		if pushPayload.Repo == nil {
			log.Errorln("Gitea web hook event 'create' cannot get the repo name")
			return
		}
		eventData = &scm.EventData{
			Type: scm.PushEventType,
			Repo: pushPayload.Repo.FullName,
			Ref:  pushPayload.Ref,
		}
		return
	default:
		log.Warningln("Skip unsupported Gitea event")
	}
	return
}
