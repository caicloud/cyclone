# User Guide

## Workflow Engine

Cyclone provides a powerful workflow engine that serves general workflow purpose. You can orchestrate any workflows you need by taking advantage of it.

[Examples](../examples) shows some examples that can help you build your own workflows. To run the examples, run

```bash
$ make run_examples SCENE=cicd REGISTRIES=<registry>/<project>
```

In addition to YAML format manifest files, you can also create your workflows via APIs provided by Cyclone Server. Refer to [API Docs](./swagger-api-docs.md) for detailed APIs.

## Functional Modules

There are 3 main functional modules in Cyclone:

* [Integration Center](#Integration-Center): integrates external systems like SCM, docker registry, S3, etc.
* Project: manages a group of workflows and their shared configs.
* Stage Templates: Manages built-in stage templates and customized templates.

## Integration Center

External systems are called integrations in Cyclone, such as `GitHub`, `docker registry`, `k8s cluster`. They can be integrated to Cyclone conveniently. There is an integration center in Cyclone Web, where users can manage integrations.
From the view of implementation, one integration is saved as a `Secret` in k8s, and information like URL, credentials are hold by it.

Cyclone provides 5 builtin type of integrations: `SCM`, `Cluster`, `DockerRegistry`, `SonarQube` and `General`. Here `SCM` means source code management system like `GitHub`, `GitLab`. Cyclone support different types of SCM: `GitHub`, `GitLab`, `BitBucket` and `SVN`.

Supported external systems including:

|  Name  | Sub Class | Supported Versions | Note  |
| :---:  | :---:     | :---:              | :---: |
| SCM    | GitHub    | [Public](https://github.com/) | enterprise edition is unsupported |
|        | GitLab    | [Public](https://gitlab.com/), Private >= 8.13.6         | May [affect 10.6 and later versions](#GitLab-Webhooks) to create a SCM webhook automatically triggered pipeline |
|        | BitBucket | \>= 5.0 | BitBucket cloud is unsupported;<br> Different edition may affect [BitBucket Token and Webhooks](#BitBucket-Token-and-Webhooks) |
|        | SVN       | All                | |
| Cluster |          | All                | |
| DockerRegistry |   | All                | |
| SonarQube |        | All                | |
| General |          | All                | |

### GitLab Webhooks

> Starting with GitLab **10.6**, all Webhook requests to the current GitLab instance server address and/or in a private network will be forbidden by default.
That means that all requests made to 127.0.0.1, ::1 and 0.0.0.0, as well as IPv4 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16 and IPv6 site-local (ffc0::/10) addresses won’t be allowed.

This behavior can be overridden by enabling the option “Allow requests to the local network from hooks and services”, accessing address:
- Version 10.6 - 11.3.x , `{your-gitlab-server}/admin/application_settings`
- Version 11.4.x and later , `{your-gitlab-server}/admin/application_settings/network`

More detailed information: [Webhooks and insecure internal web services](https://docs.gitlab.com/ee/security/webhooks.html)

### BitBucket Token and Webhooks

#### Tokens

Personal access tokens are supported starting with **5.5**, so in the previous version, you can just use username/password for auth,
but in the later version, personal access token is supported.

#### Webhooks

BitBucket server supports basic webhook starting with **5.4**, and in **5.10**, there are 2 new events available for webhook:
`Pull request modified` and `PR reviewers updated`.

More information please reference to:
- [BitBucket Server 5.4 release notes](https://confluence.atlassian.com/bitbucketserver/bitbucket-server-5-4-release-notes-935388966.html)
- [BitBucket Server 5.5 release notes](https://confluence.atlassian.com/bitbucketserver/bitbucket-server-5-5-release-notes-938037662.html)
- [BitBucket Server 5.10 release notes](https://confluence.atlassian.com/bitbucketserver/bitbucket-server-5-10-release-notes-948214779.html)

## SVN Post-Commit hook

Cyclone server supports SVN post-commit hook to trigger workflow. Using this feature, you should do two things:
- configure your SVN repository
- create SCM type workflowtriggers

### Configure your SVN repository

Login your SVN server, and do the following steps to configure hooks:

- Make sure `curl` is installed at `/usr/bin/curl` in you SVN server.

- Navigate to your repository’s hooks directory. This is almost always a directory cleverly named `hooks` right inside the top level of your repository:
```
cd /{home_svn}/{my_repository}/hooks/
```

- Create a new file called post-commit with following content, and make it executable.
```shell
$ cat <<'EOF' > post-commit
#!/bin/sh

REPOS="$1"
REV="$2"
TXN_NAME="$3"

UUID=`svnlook uuid $REPOS`

/usr/bin/curl --request POST \
              --header "Content-Type:application/json;charset=UTF-8" \
              --header "X-Subversion-Event:Post-Commit" \
              --data "{\"repoUUID\":\"${UUID}\", \"revision\":\"${REV}\"}" \
              http://{cyclone-server-address}/apis/v1alpha1/tenants/{tenant}/webhook?sourceType=SCM

EOF

$ chmod 755 ./post-commit
```
Please replace `{cyclone-server-address}` and `{tenant}` with correct value.

### Create SCM type workflowtriggers

Note that the `workflowtrigger.spc.scm.repo` field should be SVN repository's uuid, you can get it by:
```
svn info --show-item repos-uuid --username {user} --password {password} --non-interactive --trust-server-cert-failures unknown-ca,cn-mismatch,expired,not-yet-valid,other --no-auth-cache {svn-repo-url}
```
