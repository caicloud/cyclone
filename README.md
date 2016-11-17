English | [中文](README_zh-CN.md)

<h1 align="center">
	<br>
	<img width="400" src="docs/logo.jpeg" alt="cyclone">
	<br>
	<br>
</h1>

[![Go Report Card](https://goreportcard.com/badge/github.com/caicloud/cyclone)](https://goreportcard.com/report/github.com/caicloud/cyclone)
[![GoDoc](https://godoc.org/github.com/caicloud/cyclone?status.svg)](https://godoc.org/github.com/caicloud/cyclone)
[![StackShare](https://img.shields.io/badge/tech-stack-0690fa.svg?style=flat)](https://stackshare.io/gaocegege/cyclone)
[![Gitter](https://badges.gitter.im/caicloud/cyclone.svg)](https://gitter.im/caicloud/cyclone?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)

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
- **Security First**: security is an essential part of cyclone; barriers can be setup to prevent insecure images from launching into production

## Documentation

### Setup Guide

To set up a cyclone instance, check out the [setup guide](./docs/setup.md) in the documentation.

### Quick Start

You could read the [quick start to start a tour of cyclone.](./docs/quick-start.md)

### Caicloud.yml Introduction

When `caicloud.yml` is given, version creation would include what you defined in the configuration file. you could find the [introduction to caicloud.yml](./docs/caicloud-yml-introduction.md) in the documentation.

### Developer Guide

Feel free to hack on cyclone! We have [instructions to help you get started contributing.](./docs/developer-guide.md)

## Preview Feature

### Dependency Management

<div align="center">
	<img src="./docs/dependency.png" alt="Dependency Management" width="500">
</div>

### Security Scanning

<div align="center">
	<img src="./docs/security.png" alt="Security Scanning" width="500">
</div>


## Roadmap

| Milestone | New feature                              |
| --------- | ---------------------------------------- |
| 2016 Q4   | Kubernetes Support: uses k8s-job to excute the CICD workflow |
| 2017 Q1   | Web UI                                   |
| Future    | User permission management <br>Supports user plugin |
