package errors

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type ErrorReason string

/*
{
  // required, human readable, allow to use template
  "message": "name %(name)s is too short",
  // required for 400/404/409/422, hint for display, unique by error type in api group
  "reason": "admin.monitoring:CreateDashboardNameTooShort",
  // required for 400 when template is used in message or i18nMessage
  "data": {
    "name": "some-user-input-name"
  },
  ...
  // whatever else here will be logged in console for debug, but not for display
}
*/
type ApiError struct {
	Message string      `json:"message"`
	Reason  ErrorReason `json:"reason"`
	Data    interface{} `json:"data,omitempty"`
}
type FormatError struct {
	ApiError
	Code   int    `json:"httpCode,omitempty"`
	Raw    error  `json:"-"`
	ErrStr string `json:"error,omitempty"` // for json print, error can't be packed correctly
}

func NewError() *FormatError { return new(FormatError) }

func (fe *FormatError) Api() *ApiError {
	return &fe.ApiError
}

func (fe *FormatError) String() string {
	b, _ := json.Marshal(fe)
	return string(b)
}

func (fe *FormatError) Error() string {
	if fe.Raw != nil {
		return fe.Raw.Error()
	}
	return fe.String()
}

func (fe *FormatError) SetRawError(e error) {
	fe.Raw = e
	if e != nil {
		fe.ErrStr = e.Error()
	}
}

func (ae *ApiError) String() string {
	b, _ := json.Marshal(ae)
	return string(b)
}
func (ae *ApiError) Error() string {
	return ae.String()
}

func ParseResponseError(resp *http.Response) (*FormatError, error) {
	fe := new(FormatError)
	fe.Code = resp.StatusCode

	defer resp.Body.Close()
	b, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return nil, e
	}

	if e = json.Unmarshal(b, fe); e != nil {
		return nil, e
	}
	return fe, nil
}

func GetFormatError(e error) (*FormatError, bool) {
	fe, ok := e.(*FormatError)
	return fe, ok
}
