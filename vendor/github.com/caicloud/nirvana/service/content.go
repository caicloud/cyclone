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

package service

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
	"strconv"

	"github.com/caicloud/nirvana/definition"
)

// Consumer handles specifically typed data from a reader and unmarshals it into an object.
type Consumer interface {
	// ContentType returns a HTTP MIME type.
	ContentType() string
	// Consume unmarshals data from r into v.
	Consume(r io.Reader, v interface{}) error
}

// Producer marshals an object to specifically typed data and write it into a writer.
type Producer interface {
	// ContentType returns a HTTP MIME type.
	ContentType() string
	// Produce marshals v to data and write to w.
	Produce(w io.Writer, v interface{}) error
}

var consumers = map[string]Consumer{
	definition.MIMENone:        &NoneSerializer{},
	definition.MIMEText:        NewSimpleSerializer(definition.MIMEText),
	definition.MIMEJSON:        &JSONSerializer{},
	definition.MIMEXML:         &XMLSerializer{},
	definition.MIMEOctetStream: NewSimpleSerializer(definition.MIMEOctetStream),
	definition.MIMEURLEncoded:  &URLEncodedConsumer{},
	definition.MIMEFormData:    &FormDataConsumer{},
}

var producers = map[string]Producer{
	definition.MIMENone:        &NoneSerializer{},
	definition.MIMEText:        NewSimpleSerializer(definition.MIMEText),
	definition.MIMEJSON:        &JSONSerializer{},
	definition.MIMEXML:         &XMLSerializer{},
	definition.MIMEOctetStream: NewSimpleSerializer(definition.MIMEOctetStream),
}

// AllConsumers returns all consumers.
func AllConsumers() []Consumer {
	cs := make([]Consumer, 0, len(consumers))
	for _, c := range consumers {
		cs = append(cs, c)
	}
	return cs
}

// ConsumerFor gets a consumer for specified content type.
func ConsumerFor(contentType string) Consumer {
	return consumers[contentType]
}

// AllProducers returns all producers.
func AllProducers() []Producer {
	ps := make([]Producer, 0, len(producers))
	// JSON always the first one in producers.
	// The first one will be chosen when accept types
	// are not recognized.
	if p := producers[definition.MIMEJSON]; p != nil {
		ps = append(ps, p)
	}
	for _, p := range producers {
		if p.ContentType() == definition.MIMEJSON {
			continue
		}
		ps = append(ps, p)
	}
	return ps
}

// ProducerFor gets a producer for specified content type.
func ProducerFor(contentType string) Producer {
	return producers[contentType]
}

// RegisterConsumer register a consumer. A consumer must not handle "*/*".
func RegisterConsumer(c Consumer) error {
	if c.ContentType() == definition.MIMEAll {
		return invalidConsumer.Error(definition.MIMEAll)
	}
	consumers[c.ContentType()] = c
	return nil
}

// RegisterProducer register a producer. A producer must not handle "*/*".
func RegisterProducer(p Producer) error {
	if p.ContentType() == definition.MIMEAll {
		return invalidProducer.Error(definition.MIMEAll)
	}
	producers[p.ContentType()] = p
	return nil
}

// NoneSerializer implements Consumer and Producer for content types
// which can only receive data by io.Reader.
type NoneSerializer struct{}

// ContentType returns none MIME type.
func (s *NoneSerializer) ContentType() string {
	return definition.MIMENone
}

// Consume does nothing.
func (s *NoneSerializer) Consume(r io.Reader, v interface{}) error {
	return invalidTypeForConsumer.Error(s.ContentType(), reflect.TypeOf(v))
}

// Produce does nothing.
func (s *NoneSerializer) Produce(w io.Writer, v interface{}) error {
	return invalidTypeForProducer.Error(s.ContentType(), reflect.TypeOf(v))
}

// RawSerializer implements a raw serializer.
type RawSerializer struct{}

// CanConsumeData checks if raw serializer can consume type v with specified content type.
func (s *RawSerializer) CanConsumeData(contentType string, r io.Reader, v interface{}) bool {
	switch v.(type) {
	case *string, *[]byte:
		return true
	}
	return false
}

// ConsumeData reads data and converts it to string, []byte.
func (s *RawSerializer) ConsumeData(contentType string, r io.Reader, v interface{}) error {
	switch target := v.(type) {
	case *string:
		data, err := ioutil.ReadAll(r)
		if err != nil {
			return err
		}
		*target = string(data)
		return nil
	case *[]byte:
		data, err := ioutil.ReadAll(r)
		if err != nil {
			return err
		}
		*target = data
		return nil
	}
	return invalidTypeForConsumer.Error(contentType, reflect.TypeOf(v))
}

// CanProduceData checks if raw serializer can produce data for specified content type from type v.
func (s *RawSerializer) CanProduceData(contentType string, w io.Writer, v interface{}) bool {
	if _, ok := v.(io.Reader); ok {
		return true
	}
	switch v.(type) {
	case string, []byte:
		return true
	}
	return false
}

// ProduceData writes v to writer. v should be string, []byte, io.Reader.
func (s *RawSerializer) ProduceData(contentType string, w io.Writer, v interface{}) error {
	if r, ok := v.(io.Reader); ok {
		_, err := io.Copy(w, r)
		return err
	}
	switch source := v.(type) {
	case string:
		_, err := io.WriteString(w, source)
		return err
	case []byte:
		_, err := w.Write(source)
		return err
	}
	return invalidTypeForProducer.Error(contentType, reflect.TypeOf(v))
}

// SimpleSerializer implements a simple serializer.
type SimpleSerializer struct {
	RawSerializer
	contentType string
}

// NewSimpleSerializer creates a simple serializer.
func NewSimpleSerializer(contentType string) *SimpleSerializer {
	return &SimpleSerializer{
		contentType: contentType,
	}
}

// ContentType returns plain text MIME type.
func (s *SimpleSerializer) ContentType() string {
	return s.contentType
}

// Consume reads data and converts it to string, []byte.
func (s *SimpleSerializer) Consume(r io.Reader, v interface{}) error {
	return s.ConsumeData(s.ContentType(), r, v)
}

// Produce writes v to writer. v should be string, []byte, io.Reader.
func (s *SimpleSerializer) Produce(w io.Writer, v interface{}) error {
	if s.CanProduceData(s.ContentType(), w, v) {
		return s.ProduceData(s.ContentType(), w, v)
	}
	if r, ok := v.(error); ok {
		_, err := io.WriteString(w, r.Error())
		return err
	}
	if r, ok := v.(fmt.Stringer); ok {
		_, err := io.WriteString(w, r.String())
		return err
	}
	return invalidTypeForProducer.Error(s.ContentType(), reflect.TypeOf(v))
}

// URLEncodedConsumer implements Consumer for content type "application/x-www-form-urlencoded"
type URLEncodedConsumer struct{ RawSerializer }

// ContentType returns url encoded MIME type.
func (s *URLEncodedConsumer) ContentType() string {
	return definition.MIMEURLEncoded
}

// Consume reads data and converts it to string, []byte.
func (s *URLEncodedConsumer) Consume(r io.Reader, v interface{}) error {
	return s.ConsumeData(s.ContentType(), r, v)
}

// FormDataConsumer implements Consumer for content type "multipart/form-data"
type FormDataConsumer struct{ RawSerializer }

// ContentType returns form data MIME type.
func (s *FormDataConsumer) ContentType() string {
	return definition.MIMEFormData
}

// Consume reads data and converts it to string, []byte.
func (s *FormDataConsumer) Consume(r io.Reader, v interface{}) error {
	return s.ConsumeData(s.ContentType(), r, v)
}

// JSONSerializer implements Consumer and Producer for content type "application/json".
type JSONSerializer struct{ RawSerializer }

// ContentType returns json MIME type.
func (s *JSONSerializer) ContentType() string {
	return definition.MIMEJSON
}

// Consume unmarshals json from r into v.
func (s *JSONSerializer) Consume(r io.Reader, v interface{}) error {
	if s.CanConsumeData(s.ContentType(), r, v) {
		return s.ConsumeData(s.ContentType(), r, v)
	}
	err := json.NewDecoder(r).Decode(v)
	if err == io.EOF {
		return nil
	}
	return err
}

// Produce marshals v to json and write to w.
func (s *JSONSerializer) Produce(w io.Writer, v interface{}) error {
	if s.CanProduceData(s.ContentType(), w, v) {
		return s.ProduceData(s.ContentType(), w, v)
	}
	return json.NewEncoder(w).Encode(v)
}

// XMLSerializer implements Consumer and Producer for content type "application/xml".
type XMLSerializer struct{ RawSerializer }

// ContentType returns xml MIME type.
func (s *XMLSerializer) ContentType() string {
	return definition.MIMEXML
}

// Consume unmarshals xml from r into v.
func (s *XMLSerializer) Consume(r io.Reader, v interface{}) error {
	if s.CanConsumeData(s.ContentType(), r, v) {
		return s.ConsumeData(s.ContentType(), r, v)
	}
	err := xml.NewDecoder(r).Decode(v)
	if err == io.EOF {
		return nil
	}
	return err
}

// Produce marshals v to xml and write to w.
func (s *XMLSerializer) Produce(w io.Writer, v interface{}) error {
	if s.CanProduceData(s.ContentType(), w, v) {
		return s.ProduceData(s.ContentType(), w, v)
	}
	return xml.NewEncoder(w).Encode(v)
}

// Prefab creates instances for internal type. These instances are not
// unmarshaled form http request data.
type Prefab interface {
	// Name returns prefab name.
	Name() string
	// Type is instance type.
	Type() reflect.Type
	// Make makes an instance.
	Make(ctx context.Context) (interface{}, error)
}

var prefabs = map[string]Prefab{
	"context": &ContextPrefab{},
}

// PrefabFor gets a prefab by name.
func PrefabFor(name string) Prefab {
	return prefabs[name]
}

// RegisterPrefab registers a prefab.
func RegisterPrefab(prefab Prefab) error {
	prefabs[prefab.Name()] = prefab
	return nil
}

// ContextPrefab returns context from parameter of Make().
// It's usually used for generating the first parameter of api handler.
type ContextPrefab struct{}

// Name returns prefab name.
func (p *ContextPrefab) Name() string {
	return "context"
}

// Type is type of context.Context.
func (p *ContextPrefab) Type() reflect.Type {
	return reflect.TypeOf((*context.Context)(nil)).Elem()
}

// Make returns context simply.
func (p *ContextPrefab) Make(ctx context.Context) (interface{}, error) {
	return ctx, nil
}

// Converter is used to convert []string to specific type. Data must have one
// element at least or it will panic.
type Converter func(ctx context.Context, data []string) (interface{}, error)

var converters = map[reflect.Type]Converter{
	reflect.TypeOf(bool(false)):  ConvertToBool,
	reflect.TypeOf(int(0)):       ConvertToInt,
	reflect.TypeOf(int8(0)):      ConvertToInt8,
	reflect.TypeOf(int16(0)):     ConvertToInt16,
	reflect.TypeOf(int32(0)):     ConvertToInt32,
	reflect.TypeOf(int64(0)):     ConvertToInt64,
	reflect.TypeOf(uint(0)):      ConvertToUint,
	reflect.TypeOf(uint8(0)):     ConvertToUint8,
	reflect.TypeOf(uint16(0)):    ConvertToUint16,
	reflect.TypeOf(uint32(0)):    ConvertToUint32,
	reflect.TypeOf(uint64(0)):    ConvertToUint64,
	reflect.TypeOf(float32(0)):   ConvertToFloat32,
	reflect.TypeOf(float64(0)):   ConvertToFloat64,
	reflect.TypeOf(string("")):   ConvertToString,
	reflect.TypeOf(new(bool)):    ConvertToBoolP,
	reflect.TypeOf(new(int)):     ConvertToIntP,
	reflect.TypeOf(new(int8)):    ConvertToInt8P,
	reflect.TypeOf(new(int16)):   ConvertToInt16P,
	reflect.TypeOf(new(int32)):   ConvertToInt32P,
	reflect.TypeOf(new(int64)):   ConvertToInt64P,
	reflect.TypeOf(new(uint)):    ConvertToUintP,
	reflect.TypeOf(new(uint8)):   ConvertToUint8P,
	reflect.TypeOf(new(uint16)):  ConvertToUint16P,
	reflect.TypeOf(new(uint32)):  ConvertToUint32P,
	reflect.TypeOf(new(uint64)):  ConvertToUint64P,
	reflect.TypeOf(new(float32)): ConvertToFloat32P,
	reflect.TypeOf(new(float64)): ConvertToFloat64P,
	reflect.TypeOf(new(string)):  ConvertToStringP,
	reflect.TypeOf([]bool{}):     ConvertToBoolSlice,
	reflect.TypeOf([]int{}):      ConvertToIntSlice,
	reflect.TypeOf([]float64{}):  ConvertToFloat64Slice,
	reflect.TypeOf([]string{}):   ConvertToStringSlice,
}

// ConverterFor gets converter for specified type.
func ConverterFor(typ reflect.Type) Converter {
	return converters[typ]
}

// RegisterConverter registers a converter for specified type. New converter
// overrides old one.
func RegisterConverter(typ reflect.Type, converter Converter) {
	converters[typ] = converter
}

// ConvertToBool converts []string to bool. Only the first data is used.
func ConvertToBool(ctx context.Context, data []string) (interface{}, error) {
	origin := data[0]
	target, err := strconv.ParseBool(origin)
	if err != nil {
		return nil, invalidConversion.Error(origin, "bool")
	}
	return target, nil
}

// ConvertToBoolP converts []string to *bool. Only the first data is used.
func ConvertToBoolP(ctx context.Context, data []string) (interface{}, error) {
	ret, err := ConvertToBool(ctx, data)
	if err != nil {
		return nil, err
	}
	value := ret.(bool)
	return &value, nil
}

// ConvertToInt converts []string to int. Only the first data is used.
func ConvertToInt(ctx context.Context, data []string) (interface{}, error) {
	origin := data[0]
	target, err := strconv.ParseInt(origin, 10, 0)
	if err != nil {
		return nil, invalidConversion.Error(origin, "int")
	}
	return int(target), nil
}

// ConvertToIntP converts []string to *int. Only the first data is used.
func ConvertToIntP(ctx context.Context, data []string) (interface{}, error) {
	ret, err := ConvertToInt(ctx, data)
	if err != nil {
		return nil, err
	}
	value := ret.(int)
	return &value, nil
}

// ConvertToInt8 converts []string to int8. Only the first data is used.
func ConvertToInt8(ctx context.Context, data []string) (interface{}, error) {
	origin := data[0]
	target, err := strconv.ParseInt(origin, 10, 8)
	if err != nil {
		return nil, invalidConversion.Error(origin, "int8")
	}
	return int8(target), nil
}

// ConvertToInt8P converts []string to *int8. Only the first data is used.
func ConvertToInt8P(ctx context.Context, data []string) (interface{}, error) {
	ret, err := ConvertToInt8(ctx, data)
	if err != nil {
		return nil, err
	}
	value := ret.(int8)
	return &value, nil
}

// ConvertToInt16 converts []string to int16. Only the first data is used.
func ConvertToInt16(ctx context.Context, data []string) (interface{}, error) {
	origin := data[0]
	target, err := strconv.ParseInt(origin, 10, 16)
	if err != nil {
		return nil, invalidConversion.Error(origin, "int16")
	}
	return int16(target), nil
}

// ConvertToInt16P converts []string to *int16. Only the first data is used.
func ConvertToInt16P(ctx context.Context, data []string) (interface{}, error) {
	ret, err := ConvertToInt16(ctx, data)
	if err != nil {
		return nil, err
	}
	value := ret.(int16)
	return &value, nil
}

// ConvertToInt32 converts []string to int32. Only the first data is used.
func ConvertToInt32(ctx context.Context, data []string) (interface{}, error) {
	origin := data[0]
	target, err := strconv.ParseInt(origin, 10, 32)
	if err != nil {
		return nil, invalidConversion.Error(origin, "int32")
	}
	return int32(target), nil
}

// ConvertToInt32P converts []string to *int32. Only the first data is used.
func ConvertToInt32P(ctx context.Context, data []string) (interface{}, error) {
	ret, err := ConvertToInt32(ctx, data)
	if err != nil {
		return nil, err
	}
	value := ret.(int32)
	return &value, nil
}

// ConvertToInt64 converts []string to int64. Only the first data is used.
func ConvertToInt64(ctx context.Context, data []string) (interface{}, error) {
	origin := data[0]
	target, err := strconv.ParseInt(origin, 10, 64)
	if err != nil {
		return nil, invalidConversion.Error(origin, "int64")
	}
	return target, nil
}

// ConvertToInt64P converts []string to *int64. Only the first data is used.
func ConvertToInt64P(ctx context.Context, data []string) (interface{}, error) {
	ret, err := ConvertToInt64(ctx, data)
	if err != nil {
		return nil, err
	}
	value := ret.(int64)
	return &value, nil
}

// ConvertToUint converts []string to uint. Only the first data is used.
func ConvertToUint(ctx context.Context, data []string) (interface{}, error) {
	origin := data[0]
	target, err := strconv.ParseUint(origin, 10, 0)
	if err != nil {
		return nil, invalidConversion.Error(origin, "uint")
	}
	return uint(target), nil
}

// ConvertToUintP converts []string to *uint. Only the first data is used.
func ConvertToUintP(ctx context.Context, data []string) (interface{}, error) {
	ret, err := ConvertToUint(ctx, data)
	if err != nil {
		return nil, err
	}
	value := ret.(uint)
	return &value, nil
}

// ConvertToUint8 converts []string to uint8. Only the first data is used.
func ConvertToUint8(ctx context.Context, data []string) (interface{}, error) {
	origin := data[0]
	target, err := strconv.ParseUint(origin, 10, 8)
	if err != nil {
		return nil, invalidConversion.Error(origin, "uint8")
	}
	return uint8(target), nil
}

// ConvertToUint8P converts []string to *uint8. Only the first data is used.
func ConvertToUint8P(ctx context.Context, data []string) (interface{}, error) {
	ret, err := ConvertToUint8(ctx, data)
	if err != nil {
		return nil, err
	}
	value := ret.(uint8)
	return &value, nil
}

// ConvertToUint16 converts []string to uint16. Only the first data is used.
func ConvertToUint16(ctx context.Context, data []string) (interface{}, error) {
	origin := data[0]
	target, err := strconv.ParseUint(origin, 10, 16)
	if err != nil {
		return nil, invalidConversion.Error(origin, "uint16")
	}
	return uint16(target), nil
}

// ConvertToUint16P converts []string to *uint16. Only the first data is used.
func ConvertToUint16P(ctx context.Context, data []string) (interface{}, error) {
	ret, err := ConvertToUint16(ctx, data)
	if err != nil {
		return nil, err
	}
	value := ret.(uint16)
	return &value, nil
}

// ConvertToUint32 converts []string to uint32. Only the first data is used.
func ConvertToUint32(ctx context.Context, data []string) (interface{}, error) {
	origin := data[0]
	target, err := strconv.ParseUint(origin, 10, 32)
	if err != nil {
		return nil, invalidConversion.Error(origin, "uint32")
	}
	return uint32(target), nil
}

// ConvertToUint32P converts []string to *uint32. Only the first data is used.
func ConvertToUint32P(ctx context.Context, data []string) (interface{}, error) {
	ret, err := ConvertToUint32(ctx, data)
	if err != nil {
		return nil, err
	}
	value := ret.(uint32)
	return &value, nil
}

// ConvertToUint64 converts []string to uint64. Only the first data is used.
func ConvertToUint64(ctx context.Context, data []string) (interface{}, error) {
	origin := data[0]
	target, err := strconv.ParseUint(origin, 10, 64)
	if err != nil {
		return nil, invalidConversion.Error(origin, "uint64")
	}
	return target, nil
}

// ConvertToUint64P converts []string to *uint64. Only the first data is used.
func ConvertToUint64P(ctx context.Context, data []string) (interface{}, error) {
	ret, err := ConvertToUint64(ctx, data)
	if err != nil {
		return nil, err
	}
	value := ret.(uint64)
	return &value, nil
}

// ConvertToFloat32 converts []string to float32. Only the first data is used.
func ConvertToFloat32(ctx context.Context, data []string) (interface{}, error) {
	origin := data[0]
	target, err := strconv.ParseFloat(origin, 32)
	if err != nil {
		return nil, invalidConversion.Error(origin, "float32")
	}
	return float32(target), nil
}

// ConvertToFloat32P converts []string to *float32. Only the first data is used.
func ConvertToFloat32P(ctx context.Context, data []string) (interface{}, error) {
	ret, err := ConvertToFloat32(ctx, data)
	if err != nil {
		return nil, err
	}
	value := ret.(float32)
	return &value, nil
}

// ConvertToFloat64 converts []string to float64. Only the first data is used.
func ConvertToFloat64(ctx context.Context, data []string) (interface{}, error) {
	origin := data[0]
	target, err := strconv.ParseFloat(origin, 64)
	if err != nil {
		return nil, invalidConversion.Error(origin, "float64")
	}
	return target, nil
}

// ConvertToFloat64P converts []string to *float64. Only the first data is used.
func ConvertToFloat64P(ctx context.Context, data []string) (interface{}, error) {
	ret, err := ConvertToFloat64(ctx, data)
	if err != nil {
		return nil, err
	}
	value := ret.(float64)
	return &value, nil
}

// ConvertToString return the first element in []string.
func ConvertToString(ctx context.Context, data []string) (interface{}, error) {
	return data[0], nil
}

// ConvertToStringP return the first element's pointer in []string.
func ConvertToStringP(ctx context.Context, data []string) (interface{}, error) {
	return &data[0], nil
}

// ConvertToBoolSlice converts all elements in data to bool, and return []bool
func ConvertToBoolSlice(ctx context.Context, data []string) (interface{}, error) {
	ret := make([]bool, len(data))
	for i := range data {
		r, err := ConvertToBool(ctx, data[i:i+1])
		if err != nil {
			return nil, err
		}
		ret[i] = r.(bool)
	}
	return ret, nil
}

// ConvertToIntSlice converts all elements in data to int, and return []int
func ConvertToIntSlice(ctx context.Context, data []string) (interface{}, error) {
	ret := make([]int, len(data))
	for i := range data {
		r, err := ConvertToInt(ctx, data[i:i+1])
		if err != nil {
			return nil, err
		}
		ret[i] = r.(int)
	}
	return ret, nil
}

// ConvertToFloat64Slice converts all elements in data to float64, and return []float64
func ConvertToFloat64Slice(ctx context.Context, data []string) (interface{}, error) {
	ret := make([]float64, len(data))
	for i := range data {
		r, err := ConvertToFloat64(ctx, data[i:i+1])
		if err != nil {
			return nil, err
		}
		ret[i] = r.(float64)
	}
	return ret, nil
}

// ConvertToStringSlice return all strings in data.
func ConvertToStringSlice(ctx context.Context, data []string) (interface{}, error) {
	return data, nil
}
