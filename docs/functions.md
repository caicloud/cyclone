<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Function Introduction](#function-introduction)
  - [Convenient management](#convenient-management)
  - [Continuous integration](#continuous-integration)
  - [Trigger automatically](#trigger-automatically)
  - [Real-time logs](#real-time-logs)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->


# Function Introduction

## Convenient management
<!--## Relating with SCM systems -->

The convenient way to manage a group of pipelines is creating project related with a variety of SCM systems, such as GitHub, GitLab, SVN.

- Create project related with SCM 

<div align="center">
	<img src="./images/create_project.png" alt="create_project" width="500">
</div>

- List all projects

<div align="center">
	<img src="./images/list_projects.png" alt="list_projects" width="500">
</div>

## Continuous integration

All of the processes in workflow are visible. "codeCheckout" checkout code from specific repository; "package" compile the source code; "imageBuild" builds the published image; "integration" executes the integrated test; "imageRelease" publishes the image. All of the processes are shipped by container. It will wipe off the differences caused by environment.

- codeCheckout

<div align="center">
	<img src="./images/codeCheckout.png" alt="codeCheckout" width="500">
</div>

- package

<div align="center">
	<img src="./images/package.png" alt="package" width="500">
</div>

- imageBuild

<div align="center">
	<img src="./images/imageBuild.png" alt="imageBuild" width="500">
</div>

- integration

<div align="center">
	<img src="./images/integration.png" alt="integration" width="500">
</div>

- imageRelease

<div align="center">
	<img src="./images/imageRelease.png" alt="imageRelease" width="500">
</div>


## Trigger automatically


Cyclone has been integrated with a variety of VCS tools, such as git, svn, etc. After OAuth, it can pull codes from repository and create webhook. Whenever the user commits, submits a pull request or releases a version to the repository, the webhook will trigger the CI/CD workflow. 

Cyclone supports configuring GitHub/GitLab webhook while creating pipeline. Whenever the user commits, submits a pull request or releases a version to the repository, the webhook will trigger the CI/CD workflow. 

- Configure webhook

<div align="center">
	<img src="./images/webhook.png" alt="webhook" width="500">
</div>

## Real-time logs

Cyclone provides real-time logs of every pipeline stage while the pipeline is running. 

- Running log

<div align="center">
	<img src="./images/logStream.png" alt="logStream" width="500">
</div>
