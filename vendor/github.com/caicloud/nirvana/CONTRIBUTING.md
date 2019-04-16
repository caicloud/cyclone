# Nirvana Contributor Guide

Welcome to Nirvana! This document shows you how to contribute to the code base. Plaese file issues or pull requests if
you find anything missing or incorrect.

### Fork && Clone

First, you can fork this repository via the `Fork` button on the top of this page (If you like Nirvana, please star it :)).
Then you can clone Nirvana to your local path:

```
$ mkdir -p $GOPATH/src/github.com/caicloud
$ go get github.com/caicloud/nirvana

# Avoid push to caicloud/nirvana
$ git config push.default nothing

# Set user to match your github profile name:
$ user={your github profile name}

$ git remove add $user git@github.com:$user/nirvana.git
$ git fetch $user
```

There are some environment requirements:

- [Latest golang](https://golang.org/dl/)
- [Latest dep](https://github.com/golang/dep/)
- [Latest gometalinter](https://github.com/alecthomas/gometalinter)

### Find something you are interested in

Nirvana is under active development, there are quite a few incomplete features. Following is a list of items to checkout:

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

We welcome any thoughts or contributions on these items!

### File an issue to discuss your idea or design

Before start writing code, we strongly recommend you to file an issue to discuss the motivation and proposals.
Here are some workflows for you:

#### Write a new feature

1. Search in [Nirvana Issues](https://github.com/caicloud/nirvana/issues) to check if there is similar issue.
2. If not, file an issue to describe the feature you want to work on.
3. Discuss with other contributors and adjust your design accordingly.
4. Finalize your design.
5. Implement the feature and push to your own repository.
6. Create a pull request to Nirvana.
7. Code review and change your PR if necessary.
8. Other contributors will merge your pull request.

#### Take tasks from an issue

1. Find an issue you are intersted in.
2. Take the tasks by discussing with issue author and code owners.
3. Implement the feature (or fix the bug), and push to your own repository.
4. Create a pull request to Nirvana.
7. Code review and change your PR if necessary.
8. Other contributors will merge your pull request.

#### Fix typos, commments or submit docs

1. Modify and push to your own repository.
2. Create a pull request to Nirvana.
7. Code review and change your PR if necessary.
8. Other contributors will merge your pull request.

### Code review

Nirvana use an [automatic tool](https://github.com/caicloud-bot) to manage the project. Its basic functionality
is assigning pull requests to reviewers and merging pull requests if it received `/lgtm` and `/approve`.

Many directories of Nirvana have `OWNERS`, these files contains github account names of reviewers and approvers.
- Reviewers can review pull requests and reply `/lgtm` if they are satisfied with those pull requests.
- Approvers can give a final approval by `/approve` to indicate whether a change to a directory or subdirectory should be accepted.

If a pull request receives both `/lgtm` and `/approve`, it will be merged in a short duration.
