# Operator

包路径: `github.com/caicloud/nirvana/definition`

在每个 API Definition 中，都有一组 Parameters 和 Results。其中 Parameters 和业务函数的参数一一对应，而 Results 则和业务函数的返回值一一对应。一般情况下，如果我们需要对参数的合法性和结构进行验证和转换，那么就需要在业务函数中实现这些功能。为了让数据验证和结构转换等逻辑更加通用化和标准化，Nirvana 提供了 Operator 接口，可用于针对单个参数或返回值进行验证和修改：
```go
// Operator is used to operate an object and return an replacement object.
//
// For example:
//  A converter:
//    type ConverterForAnObject struct{}
//    func (c *ConverterForAnObject) Kind() {return "converter"}
//    func (c *ConverterForAnObject) In() reflect.Type {return definition.TypeOf(&ObjectV1{})}
//    func (c *ConverterForAnObject) Out() reflect.Type {return definition.TypeOf(&ObjectV2{})}
//    func (c *ConverterForAnObject) Operate(ctx context.Context, object interface{}) (interface{}, error) {
//        objV2, err := convertObjectV1ToObjectV2(object.(*ObjectV1))
//        return objV2, err
//    }
//
//  A validator:
//    type ValidatorForAnObject struct{}
//    func (c *ValidatorForAnObject) Kind() {return "validator"}
//    func (c *ValidatorForAnObject) In() reflect.Type {return definition.TypeOf(&Object{})}
//    func (c *ValidatorForAnObject) Out() reflect.Type {return definition.TypeOf(&Object{})}
//    func (c *ValidatorForAnObject) Operate(ctx context.Context, object interface{}) (interface{}, error) {
//        if err := validate(object.(*Object)); err != nil {
//            return nil, err
//        }
//        return object, nil
//    }
type Operator interface {
	// Kind indicates operator type.
	Kind() string
	// In returns the type of the only object parameter of operator.
	// The type must be a concrete struct or built-in type. It should
	// not be interface unless it's a file or stream reader.
	In() reflect.Type
	// Out returns the type of the only object result of operator.
	// The type must be a concrete struct or built-in type. It should
	// not be interface unless it's a file or stream reader.
	Out() reflect.Type
	// Operate operates an object and return one.
	Operate(ctx context.Context, field string, object interface{}) (interface{}, error)
}
```

在没有 Operator 的情况下，Nirvana 从业务函数的参数中确定数据类型，并将请求数据转换为这个类型。但是如果设置了 Operator，那么 Nirvana 会从第一个 Operator 的 `In()` 方法获取类型，
并且会检查最后一个 Operator 的 `Out()` 类型是否和业务函数的参数类型一致。

在实际的使用过程中，并不需要实现这个复杂的接口。Nirvana 提供了两种类型的 Operator：Validator 和 Converter。

## Validator

包路径: `github.com/caicloud/nirvana/operators/validator`

validator 包的实现基于 [go-playground/validator](https://godoc.org/gopkg.in/go-playground/validator.v9)，提供了用于生成 Operator 的方法：
```go
func Struct(instance interface{}) Validator
func String(tag string) Validator
func Int(tag string) Validator
func Int64(tag string) Validator
func Int32(tag string) Validator
func Int16(tag string) Validator
func Int8(tag string) Validator
func Byte(tag string) Validator
func Uint(tag string) Validator
func Uint64(tag string) Validator
func Uint32(tag string) Validator
func Uint16(tag string) Validator
func Uint8(tag string) Validator
func Bool(tag string) Validator
```
对于结构体类型，在需要的字段上添加名为 `validate` 的 tag。

### 自定义验证器
有时候默认的验证器不能覆盖复杂的验证需求，因此 validator 包还提供了方法用于创建自定义验证器：
```go
// NewCustom calls f for validation, using description for doc gen.
// User should only do custom validation in f.
// Validations which can be done by other way should be done in another Operator.
// Exp:
// []definition.Operator{NewCustom(f,"custom validation description")}
// f should be func(ctx context.Context, object AnyType) error
func NewCustom(f interface{}, description string) Validator
```
验证器函数必须符合签名 `func(ctx context.Context, object AnyType) error`。其中 AnyType 是具体要验证的类型，不能使用接口。

## Converter

包路径: `github.com/caicloud/nirvana/operators/converter`

除了对参数进行验证以外，在某些场景下还需要对参数进行类型转换。因此 converter 包提供了工具方法用于将转换函数包装成 Operator：
```go
// For creates converter for a converter func.
//
// A converter func should has signature:
//  func f(context.Context, string, AnyType) (AnyType, error)
// The second parameter is a string that is used to generate error.
// AnyType can be any type in go. But struct type and
// built-in data type is recommended.
func For(f interface{}) Converter
```
转换函数必须符合 `func f(context.Context, string, AnyType) (AnyType, error)`。其中参数的 AnyType 和返回值的 AnyType 可以不同。

## 在 Definition 中使用 Operator
这是一个在 List Messages 的 API 中添加 Operator 的示例：
```go
// Definition
var listMessages = def.Definition{
	Method:      def.List,
	Summary:     "List Messages",
	Description: "Query a specified number of messages and returns an array",
	Function:    message.ListMessages,
	Parameters: []def.Parameter{
		{
			Source:  def.Query,
			Name:    "count",
			Default: 10,
			Operators: []def.Operator{
				validator.Int("min=1"),
				converter.For(func(ctx context.Context, field string, value int) (uint, error) {
					return uint(value), nil
				}),
			},
			Description: "Number of messages",
		},
	},
	Results: def.DataErrorResults("A list of messages"),
}

// 业务函数
// ListMessages returns all messages.
func ListMessages(ctx context.Context, count uint) ([]Message, error) {
	messages := make([]Message, count)
	for i := 0; i < int(count); i++ {
		messages[i].ID = i
		messages[i].Title = fmt.Sprintf("%s %d", m.Title, i)
		messages[i].Content = fmt.Sprintf("%s %d", m.Content, i)
	}
	return messages, nil
}
```
这个例子中，验证器要求 count 的最小值为 1，并且把 int 类型转换为了 uint 类型。业务函数的参数也响应的变成了 uint 类型。

注意：Operator 是链式调用的，也就是说上一个 Operator 的返回值会作为下一个 Operator 的参数。因此如果将验证器和转换器的顺序调换，则需要将验证器的类型修改为转换器返回的类型：
```go
Operators: []def.Operator{
	converter.For(func(ctx context.Context, field string, value int) (uint, error) {
		return uint(value), nil
	}),
	validator.Uint("min=1"),
},
```
但是一般情况下，始终建议验证器放在前面，转换器放在后面。
