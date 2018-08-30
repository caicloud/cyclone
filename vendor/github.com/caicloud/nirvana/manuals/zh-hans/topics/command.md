# nirvana 命令

Nirvana 命令对应的包在 `cmd/nirvana` 中，目前包括三个命令：
1. init，用于初始化标准项目目录结构和必要文件
2. api，用于生成 API 文档（需要确保使用的是标准的项目结构，否则可能无法正常工作）
3. client，用于生成 API 对应的客户端（需要确保使用的是标准的项目结构，否则可能无法正常工作）。

每个命令都是一个目录，互相之间不干扰。每个目录都有一个 init.go 的文件用于把当前的命令加入到 Nirvana 根命令中，比如：
```go
package project

import "github.com/spf13/cobra"

// Register registers all commands.
func Register(root *cobra.Command) {
	root.AddCommand(newInitCommand())
}
```
然后在 main.go 中 import 这个包并进行命令注册：
```go
import (
	"github.com/caicloud/nirvana/cmd/nirvana/api"
	"github.com/caicloud/nirvana/cmd/nirvana/client"
	"github.com/caicloud/nirvana/cmd/nirvana/project"
	"github.com/caicloud/nirvana/log"
	"github.com/spf13/cobra"
)

var root = &cobra.Command{
	Use:   "nirvana",
	Short: "Nirvana toolchains",
}

func main() {
	project.Register(root)
	api.Register(root)
	client.Register(root)
	if err := root.Execute(); err != nil {
		log.Fatalln(err)
	}
}
```

接下来以 init 命令为例，说明单个命令的基本结构：
```go
func newInitCommand() *cobra.Command {
	options := &initOptions{}
	cmd := &cobra.Command{
		Use:   "init /path/to/project",
		Short: "Create a basic project structure",
		Long:  options.Manuals(),
		Run: func(cmd *cobra.Command, args []string) {
			if err := options.Validate(cmd, args); err != nil {
				log.Fatalln(err)
			}
			if err := options.Run(cmd, args); err != nil {
				log.Fatalln(err)
			}
		},
	}
	options.Install(cmd.PersistentFlags())
	return cmd
}

type initOptions struct {
	Boilerplate  string
	Version      string
	Registries   []string
	ImagePrefix  string
	ImageSuffix  string
	BuildImage   string
	RuntimeImage string
}

func (o *initOptions) Install(flags *pflag.FlagSet) {
	flags.StringVar(&o.Boilerplate, "boilerplate", "", "Path to boilerplate")
	flags.StringVar(&o.Version, "version", "v0.1.0", "First version of the project")
	flags.StringSliceVar(&o.Registries, "registries", []string{}, "Docker image registries")
	flags.StringVar(&o.ImagePrefix, "image-prefix", "", "Docker image prefix")
	flags.StringVar(&o.ImageSuffix, "image-suffix", "", "Docker image suffix")
	flags.StringVar(&o.BuildImage, "build-image", "golang:latest", "Golang image for building the project")
	flags.StringVar(&o.RuntimeImage, "runtime-image", "debian:jessie", "Docker base image for running the project")
}

func (o *initOptions) Validate(cmd *cobra.Command, args []string) error

func (o *initOptions) Run(cmd *cobra.Command, args []string) error
```
基本结构如下：
1. 一个创建命令的私有函数 newInitCommand
1. 一个表示当前命令的所有参数的 initOptions
  - Options 实现的 Install 方法用于安装 flag 到命令中
  - Options 实现的 Validate 方法用于验证参数是否正确
  - Options 实现的 Run 方法真正执行命令对应的逻辑

如果需要新增命令扩展 Nirvana 的功能，需要按照上述结构进行开发。
