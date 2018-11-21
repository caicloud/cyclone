/*
Copyright 2017 Caicloud Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package errors

// Builder can build error factories and errros.
type Builder interface {
	// Build builds a factory to generate errors with predefined format.
	Build(reason Reason, format string) Factory
	// Error immediately creates an error without reason.
	Error(format string, v ...interface{}) error
}

// kind maps to http code.
// And it can be used to make an error factory.
type kind int

// newKind creates a new builder with code.
func newKind(code int) Builder {
	return kind(code)
}

// Build builds a factory to generate errors with predefined format.
func (t kind) Build(reason Reason, format string) Factory {
	return &factory{code: int(t), reason: reason, format: format}
}

// Error immediately creates an error without reason.
func (t kind) Error(format string, v ...interface{}) error {
	msg, data := expand(format, v...)
	return &err{
		message: message{
			Message: msg,
			Data:    data,
		},
		factory: &factory{code: int(t)},
	}
}

// NewFactory creates a factory to create errors. The usage of this function is
// not recommended. Prefab kinds are more preferable.
func NewFactory(code int, reason Reason, format string) Factory {
	return newKind(code).Build(reason, format)
}

// These factory builders is used to build error factory.
var (
	BadRequest                   = newKind(400) // RFC 7231, 6.5.1
	Unauthorized                 = newKind(401) // RFC 7235, 3.1
	PaymentRequired              = newKind(402) // RFC 7231, 6.5.2
	Forbidden                    = newKind(403) // RFC 7231, 6.5.3
	NotFound                     = newKind(404) // RFC 7231, 6.5.4
	MethodNotAllowed             = newKind(405) // RFC 7231, 6.5.5
	NotAcceptable                = newKind(406) // RFC 7231, 6.5.6
	ProxyAuthRequired            = newKind(407) // RFC 7235, 3.2
	RequestTimeout               = newKind(408) // RFC 7231, 6.5.7
	Conflict                     = newKind(409) // RFC 7231, 6.5.8
	Gone                         = newKind(410) // RFC 7231, 6.5.9
	LengthRequired               = newKind(411) // RFC 7231, 6.5.10
	PreconditionFailed           = newKind(412) // RFC 7232, 4.2
	RequestEntityTooLarge        = newKind(413) // RFC 7231, 6.5.11
	RequestURITooLong            = newKind(414) // RFC 7231, 6.5.12
	UnsupportedMediaType         = newKind(415) // RFC 7231, 6.5.13
	RequestedRangeNotSatisfiable = newKind(416) // RFC 7233, 4.4
	ExpectationFailed            = newKind(417) // RFC 7231, 6.5.14
	Teapot                       = newKind(418) // RFC 7168, 2.3.3
	UnprocessableEntity          = newKind(422) // RFC 4918, 11.2
	Locked                       = newKind(423) // RFC 4918, 11.3
	FailedDependency             = newKind(424) // RFC 4918, 11.4
	UpgradeRequired              = newKind(426) // RFC 7231, 6.5.15
	PreconditionRequired         = newKind(428) // RFC 6585, 3
	TooManyRequests              = newKind(429) // RFC 6585, 4
	RequestHeaderFieldsTooLarge  = newKind(431) // RFC 6585, 5
	UnavailableForLegalReasons   = newKind(451) // RFC 7725, 3

	InternalServerError           = newKind(500) // RFC 7231, 6.6.1
	NotImplemented                = newKind(501) // RFC 7231, 6.6.2
	BadGateway                    = newKind(502) // RFC 7231, 6.6.3
	ServiceUnavailable            = newKind(503) // RFC 7231, 6.6.4
	GatewayTimeout                = newKind(504) // RFC 7231, 6.6.5
	HTTPVersionNotSupported       = newKind(505) // RFC 7231, 6.6.6
	VariantAlsoNegotiates         = newKind(506) // RFC 2295, 8.1
	InsufficientStorage           = newKind(507) // RFC 4918, 11.5
	LoopDetected                  = newKind(508) // RFC 5842, 7.2
	NotExtended                   = newKind(510) // RFC 2774, 7
	NetworkAuthenticationRequired = newKind(511) // RFC 6585, 6
)
