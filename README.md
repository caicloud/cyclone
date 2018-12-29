# Cyclone

<h1 align="center">
	<br>
	<img width="400" src="docs/images/logo.jpeg" alt="cyclone">
	<br>
	<br>
</h1>

Cyclone is a cloud native workflow engine not only for CI/CD pipelines, but also for general workflows. It defines workflow by assembling workflow stages as DAG. As Cyclone totally bases on Kubernetes, it can be easily deployed and used.

Cyclone defines each workflow stage as a pod, and assemble them to workflow by capturing dependencies between stages as DAG. 

## Features
- **Kubernetes Native**: Totally based on Kubernetes without any other dependencies..
- **Workflow Engine**: Supports CI/CD pipelines, AI pipelines and any other general workflows.
- **DAG**: Cyclone supports any DAG workflow.
- **Pluggability**: Cyclone can easily integrate with external systems, such as SCM, Docker Registry, S3, etc.
- **Triggers**: Workflow can be triggered by time schedule, webhooks.
- **Controllable**: Execution of workflow can be paused, resumed or cancelled.
- **Multi-Tenant**: Cyclone supports multi-tenant.
- **Garbadge Collection**: Cleanup after workflow finished and results collected.

## Quick Start

Build images
```bash
$ make container
```

Deploy to Kubernetes cluster
```bash
$ kubectl create -f manifests/cyclone.yaml
```

## Community

- **Slack**: Join [Cyclone Community](https://caicloud-cyclone.slack.com/join/signup) for disscussions and ask questions.

## Roadmap

[Cyclone Roadmap](./docs/ROADMAP.md)

## History

Before v0.9.0, Cyclone focuses on CI/CD pipelines and run pipelines in both Docker and Kubernetes.

From v0.9.0, Cyclone has been totally refactored, new Cyclone not only aims at CI/CD pipeline, but also provides a general workflow engine. Compared with previous one, it's more flexible and powerful.
