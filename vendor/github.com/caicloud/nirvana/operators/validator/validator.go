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

package validator

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	val "gopkg.in/go-playground/validator.v9"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/errors"
)

var std = val.New()

// OperatorKind means opeartor kind. All operators generated in this package
// are has kind `validator`.
const OperatorKind = "validator"

// Category distinguishs validation type based on different Validator implementation.
type Category string

const (
	// CategoryVar indicates that the validator can validate basic built-in types.
	// Types: string, int*, uint*, bool.
	CategoryVar Category = "Var"
	// CategoryStruct indicates that the validator can validate struct.
	CategoryStruct Category = "Struct"
	// CategoryCustom indicates the validator is a custom validator.
	CategoryCustom Category = "Custom"
)

// Validator describes an interface for all validator.
type Validator interface {
	definition.Operator
	// Category indicates validator type.
	Category() Category
	// Tag returns tag.
	Tag() string
	// Description returns description of current validator.
	Description() string
}

type validator struct {
	in          reflect.Type
	out         reflect.Type
	f           func(ctx context.Context, field string, object interface{}) (interface{}, error)
	category    Category
	tag         string
	description string
}

// Kind indicates operator type.
func (o *validator) Kind() string {
	return OperatorKind
}

// In returns the type of the only object parameter of operator.
func (o *validator) In() reflect.Type {
	return o.in
}

// Out returns the type of the only object result of operator.
func (o *validator) Out() reflect.Type {
	return o.out
}

// Operate operates an object and return one.
func (o *validator) Operate(ctx context.Context, field string, object interface{}) (interface{}, error) {
	return o.f(ctx, field, object)
}

// Category indicates validator type.
func (o *validator) Category() Category {
	return o.category
}

// Tag returns tag.
func (o *validator) Tag() string {
	return o.tag
}

// Description returns description of current validator.
func (o *validator) Description() string {
	return o.description
}

// NewCustom calls f for validation, using description for doc gen.
// User should only do custom validation in f.
// Validations which can be done by other way should be done in another Operator.
// Exp:
// []definition.Operator{NewCustom(f,"custom validation description")}
// f should be func(ctx context.Context, object AnyType) error
func NewCustom(f interface{}, description string) Validator {
	typ := reflect.TypeOf(f)
	if typ.Kind() != reflect.Func {
		panic(fmt.Sprintf("Parameter f in Custom Validator Func must be a function, but got: %s", typ.Kind()))
	}
	if typ.NumIn() != 2 {
		panic(fmt.Sprintf("Custom Validator Func must have 2 parameters, but got: %d", typ.NumIn()))
	}
	if typ.In(0) != reflect.TypeOf((*context.Context)(nil)).Elem() {
		panic(fmt.Sprintf("The first parameter of Custom Validator Func must be context.Context, but got: %v", typ.In(0)))
	}
	if typ.NumOut() != 1 {
		panic(fmt.Sprintf("Parameter f in Custom Validator Func must have two results, but got: %d", typ.NumOut()))
	}
	if typ.Out(0).String() != "error" {
		panic(fmt.Sprintf("The result of parameter f in Custom Validator Func must be error, but got: %v", typ.Out(1)))
	}
	op := definition.NewOperator(OperatorKind, typ.In(1), typ.In(1), func(ctx context.Context, _ string, object interface{}) (interface{}, error) {
		v := reflect.ValueOf(f).Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(object)})
		if v[0].IsNil() {
			return object, nil
		}
		return object, v[0].Interface().(error)
	})
	return &validator{
		in:          op.In(),
		out:         op.Out(),
		f:           op.Operate,
		category:    CategoryCustom,
		description: description,
	}
}

// Struct returns an operator to validate a structs exposed fields, and automatically validates nested structs, unless otherwise specified
// and also allows passing of context.Context for contextual validation information.
func Struct(instance interface{}) Validator {
	return &validator{
		in:  reflect.TypeOf(instance),
		out: reflect.TypeOf(instance),
		f: func(ctx context.Context, field string, object interface{}) (interface{}, error) {
			err := std.StructCtx(ctx, object)
			return object, decorateStructErr(err)
		},
		category: CategoryStruct,
	}
}

// String creates validator for string type.
func String(tag string) Validator {
	return varFor(tag, "")
}

// Int creates validator for int type.
func Int(tag string) Validator {
	return varFor(tag, int(0))
}

// Int64 creates validator for int64 type.
func Int64(tag string) Validator {
	return varFor(tag, int64(0))
}

// Int32 creates validator for int32 type.
func Int32(tag string) Validator {
	return varFor(tag, int32(0))
}

// Int16 creates validator for int16 type.
func Int16(tag string) Validator {
	return varFor(tag, int16(0))
}

// Int8 creates validator for int8 type.
func Int8(tag string) Validator {
	return varFor(tag, int8(0))
}

// Byte creates validator for byte type.
func Byte(tag string) Validator {
	return varFor(tag, byte(0))
}

// Uint creates validator for uint type.
func Uint(tag string) Validator {
	return varFor(tag, uint(0))
}

// Uint64 creates validator for uint64 type.
func Uint64(tag string) Validator {
	return varFor(tag, uint64(0))
}

// Uint32 creates validator for uint32 type.
func Uint32(tag string) Validator {
	return varFor(tag, uint32(0))
}

// Uint16 creates validator for uint16 type.
func Uint16(tag string) Validator {
	return varFor(tag, uint16(0))
}

// Uint8 creates validator for uint8 type.
func Uint8(tag string) Validator {
	return varFor(tag, uint8(0))
}

// Bool creates validator for bool type.
func Bool(tag string) Validator {
	return varFor(tag, bool(false))
}

// var returns an operator to validate a single variable using tag style validation and allows passing of contextual
// validation information via context.Context.
func varFor(tag string, instance interface{}) Validator {
	return &validator{
		in:  reflect.TypeOf(instance),
		out: reflect.TypeOf(instance),
		f: func(ctx context.Context, field string, object interface{}) (interface{}, error) {
			err := std.VarCtx(ctx, object, tag)
			return object, decorateErr(err, field, object, tag)
		},
		category: CategoryVar,
		tag:      tag,
	}
}

func decorateErr(err error, field string, object interface{}, tag string) error {
	if err == nil {
		return nil
	}
	return errors.BadRequest.Error("value '${value}' on query param '${field}' cannot pass validator tag '${tag}'", object, field, tag)
}

func decorateStructErr(err error) error {
	if err == nil {
		return nil
	}
	if err, ok := err.(val.ValidationErrors); ok {
		es := make([]string, 0, len(err))
		for _, fe := range err {
			es = append(es, fmt.Sprintf("value '%s' on struct field '%s' cannot pass validator tag '%s'", fe.Value(), fe.Field(), fe.Tag()))
		}
		return errors.BadRequest.Error("${err}", strings.Join(es, "; "))
	}

	return errors.BadRequest.Error("${err}", err)
}
