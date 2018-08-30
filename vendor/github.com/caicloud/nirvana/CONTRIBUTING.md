# Nirvana Contributor Guide

Welcome to Nirvana! This document shows you how to contribute to the code base. Plaese file issues or pull requests if you find something is missing or incorrect.

### Fork && Clone

First, you can fork this repository via the `Fork` button on the top of this page (If you like Nirvana, please star it :).  
After that, you can immediately clone `https://github.com/<your-account-name>/nirvana` to your local path `$GOPATH/src/github.com/caicloud/nirvana`.  
There are some environment requirements:  

- [Latest golang](https://golang.org/dl/)
- [Latest dep](https://github.com/golang/dep/)
- [Latest gometalinter](https://github.com/alecthomas/gometalinter)

### Find something you are interested in

Nirvana is in rapid development, many things are not complete. There are many components that you can dig into:

- Docs
  - English
  - 中文
  - ...
- Loggers
  - File Logger
  - Remote Logger
  - ...
- Operators
  - More Powerful Vaditor
  - ...
- Plugins
- OpenAPI Generation
- Clients Generation
  - go
  - graphql
  - ...
- Test Cases
- Core
  - API Definition
  - HTTP Service
  - Router
  - Code Refactor
  - ...
- Bugs
  - File Issues
  - Fix Issues
  - Fix Typo && Comments
  - ...

We are waiting for your thoughts and actions!


### File an issue to discuss your idea or design

Before you starting to write codes for Nirvana, we strongly recommend you file an issue to show us what you think.  
Here are some workflows for you:

#### Write a new feature

1. Search in [Nirvana Issues](https://github.com/caicloud/nirvana) to check there is no same issue.
2. File a issue to describe the feature which you want to drive.
3. Discuss with other contributors and modify your design if it's necessary.
4. Finalize your design.
5. Implement it and push to your own repository.
6. Create a pull request to Nirvana.
7. Code review and modify your codes if it's necessary.
8. Merge your pull request.

#### Take works from an issue
1. Find an issue which you are intersted in.
2. Take the works by discussing with issue author and code owners.
3. Implement it and push to your own repository.
4. Create a pull request to Nirvana.
5. Code review and modify your codes if it's necessary.
6. Merge your pull request.

#### Fix typos, commments or submit docs
1. Modify and push to your own repository.
2. Create a pull request to Nirvana.
3. Code review and modify your codes if it's necessary.
4. Merge your pull request.

### Code review

Nirvana use an automatic tool to manage the project. Its basic functionalities is assigning pull requests to reviewers and merging pull requests if it received `/lgtm` and `/approve`.  
Many directories of Nirvana have `OWNERS`, these files contains github account names of reviewers and approvers.

- Reviewers can review pull requests and reply `/lgtm` if they are satisfied with those pull requests.
- Approvers can provide a final approval by `/approve` to indicate whether a change to a directory or subdirectory should be accepted.

If a pull request collected `/lgtm` and `/approve`, it will be merged in a short duration.
