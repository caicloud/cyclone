# Nirvana

<img align="right" width="225px" src="https://user-images.githubusercontent.com/2191361/35839723-e9e5cdfa-0b2c-11e8-853a-8d3870f9e7ac.png">

[![Build Status](https://travis-ci.org/caicloud/nirvana.svg?branch=master)](https://travis-ci.org/caicloud/nirvana)
[![Coverage Status](https://coveralls.io/repos/github/caicloud/nirvana/badge.svg?branch=master)](https://coveralls.io/github/caicloud/nirvana?branch=master)
[![GoDoc](http://godoc.org/github.com/caicloud/nirvana?status.svg)](http://godoc.org/github.com/caicloud/nirvana)
[![Go Report Card](https://goreportcard.com/badge/github.com/caicloud/nirvana)](https://goreportcard.com/report/github.com/caicloud/nirvana)
[![OpenTracing Badge](https://img.shields.io/badge/OpenTracing-enabled-blue.svg)](http://opentracing.io)

Nirvana is a golang API framework designed for productivity and usability. It aims to be the building block for
all golang services at Caicloud. The high-level goals and features include:

* consistent API behavior, structure and layout across all golang projects
* improve engineering productivity with openAPI and client generation, etc
* validation can be added by declaring validation method as part of API definition
* out-of-box instrumentation support, e.g. metrics, profiling, tracing, etc
* easy and standard configuration management, as well as standard cli interface

Nirvana is also extensible and performant, with the goal to support fast developmenet velocity.

## Getting Started

Nirvana provides documentations in two languages to help you expore this framework. Note right now, only Chinese
docs are kept up-to-date.

- [中文](https://caicloud.github.io/nirvana/zh-hans)
- [English (Expired)](https://caicloud.github.io/nirvana/en)

## Features

- API Framework based on Descriptors
- Request Filter
- Middleware
- Validator
- Plugins
- API Doc Generation
- Client Generation

## Contributing

If you are interested in contributing to Nirvana, please checkout [CONTRIBUTING.md](./CONTRIBUTING.md).
We welcome any code or non-code contribution!

## Licensing

Nirvana is licensed under the Apache License, Version 2.0. See [LICENSE](./LICENSE) for the full license text.
