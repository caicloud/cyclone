# validator 包

validator 包提供了一系列的方法帮助用户快速生成用于校验参数的 Operator：
```go
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

// NewCustom calls f for validation, using description for doc gen.
// User should only do custom validation in f.
// Validations which can be done by other way should be done in another Operator.
// Exp:
// []definition.Operator{NewCustom(f,"custom validation description")}
// f should be func(ctx context.Context, object AnyType) error
func NewCustom(f interface{}, description string) Validator

// Struct returns an operator to validate a structs exposed fields, and automatically validates nested structs, unless otherwise specified
// and also allows passing of context.Context for contextual validation information.
func Struct(instance interface{}) Validator

// String creates validator for string type.
func String(tag string) Validator

// Int creates validator for int type.
func Int(tag string) Validator

// Int64 creates validator for int64 type.
func Int64(tag string) Validator

// Int32 creates validator for int32 type.
func Int32(tag string) Validator

// Int16 creates validator for int16 type.
func Int16(tag string) Validator

// Int8 creates validator for int8 type.
func Int8(tag string) Validator

// Byte creates validator for byte type.
func Byte(tag string) Validator

// Uint creates validator for uint type.
func Uint(tag string) Validator

// Uint64 creates validator for uint64 type.
func Uint64(tag string) Validator

// Uint32 creates validator for uint32 type.
func Uint32(tag string) Validator

// Uint16 creates validator for uint16 type.
func Uint16(tag string) Validator

// Uint8 creates validator for uint8 type.
func Uint8(tag string) Validator

// Bool creates validator for bool type.
func Bool(tag string) Validator
```
目前支持三种类型的验证，分别对应 golang 基础类型，结构体类型和自定义类型。目前验证方法基于 `gopkg.in/go-playground/validator.v9`，如果希望使用其他工具，可以自定义一套类似的验证函数。
