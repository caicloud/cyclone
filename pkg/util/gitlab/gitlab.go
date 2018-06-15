package gitlab

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/xanzy/go-gitlab"
)

const (
	// gitlabEventTypeHeader represents the Gitlab header key used to pass the event type.
	gitlabEventTypeHeader = "X-Gitlab-Event"

	NoteHookEvent         = "Note Hook"
	MergeRequestHookEvent = "Merge Request Hook"
	TagPushHookEvent      = "Tag Push Hook"
	PushHookEvent         = "Push Hook"
)

// ParseWebHook parses the body from webhook requeset.
func ParseWebHook(r *http.Request) (payload interface{}, err error) {
	eventType := r.Header.Get(gitlabEventTypeHeader)
	switch eventType {
	case NoteHookEvent:
		//payload = &gitlab.MergeCommentEvent{}
		// can not unmarshal request body to gitlab.MergeCommentEvent{}
		// due to gitlab.MergeCommentEvent.MergeRequest.CreatedAt's type(*time.Time),
		// parsing time "2018-05-31 02:19:38 UTC" as "2006-01-02T15:04:05Z07:00" will fail.
		payload = &MergeCommentEvent{}
	case MergeRequestHookEvent:
		payload = &gitlab.MergeEvent{}
	case TagPushHookEvent:
		payload = &gitlab.TagEvent{}
	case PushHookEvent:
		payload = &gitlab.PushEvent{}
	default:
		return nil, fmt.Errorf("event type %v not support", eventType)
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("Fail to read request body")
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	return payload, nil
}
