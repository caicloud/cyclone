中文 | [English](README.md)
# Cyclone
![logo](docs/logo.jpeg)

[![Build Status](https://travis-ci.org/caicloud/cyclone.svg?branch=master)](https://travis-ci.org/caicloud/cyclone)
[![GoDoc](https://godoc.org/github.com/caicloud/cyclone?status.svg)](https://godoc.org/github.com/caicloud/cyclone)
[![Go Report Card](https://goreportcard.com/badge/github.com/caicloud/cyclone)](https://goreportcard.com/report/github.com/caicloud/cyclone)
[![Gitter](https://badges.gitter.im/caicloud/cyclone.svg)](https://gitter.im/caicloud/cyclone?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)

Cyclone是一个打造容器工作流的云原生持续集成持续发布平台。

Cyclone主要致力于将代码从本地开发环境用任意容器引擎封装搬运到测试或者生产环境运行。Cyclone包括一下特性：

- **容器原生**: 每次构建、集成、部署均在容器中运行，完全解决运行时环境不一致的问题。

- **依赖关系**: 定义依赖规则或简单的组件关系，确保执行顺序依照既定策略。

![dependency](docs/dependency.png)

- **版本控制**: 基于版本控制构建，检索镜像／流水线历史就像查询版本管理接口一样简单。

- **双向绑定**: 记录每次CI／CD操作用于回答类似问题：“各容器镜像部署在集群哪个角落？”

- **安全第一**: 安全是基本要素，有效阻拦不安全镜像进入生产环境。

![security](docs/security.png)

## Documentation
* [安装手册](./docs/setup_zh-CN.md)
* [快速开始](./docs/quick-start_zh-CN.md)
* [caicloud.yml简介](./docs/caicloud-yml-introduction_zh-CN.md)
* [API手册](http://118.193.142.27:7099/apidocs/)
* [原理](./docs/principle_zh-CN.md)
