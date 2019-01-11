# Cyclone Roadmap

This document defines a high level roadmap of Cyclone project. It should serve as a reference point for Cyclone users and contributors to understand where Cyclone is heading.

## Goals

### Third-party Storage

Make resource resolvers to support third-part storage, such as Amazon S3, Aliyun OSS to persistent resources, when run workflow, data can be pull/push to these storage. 

###  Multi-cluster Support

Run workflow in different clusters.

### Multi-tenant Support

Support multi-tenant.

### User System

Build Cyclone's own user system with auth management and also support LDAP.

### Cyclone CLI

Make a CLI tool to interact with Cyclone from command line.

### AI Pipeline

Define stage templates for AI pipelines.

### Auditing

Above the user system, enable auditing for Cyclone operations.

## Release Planning

Dates in the following milestones should not be considered authoritative, but rather indicative of the project timeline. Please refer to [milestones](https://github.com/caicloud/cyclone/milestones) defined in GitHub for the most up-to-date and issue-for-issue plans.

### v1.0.0 (2019/03)

Main features:
- Support project management
- Integration center to integrate external services
- Build-in support for CI/CD pipelines by templates
- User friendly web UI to operate
- Webhook trigger
- Multiple cluster support