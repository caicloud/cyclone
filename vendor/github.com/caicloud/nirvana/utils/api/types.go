/*
Copyright 2018 Caicloud Authors

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

package api

import (
	"fmt"
	"go/types"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"unsafe"
)

// TypeName is unique name for go types.
type TypeName string

// TypeNameInvalid indicates an invalid type name.
const TypeNameInvalid = ""

// StructField describes a field of a struct.
type StructField struct {
	// Name is the field name.
	Name string
	// Comments of the type.
	Comments string
	// PkgPath is the package path that qualifies a lower case (unexported)
	// field name. It is empty for upper case (exported) field names.
	PkgPath string
	// Type is field type name.
	Type TypeName
	// Tag is field tag.
	Tag reflect.StructTag
	// Offset within struct, in bytes.
	Offset uintptr
	// Index sequence for Type.FieldByIndex.
	Index []int
	// Anonymous shows whether the field is an embedded field.
	Anonymous bool
}

// FuncField describes a field of function.
type FuncField struct {
	// Name is the field name.
	Name string
	// Type is field type name.
	Type TypeName
}

// Type describes an go type.
type Type struct {
	// Name is short type name.
	Name string
	// Comments of the type.
	Comments string
	// PkgPath is the package for this type.
	PkgPath string
	// Kind is type kind.
	Kind reflect.Kind
	// Key is map key type. Only used in map.
	Key TypeName
	// Elem is the element type of map, slice, array, pointer.
	Elem TypeName
	// Fields contains all struct fields of a struct.
	Fields []StructField
	// In presents fields of function input parameters.
	In []FuncField
	// Out presents fields of function output results.
	Out []FuncField
	// Conflict identifies the index of current type in a list of
	// types which have same type names. In most cases, this field is 0.
	Conflict int
}

// RawTypeName returns raw type name without confliction.
func (t *Type) RawTypeName() TypeName {
	if t.Name == "" {
		return TypeNameInvalid
	}
	if t.PkgPath == "" {
		return TypeName(t.Name)
	}
	return TypeName(t.PkgPath + "." + t.Name)
}

// TypeName returns type unique name.
func (t *Type) TypeName() TypeName {
	tn := t.RawTypeName()
	if tn != TypeNameInvalid && t.Conflict > 0 {
		tn = TypeName(fmt.Sprintf("%s [%d]", tn, t.Conflict))
	}
	return tn
}

// TypeContainer contains types.
type TypeContainer struct {
	lock  sync.RWMutex
	types map[TypeName]*Type
	real  map[TypeName][]reflect.Type
}

// NewTypeContainer creates a type container.
func NewTypeContainer() *TypeContainer {
	return &TypeContainer{
		types: make(map[TypeName]*Type),
		real:  make(map[TypeName][]reflect.Type),
	}
}

// Type gets type via type name.
func (tc *TypeContainer) Type(name TypeName) *Type {
	if name == TypeNameInvalid {
		return nil
	}
	tc.lock.RLock()
	defer tc.lock.RUnlock()
	return tc.types[name]
}

// Types returns all types
func (tc *TypeContainer) Types() map[TypeName]*Type {
	return tc.types
}

// setType sets type with unique name.
func (tc *TypeContainer) setType(t *Type, typ reflect.Type) {
	tn := t.RawTypeName()
	if tn == TypeNameInvalid {
		return
	}
	tc.lock.Lock()
	defer tc.lock.Unlock()
	types := tc.real[tn]
	index := len(types)
	for i, originalType := range types {
		if originalType == typ {
			index = i
		}
	}
	t.Conflict = index
	tn = t.TypeName()
	tc.types[tn] = t
	tc.real[tn] = append(types, typ)
}

// NameOf gets an unique name of a type.
func (tc *TypeContainer) NameOf(typ reflect.Type) TypeName {
	t := &Type{
		Name:    typ.Name(),
		PkgPath: typ.PkgPath(),
		Kind:    typ.Kind(),
	}
	if t.Name == "" && t.Kind == reflect.Interface {
		// Type interface{} has no name. Set it.
		t.Name = "interface{}"
	}
	tn := t.TypeName()
	if tn != TypeNameInvalid && tc.Type(tn) != nil {
		return tn
	}
	switch t.Kind {
	case reflect.Array, reflect.Slice:
		t.Elem = tc.NameOf(typ.Elem())
		if t.Name == "" {
			t.Name = fmt.Sprint("[]", t.Elem)
		}
	case reflect.Ptr:
		t.Elem = tc.NameOf(typ.Elem())
		if t.Name == "" {
			t.Name = fmt.Sprint("*", t.Elem)
		}
	case reflect.Map:
		t.Key = tc.NameOf(typ.Key())
		t.Elem = tc.NameOf(typ.Elem())
		if t.Name == "" {
			t.Name = fmt.Sprintf("map[%s]%s", t.Key, t.Elem)
		}
	case reflect.Chan:
		t.Elem = tc.NameOf(typ.Elem())
		if t.Name == "" {
			t.Name = fmt.Sprint("chan ", t.Elem)
		}
	case reflect.Struct:
		for i := 0; i < typ.NumField(); i++ {
			f := typ.Field(i)
			field := StructField{
				Name:      f.Name,
				PkgPath:   f.PkgPath,
				Tag:       f.Tag,
				Offset:    f.Offset,
				Index:     f.Index,
				Anonymous: f.Anonymous,
				Type:      tc.NameOf(f.Type),
			}
			t.Fields = append(t.Fields, field)
		}
	case reflect.Func:
		tc.fillFunctionSignature(t, typ)
		if t.Name == "" {
			t.Name = fmt.Sprintf("func(%s)", strings.Join(*(*[]string)(unsafe.Pointer(&t.In)), ", "))
			if len(t.Out) == 1 {
				t.Name += fmt.Sprintf(" %s", t.Out[0])
			} else if len(t.Out) > 1 {
				t.Name += fmt.Sprintf(" (%s)", strings.Join(*(*[]string)(unsafe.Pointer(&t.Out)), ", "))
			}
		}
	}
	tc.setType(t, typ)
	return t.TypeName()
}

func (tc *TypeContainer) fillFunctionSignature(t *Type, typ reflect.Type) {
	for i := 0; i < typ.NumIn(); i++ {
		t.In = append(t.In, FuncField{Type: tc.NameOf(typ.In(i))})
	}
	for i := 0; i < typ.NumOut(); i++ {
		t.Out = append(t.Out, FuncField{Type: tc.NameOf(typ.Out(i))})
	}
}

// NameOfInstance gets type name of an instance.
func (tc *TypeContainer) NameOfInstance(ins interface{}) TypeName {
	typ := reflect.TypeOf(ins)
	if typ.Kind() != reflect.Func {
		return tc.NameOf(typ)
	}
	funcInfo := runtime.FuncForPC(reflect.ValueOf(ins).Pointer())
	t := &Type{
		Kind: typ.Kind(),
		Name: funcInfo.Name(),
	}
	if index := strings.LastIndexByte(t.Name, '.'); index >= 0 {
		t.PkgPath = t.Name[:index]
		t.Name = t.Name[index+1:]
	}
	tc.fillFunctionSignature(t, typ)
	tc.setType(t, typ)
	return t.TypeName()
}

// Complete fills comments for all types.
func (tc *TypeContainer) Complete(analyzer *Analyzer) error {
	tc.lock.Lock()
	defer tc.lock.Unlock()
	errors := []error{}
	for _, typ := range tc.types {
		if typ.PkgPath == "" {
			continue
		}
		obj, err := analyzer.ObjectOf(typ.PkgPath, typ.Name)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		if comments := analyzer.Comments(obj.Pos()); comments != nil {
			typ.Comments = comments.Text()
		}
		switch typ.Kind {
		case reflect.Struct:
			o, ok := obj.Type().(*types.Named)
			if !ok {
				continue
			}
			st, ok := o.Underlying().(*types.Struct)
			if !ok {
				continue
			}
			for i := 0; i < st.NumFields(); i++ {
				field := st.Field(i)
				for j := 0; j < len(typ.Fields); j++ {
					if typ.Fields[j].Name == field.Name() {
						if comments := analyzer.Comments(field.Pos()); comments != nil {
							typ.Fields[j].Comments = comments.Text()
						}
						break
					}
				}
			}
		case reflect.Func:
			o, ok := obj.Type().(*types.Signature)
			if !ok {
				continue
			}
			for i := 0; i < o.Params().Len(); i++ {
				param := o.Params().At(i)
				typ.In[i].Name = param.Name()
			}
			for i := 0; i < o.Results().Len(); i++ {
				result := o.Results().At(i)
				typ.Out[i].Name = result.Name()
			}
		}
	}
	return fmt.Errorf("%v", errors)
}
