English | [中文](README_zh-CN.md)

<h1 align="center">
	<br>
	<img width="400" src="docs/images/logo.jpeg" alt="cyclone">
	<br>
	<br>
</h1>

[![Go Report Card](https://goreportcard.com/badge/github.com/caicloud/cyclone?style=flat-square)](https://goreportcard.com/report/github.com/caicloud/cyclone)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/caicloud/cyclone)
[![Gitter](https://img.shields.io/gitter/room/caicloud/cyclone.svg?style=flat-square)](https://gitter.im/caicloud/cyclone?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)

[![StackShare](https://img.shields.io/badge/tech-stack-0690fa.svg?style=flat-square)](https://stackshare.io/gaocegege/cyclone)
[![GitHub contributors](https://img.shields.io/github/contributors/caicloud/cyclone.svg?style=flat-square)](https://github.com/caicloud/cyclone/graphs/contributors)
[![Issue Stats](https://img.shields.io/issuestats/i/github/caicloud/cyclone.svg?style=flat-square)](https://github.com/caicloud/cyclone/issues)
[![Issue Stats](https://img.shields.io/issuestats/p/github/caicloud/cyclone.svg?style=flat-square)](https://github.com/caicloud/cyclone/pulls)

Unit testing:
[![Build Status](https://travis-ci.org/caicloud/cyclone.svg?branch=master)](https://travis-ci.org/caicloud/cyclone)
End-to-end testing:
![Build Status](https://img.shields.io/badge/e2e--test-comming%20soon-brightgreen.svg)

Cyclone is a cloud native CI/CD platform built for container workflow.

The primary directive of cyclone is to ship code from local development all the way to container engine of choice, either running in test or production environment. Features of cyclone includes:

- **Container Native**: every build, integration and deployment runs in container, completely excludes inconsistency between runtime environment
- **Dependency Aware**: define dependency rules, or simply component relationship, cyclone takes care of execution order as well as rollout strategy
- **Version Control**: cyclone is built with version control in mind; retrieving image/pipeline history is as simple as querying its version management interface
- **Two-way Binding**: cyclone records every CI/CD operation and its effect to answer questions like "how various container images are deployed across the fleet?"

## Documentation

### Setup Guide

To set up a cyclone instance, check out the [setup guide](./docs/setup.md) in the documentation.

### Quick Start

You could read the [quick start to start a tour of cyclone.](./docs/quick-start.md)

### Developer Guide

Feel free to hack on cyclone! We have [instructions to help you get started contributing.](./docs/developer-guide.md)

## Preview Feature

### Dependency Management

Please watch the [Fuctions Introduction](./docs/functions.md) for our features.

## Community

Welcome to join us for discussion and asking questions:
- Slack: caicloud-cyclone.slack.com, and you can use [this link](https://caicloud-cyclone.slack.com/join/signup) to signup.
- Email: zhujian@caicloud.io

## Roadmap

Please watch the [Github milestones](https://github.com/caicloud/cyclone/milestones) for our future plans.
