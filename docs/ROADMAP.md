# Cyclone Roadmap

This document defines a high level roadmap of the Cyclone project. It should serve as a reference point for Cyclone users and contributors to understand where Cyclone is heading.

## Goals

### Third-party Storage

Make resource resolvers to support third-part storage, such as Amazon S3, Aliyun OSS to persistent resources, so that when running workflows, data can be pulled/pushed to these storage. 

###  Multi-cluster Support

Run workflow in different clusters.

### Multi-tenant Support

Support multi-tenancy.

### User System

Build Cyclone's own user system with auth management and also support LDAP.

### Cyclone CLI

Make a CLI tool to interact with Cyclone from command line.

### AI Pipeline

Define curated stage templates for AI pipelines.

### Auditing

Enable auditing for Cyclone operations on top of the user system.

## Release Planning

Dates in the following milestones should not be considered authoritative but rather indicative. Please refer to [milestones](https://github.com/caicloud/cyclone/milestones) for the most up-to-date and issue-by-issue plans.

### v1.0.0 (2019/03)

Main features:
- Support project management
- Integration center to integrate external services
- Build-in support for CI/CD pipelines by templates
- User friendly web UI to operate
- Webhook trigger
- Multiple cluster support
