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

package converter

import "github.com/caicloud/nirvana/definition"

// OperatorKind means opeartor kind. All operators generated in this package
// are has kind `converter`.
const OperatorKind = "converter"

// Converter describes a converter.
type Converter interface {
	definition.Operator
}

// For creates converter for a converter func.
//
// A converter func should has signature:
//  func f(context.Context, string, AnyType) (AnyType, error)
// The second parameter is a string that is used to generate error.
// AnyType can be any type in go. But struct type and
// built-in data type is recommended.
func For(f interface{}) Converter {
	return definition.OperatorFunc(OperatorKind, f)
}
