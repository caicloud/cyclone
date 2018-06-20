<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [功能介绍](#%E5%8A%9F%E8%83%BD%E4%BB%8B%E7%BB%8D)
  - [统一管理](#%E7%BB%9F%E4%B8%80%E7%AE%A1%E7%90%86)
  - [持续集成](#%E6%8C%81%E7%BB%AD%E9%9B%86%E6%88%90)
  - [自动触发](#%E8%87%AA%E5%8A%A8%E8%A7%A6%E5%8F%91)
  - [实时日志](#%E5%AE%9E%E6%97%B6%E6%97%A5%E5%BF%97)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# 功能介绍

## 统一管理
<!--## 关联软件配置管理系统 -->

Cyclone 支持通过创建项目关联多种软件配置管理系统（GitHub、GitLab、SVN）来统一管理一组流水线。

- 创建与软件配置管理系统关联的项目

<div align="center">
	<img src="./images/create_project.png" alt="create_project" width="500">
</div>

- 列出所有项目

<div align="center">
	<img src="./images/list_projects.png" alt="list_projects" width="500">
</div>

## 持续集成

全过程可视化的工作流水线："codeCheckout"从指定仓库拉去代码，"package" 编译可执行文件，"imageBuild" 构建发布镜像， "integration" 集成测试，"imageRelease" 发布镜像。所有过程均以容器为载体，消除环境差异。

- 检出代码

<div align="center">
	<img src="./images/codeCheckout.png" alt="codeCheckout" width="500">
</div>

- 构建代码

<div align="center">
	<img src="./images/package.png" alt="package" width="500">
</div>

- 镜像构建

<div align="center">
	<img src="./images/imageBuild.png" alt="imageBuild" width="500">
</div>

- 集成测试

<div align="center">
	<img src="./images/integration.png" alt="integration" width="500">
</div>

- 发布仓库

<div align="center">
	<img src="./images/imageRelease.png" alt="imageRelease" width="500">
</div>


## 自动触发

支持在创建流水线时设置GitHub/GitLab的webhook，当用户向代码库中提交 commit、release 版本等操作时自动触发 CI/CD 工作流水线。

- 创建webhook

<div align="center">
	<img src="./images/webhook.png" alt="webhook" width="500">
</div>

## 实时日志

触发流水线执行后，支持查看各阶段实时日志。

<div align="center">
	<img src="./images/logStream.png" alt="logStream" width="500">
</div>
