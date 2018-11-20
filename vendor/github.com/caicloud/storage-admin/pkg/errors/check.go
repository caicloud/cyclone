package errors

import (
	"encoding/json"
	"strings"
)

func IsObjectNotFound(fe *FormatError) bool {
	return fe != nil && fe.Reason == ErrorReasonObjectNotFound
}

func ParseQuotaNotCompleteApiErr(e error) *ApiError {
	if e == nil {
		return nil
	}
	ae := new(ApiError)
	if je := json.Unmarshal([]byte(e.Error()), ae); je != nil {
		return nil
	}
	if ae.Reason != ErrorReasonQuotaNotComplete {
		return nil
	}
	return ae
}

func IsQuotaExceeded(e error) bool {
	return e != nil && strings.Contains(e.Error(), "exceeded quota")
}
