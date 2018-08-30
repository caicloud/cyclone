# converter 包

converter 包提供了一个 Converter 实现，帮助用户快速构建名为 converter 的 Operator：
```go
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
```
这个包非常简单，只是提供了一个方法帮助用户将转换函数生成为 Operator。
