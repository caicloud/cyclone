# utils 系列包

utils 系列包包含：
1. api 包，用于读取项目源码，并生成与 API 有关的所有数据结构，产出的结构可用于生成文档和客户端。
1. builder 包，利用 api 包读取源码并生成 API 数据。
1. generators/golang 包，利用 api 包提供的数据结构生成 golang 客户端。
1. generators/swagger 包，利用 api 包提供的数据结构生成 API 文档。
1. generators/utils 包，提供公共工具给其他生成器使用。
1. printer 包，提供了一个在 Terminal 中打印表格的功能。
1. project 包，提供了基础工具用于读取项目配置文件，通常是 nirvana.yaml。

除了 printer 包以外，其他包都是用于生成文档和客户端用的。

在 api 包中，提供了如下功能：
1. 对应 golang type 的 Type  
  为了能让 golang type 能转换为可读的数据结构，构建了 Type 相关类型：
  ```go
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
  ```
1. 对应 Nirvana API 的 Definition  
  此处的 Definition 大部分字段与 definition 包中的相同，之所以不复用之前的结构，是因为有部分类型相关的字段稍有不同。
  ```go
  // Parameter describes a function parameter.
  type Parameter struct {
  	// Source is the parameter value generated from.
  	Source definition.Source
  	// Name is the name to get value from a request.
  	Name string
  	// Description describes the parameter.
  	Description string
  	// Type is parameter object type.
  	Type TypeName
  	// Default is encoded default value.
  	Default []byte
  }
  
  // Result describes a function result.
  type Result struct {
  	// Destination is the target for the result.
  	Destination definition.Destination
  	// Description describes the result.
  	Description string
  	// Type is result object type.
  	Type TypeName
  }
  
  // Example is just an example.
  type Example struct {
  	// Description describes the example.
  	Description string
  	// Type is result object type.
  	Type TypeName
  	// Instance is encoded instance data.
  	Instance []byte
  }
  
  // Definition is complete version of def.Definition.
  type Definition struct {
  	// Method is definition method.
  	Method definition.Method
  	// HTTPMethod is http method.
  	HTTPMethod string
  	// HTTPCode is http success code.
  	HTTPCode int
  	// Summary is a brief of this definition.
  	Summary string
  	// Description describes the API handler.
  	Description string
  	// Consumes indicates how many content types the handler can consume.
  	// It will override parent descriptor's consumes.
  	Consumes []string
  	// Produces indicates how many content types the handler can produce.
  	// It will override parent descriptor's produces.
  	Produces []string
  	// ErrorProduces is used to generate data for error. If this field is empty,
  	// it means that this field equals to Produces.
  	// In some cases, succeessful data and error data should be generated in
  	// different ways.
  	ErrorProduces []string
  	// Function is a function handler. It must be func type.
  	Function TypeName
  	// Parameters describes function parameters.
  	Parameters []Parameter
  	// Results describes function retrun values.
  	Results []Result
  	// Examples contains many examples for the API handler.
  	Examples []Example
  }
  ```
1. 用于表示代码注释的 Comments  
  在源码解析的时候，通常需要通过注释实现一些特殊功能，因此要对注释进行分析。目前注释特殊格式是 `+nirvana:api=option:"value"`。
  ```go
  const (
  	// CommentsOptionDescriptors is the option name of descriptors.
  	CommentsOptionDescriptors = "descriptors"
  	// CommentsOptionModifiers is the option name of modifiers.
  	CommentsOptionModifiers = "modifiers"
  	// CommentsOptionAlias is the option name of alias.
  	CommentsOptionAlias = "alias"
  	// CommentsOptionOrigin is the option name of original name.
  	CommentsOptionOrigin = "origin"
  )
  
  // Comments is parsed from go comments.
  type Comments struct {
  	lines   []string
  	options map[string][]string
  }
  
  var optionsRegexp = regexp.MustCompile(`^[ \t]*\+nirvana:api[ \t]*=(.*)$`)
  var options = []string{CommentsOptionDescriptors, CommentsOptionModifiers, CommentsOptionAlias}
  
  // ParseComments parses comments and extracts nirvana options.
  func ParseComments(comments string) *Comments
  ```
1. 用于分析源码的 Analyzer  
  Analyzer 可以读取源码，获取结构对象和注释信息。
  ```go
  // Analyzer analyzes go packages.
  type Analyzer struct {
  	...
  }
  
  // NewAnalyzer creates a code ananlyzer.
  func NewAnalyzer(root string) *Analyzer
  
  // Import imports a package and all packages it depends on.
  func (a *Analyzer) Import(path string) (*types.Package, error)
  
  // PackageComments returns comments above package keyword.
  // Import package before calling this method.
  func (a *Analyzer) PackageComments(path string) []*ast.CommentGroup
   
  // Packages returns packages under specified directory (including itself).
  // Import package before calling this method.
  func (a *Analyzer) Packages(parent string, vendor bool) []string
  
  // FindPackages returns packages which contain target.
  // Import package before calling this method.
  func (a *Analyzer) FindPackages(target string) []string
   
  // Comments returns immediate comments above pos.
  // Import package before calling this method.
  func (a *Analyzer) Comments(pos token.Pos) *ast.CommentGroup
  
  // ObjectOf returns declaration object of target.
  func (a *Analyzer) ObjectOf(pkg, name string) (types.Object, error)
  ```
1. 集合上述所有功能的 Container  
  Container 读取源码并进行分析，产出 API 相关的所有定义和类型信息。API 定义和类型信息可以用来生成 API 文档和客户端。
  ```go
  // Definitions describes all APIs and its related object types.
  type Definitions struct {
  	// Definitions holds mappings between path and API descriptions.
  	Definitions map[string][]Definition
  	// Types contains all types used by definitions.
  	Types map[TypeName]*Type
  }
  
  // Container contains informations to generate APIs.
  type Container struct {
  	...
  }
  
  // NewContainer creates API container.
  func NewContainer(root string) *Container
  
  // AddModifier add definition modifiers to container.
  func (ac *Container) AddModifier(modifiers ...service.DefinitionModifier)
  
  // AddDescriptor add descriptors to container.
  func (ac *Container) AddDescriptor(descriptors ...definition.Descriptor)
   
  // Generate generates API definitions.
  func (ac *Container) Generate() (*Definitions, error)
  ```


builder 包相对 api 包来说就简单很多了，这个包里包含一个 API Builder：
```go
// APIBuilder builds api definitions by specified package.
type APIBuilder struct {
	...
}

// NewAPIBuilder creates an api builder.
func NewAPIBuilder(root string, paths ...string) *APIBuilder

// Build build api definitions.
func (b *APIBuilder) Build() (*api.Definitions, error)
```
API Builder 首先会利用 Analyzer 去读取指定路径的源码，然后从中找到标记了 modifiers 和 descriptors 选项的注释。两个选项的值对应两个函数，分别返回 Modifier 和 Descriptor。然后动态生成一个 main.go 文件 import 这两个函数对应包，然后调用这两个函数，取得返回值。得到返回值之后则利用 api 包的 Container 来生成 API 定义和类型信息。API Builder 执行 main.go 后通过 stdout 取得返回值，反序列化成 Definitions 结构，然后再返回给 API Builder 的调用者。这样就完成了对一个项目的 API 信息的提取。

golang 包和 swagger 包实际上都是利用了 API Builder 的返回结果，构建出相应的客户端和文档。golang 包会为每一个 API 生成一个客户端函数，这个函数的结构和服务读基本一致。函数体的实现则是直接使用了 rest 包 Client 接口。swagger 包则是利用了 `github.com/go-openapi/spec` 将 API 定义和类型转换成了 openapi 的定义。

