<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Cyclone API V1](#cyclone-api-v1)
  - [Endpoint](#endpoint)
  - [API Summary](#api-summary)
    - [Project APIs](#project-apis)
    - [Pipeline APIs](#pipeline-apis)
    - [Pipeline Record APIs](#pipeline-record-apis)
    - [Pipeline Record Logs API](#pipeline-record-logs-api)
    - [SCM API](#scm-api)
    - [Webhook API](#webhook-api)
    - [Stats API](#stats-api)
    - [Cloud API](#cloud-api)
    - [Template API](#template-api)
    - [Integration API](#integration-api)
  - [API Common](#api-common)
    - [Path Parameter Explanation](#path-parameter-explanation)
  - [API Details](#api-details)
    - [ProjectObject](#projectobject)
    - [List projects](#list-projects)
    - [Create project](#create-project)
    - [Get project](#get-project)
    - [Update project](#update-project)
    - [Delete project](#delete-project)
    - [PipelineObject](#pipelineobject)
    - [PipelineRecordObject](#pipelinerecordobject)
    - [ListedPipelineObject](#listedpipelineobject)
    - [List pipelines](#list-pipelines)
    - [Create pipeline](#create-pipeline)
    - [Get Pipeline](#get-pipeline)
    - [Update pipeline](#update-pipeline)
    - [Delete pipeline](#delete-pipeline)
    - [PipelinePerformParamsObject](#pipelineperformparamsobject)
    - [List pipeline records](#list-pipeline-records)
    - [Create pipeline record](#create-pipeline-record)
    - [Get pipeline record](#get-pipeline-record)
    - [Delete pipeline record](#delete-pipeline-record)
    - [Update Pipeline Record Status](#update-pipeline-record-status)
    - [Get Pipeline Record Log](#get-pipeline-record-log)
    - [Get Realtime Pipeline Record Log](#get-realtime-pipeline-record-log)
    - [RepoObject](#repoobject)
    - [List SCM Repos](#list-scm-repos)
    - [List SCM Branches](#list-scm-branches)
    - [List SCM Tags](#list-scm-tags)
    - [Get SCM Templatetype](#get-scm-templatetype)
    - [Github webhook](#github-webhook)
    - [Gitlab webhook](#gitlab-webhook)
    - [SVN hooks](#svn-hooks)
      - [Config your svn repository](#config-your-svn-repository)
    - [PipelineStatusStatsObject](#pipelinestatusstatsobject)
    - [Get Project Stats](#get-project-stats)
    - [Get Pipeline Stats](#get-pipeline-stats)
    - [CloudObject](#cloudobject)
    - [Create cloud](#create-cloud)
    - [List clouds](#list-clouds)
    - [Get cloud](#get-cloud)
    - [Update cloud](#update-cloud)
    - [Delete cloud](#delete-cloud)
    - [WorkerInstance](#workerinstance)
    - [List cyclone workers](#list-cyclone-workers)
    - [ConfigTemplateObject](#configtemplateobject)
    - [List config templates](#list-config-templates)
    - [IntegrationObject](#integrationobject)
    - [Create integration](#create-integration)
    - [List integrations](#list-integrations)
    - [Get integration](#get-integration)
    - [Update integration](#update-integration)
    - [Delete integration](#delete-integration)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Cyclone API V1

## Endpoint

```
    +--------+
    | caller |
    +-+------+
      |
      | APIs
      |
+-------------+ Server Side
      |
      v 
    +-+-------+
    | cyclone |
    +---------+
```

## API Summary

### Project APIs

- [Project Object](#ProjectObject)

| API | Path | Detail |
| --- | --- | --- |
| List | GET `/api/v1/projects` | [link](#list-projects) |
| Create | POST `/api/v1/projects` | [link](#create-project) |
| Get | GET `/api/v1/projects/{project}` | [link](#get-project) |
| Update | PATCH `/api/v1/projects/{project}` | [link](#update-project) |
| Delete | DELETE `/api/v1/projects/{project}` | [link](#delete-project) |

### Pipeline APIs

- [Pipeline Object](#PipelineObject)
- [Pipeline Record Object](#PipelineRecordObject)
- [Listed Pipeline Object](#listedpipelineobject)

| API | Path | Detail |
| --- | --- | --- |
| List | GET `/api/v1/projects/{project}/pipelines` | [link](#list-pipelines) |
| Create | POST `/api/v1/projects/{project}/pipelines` | [link](#create-pipeline) |
| Get | GET `/api/v1/projects/{project}/pipelines/{pipeline}` | [link](#get-pipeline) |
| Update | PUT `/api/v1/projects/{project}/pipelines/{pipeline}` | [link](#update-pipeline) |
| Delete | DELETE `/api/v1/projects/{project}/pipelines/{pipeline}` | [link](#delete-pipeline) |

### Pipeline Record APIs

- [Pipeline Record Object](#PipelineRecordObject)

| API | Path | Detail |
| --- | --- | --- |
| List | GET `/api/v1/projects/{project}/pipelines/{pipeline}/records` | [link](#list-pipeline-records) |
| Create | POST `/api/v1/projects/{project}/pipelines/{pipeline}/records` | [link](#create-pipeline-record) |
| Get | GET `/api/v1/projects/{project}/pipelines/{pipeline}/records/{recordid}` | [link](#get-pipeline-record) |
| Delete | DELETE `/api/v1/projects/{project}/pipelines/{pipeline}/records/{recordid}` | [link](#delete-pipeline-record) |
| Update Status | PATCH `/api/v1/projects/{project}/pipelines/{pipeline}/records/{recordid}/status` | [link](#update-pipeline-record-status) |

### Pipeline Record Logs API

| API | Path | Detail |
| --- | --- | --- |
| Get | GET `/api/v1/projects/{project}/pipelines/{pipeline}/records/{recordid}/logs[?stage=&task=&download=]` | [link](#get-pipeline-record-log) |
| Get | GET `/api/v1/projects/{project}/pipelines/{pipeline}/records/{recordid}/logstream[?stage=&task=]` | [link](#get-realtime-pipeline-record-log) |

### SCM API

- [Repo Object](#RepoObject)

| API | Path | Detail |
| --- | --- | --- |
| List | GET `/api/v1/projects/{project}/repos` | [link](#list-scm-repos) |
| List | GET `/api/v1/projects/{project}/branches?repo=` | [link](#list-scm-branches) |
| List | GET `/api/v1/projects/{project}/tags?repo=` | WIP, [link](#list-scm-tags) |
| Get  | GET `/api/v1/projects/{project}/templatetype?repo=` | WIP, [link](#get-scm-templatetype) |

### Webhook API

| API | Path | Detail |
| --- | --- | --- |
| Create | POST `/api/v1/pipelines/{pipelineid}/githubwebhook` | WIP, [link](#github-webhook) |
| Create | POST `/api/v1/pipelines/{pipelineid}/gitlabwebhook` | WIP, [link](#gitlab-webhook) |
| Create | POST `/api/v1/subversion/{svnrepoid}/postcommithook` | [link](#svn-hooks) |

### Stats API

- [PipelineStatusStatsObject](#pipelinestatusstatsobject)

| API | Path | Detail |
| --- | --- | --- |
| Get | GET `/api/v1/projects/{project}/stats?[startTime=&endTime=]` | WIP, [link](#get-project-stats) |
| Get | GET `/api/v1/projects/{project}/pipelines/{pipeline}/stats?[startTime=&endTime=]` | WIP, [link](#get-pipeline-stats) |

### Cloud API

- [CloudObject](#cloudobject)

| API | Path | Detail |
| --- | --- | --- |
| Create | POST `/api/v1/clouds` | WIP, [link](#create-cloud) |
| List | GET `/api/v1/clouds` | WIP, [link](#list-clouds) |
| Get | GET `/api/v1/clouds/{cloud}` | WIP, [link](#get-cloud) |
| Update | PUT `/api/v1/clouds/{cloud}` | WIP, [link](#update-cloud) |
| Delete | DELETE `/api/v1/clouds/{cloud}` | WIP, [link](#delete-cloud) |

### Template API

- [PipelineTemplateObject](#pipelinetemplateobject)

| API | Path | Detail |
| --- | --- | --- |
| List | GET `/api/v1/configtemplates` | WIP, [link](#list-config-templates) |

### Integration API
- [IntegrationObject](#integrationobject)

| API | Path | Detail |
| --- | --- | --- |
| Create | GET `/api/v1/integrations` | WIP, [link](#create-integration) |
| List | GET `/api/v1/integrations` | WIP, [link](#list-integrations) |
| Get | GET `/api/v1/integrations/{integration}` | WIP, [link](#get-integration) |
| Update | GET `/api/v1/integrations/{integration}` | WIP, [link](#update-integration) |
| Delete | GET `/api/v1/integrations/{integration}` | WIP, [link](#delete-integration) |


## API Common

### Path Parameter Explanation

In path parameter, both `{project}` and `{pipiline}` are `name`, and only `{recordid}` is `ID`

## API Details

### ProjectObject

```
{
  "name": "string",                             // name of the project, if not provided, will be generated from alias.
  "alias": "string",                            // alias of the project
  "description": "string",                      // description of the project
  "owner": "string",                            // owner of the project
  "scm": {                                      // required
    "type": "",                                 // string, required. Only support "Gitlab","Github" and "SVN".
    "server": "",                               // string, required
    "username": "",                             // string, optional
    "password": "",                             // string, optional
    "token": ""                                 // string, optional
  },
  "registry": {                                 // optional
    "server": "",                               // string, required
    "repository": "",                           // string, required
    "username": "",                             // string, optional
    "password": ""                              // string, optional
  },
  "worker": {                                   // optional
    "location": {
        "cloudName": "k8s",                     // string, required
        "namespace": "cyclone-worker"           // string, optional
    },
    "dependencyCaches": {                       // map, optional
        "Maven": {
            "name": "devops-maven-cache"        // string, required
        },
    },
    "quota": {
        "requests.cpu": "0.5",                  // string, optional
        "requests.memory": "1G",                // string, optional
        "limits.cpu": "1",                      // string, optional
        "limits.memory": "2G"                   // string, optional
    }
  },
  "creationTime": "2017-08-23T09:44:08.653Z",   // created time of the project
  "lastUpdateTime": "2017-08-23T09:44:08.653Z", // updated time of the project
}
```

Note:

| Field | Note |
| --- | --- |
| cloudName | Required，name of cluster for workers to run. |
| namespace | Optional，only for k8s cloud, k8s namespace for workers to run. If not provided, workers will run in the same namespace with server. |
| quota | Optional，quota of worker, if not provided, will use the default. |

Project is responsible for managing a set of applications, in this system, mainly managing a number of pipelines corresponding to this set of applications. It manages the common configs shared by all pipelines in this project.

### List projects

List all projects.

**Request**

URL: `GET /api/v1/projects[?start=&limit=]`

Args:

| Name | Type | Detail |
| --- | --- | --- |
| start | number, optional | The starting position of the query. |
| limit | number, optional | The quantity limit for the result. |

**Response**

Success:

```
200 OK

{
    "metadata": {
        "total": 0,            // number, always 
    },
    "items": [ <ProjectObject>, ... ]
}
```

### Create project

Create a new project.

**Request**

URL: `POST /api/v1/projects`

Body:

```
{
  "name": "string",                             // name of the project, if not provided, will be generated from alias.
  "alias": "string",                            // alias of the project
  "description": "string",                      // description of the project
  "owner": "string"                             // owner of the project
  "scm": {                                      // required
    "type": "",                                 // string, required. Supports "Gitlab", "Github" and "SVN".
    "server": "",                               // string, required
    "username": "",                             // string, optional
    "password": "",                             // string, optional
    "token": "",                                // string, optional
  },
  "registry": {                                 // optional
    "server": "",                               // string, required
    "repository": "",                           // string, required
    "username": "",                             // string, optional
    "password": "",                             // string, optional
  },
  "worker": {                                   // optional
    "namespace": "devops",                      // string, optional
    "dependencyCaches": {                       // map, optional
        "Maven": {
            "name": "devops-maven-cache"        // string, required
        },
    },
  }
}
```

Note:

| Field | Note |
| --- | --- |
| name | ^[a-z0-9]+((?:[._-][a-z0-9]+)*){1,29}$ |
| description | The length is limited to 100 characters |
| type | Supports `Gitlab`, `Github` and `SVN` |

SCM supports two types of auth: 1. username and password; 2. token. At least one type of auth should be provided. If both types are provided, username and password will be used.
When username and password are provided, the password will not be stored, but a token will be generated and stored instead, the token will be used to access SCM server.

Difference between the auth of Github and Gitlab:

* Gitlab: When use token, username is not required. When generate new token through username and password, the old token is still valid.
* Github: When use token, username is still required, token just equals password. When generate new token through username and password, the old token will be invalid. If enable 2-factor authorization, please use personal access token.

**Response**

Success:

```
201 Created

{
    // all <ProjectObject> fields
}
```

### Get project

Get the information of a project.

**Request**

URL: `GET /api/v1/projects/{project}`

**Response**

Success:

```
200 OK

{
    // all <ProjectObject> fields
}
```

### Update project

Update the information of a project.

**Request**

URL: `PUT /api/v1/projects/{project}`

Body:

```
{
    // all <ProjectObject> fields
}
```

**Response**

Success:

```
200 OK

{
    // all <ProjectObject> fields
}
```

### Delete project

Delete project.

**Request**

URL: `DELETE /api/v1/projects/{project}`

**Response**

Success:

```
204 NoContent
```

### PipelineObject

```
{
    "id": "string", 
    "name": "string", 
    "alias": "string", 
    "description": "string", 
    "owner": "string", 
    "projectID": "string", 
    "notification": {                 // not implemented
      "policy": "Always",             // string  Always,Success,Failure
      "receivers": [
        {
          "type": "email",            // string  email,webhook,slack
          "addresses": [
            "abc@caicloud.io",
            "def@caicloud.io"
          ],
          "groups": [
            "group1",
            "group2"
          ]
        },
        {
          "type": "webhook",
          "addresses": [
            "http://remote/call"
          ],
          "groups": [
            "group1",
            "group2"
          ]
        }
      ]
    }
    "autoTrigger": {
        "cronTrigger": {                    // not implemented
            "crons": [
                {
                    "expression": "string", 
                    "stages": ["string", ...]
                }
            ]
        }, 
        "scmTrigger": {                     // return error if can not create webhook as not admin
            "push": {
                "stages": ["string", ...],
                "branches": ["string", ...]
            }, 
            "tagRelease": {
                "stages": ["string", ...]
            },
            "pullRequest": {
                "stages": ["string", ...]
            },
            "pullRequestComment": {
                "stages": ["string", ...],
                "comments": ["string", ...]
            },
            "postCommit":{
                "stages": ["string", ...]
            }
        }
    }, 
    "build": {
        "builderImage": {
            "envVars": [
                {
                    "name": "string", 
                    "value": "string"
                }
            ], 
            "image": "string"
        }, 
        "stages": {
            "codeCheckout": {
                "mainRepo": {                             // Required
                    "github": {
                        "ref": "refs/heads/master",       // string, required
                        "url": "string"
                    },
                    "gitlab": {
                        "ref": "refs/heads/master",       // string, required
                        "url": "string"
                    },
                    "otherCodeSource": {
                        "password": "string",
                        "path": "string",
                        "url": "string",
                        "username": "string"
                    },
                    "type": "string"                 // type of code source, it must be Github or Gitlab or other.
                },
                "depRepos": [                        // Not support now
                    {
                        "github": {
                            "ref": "refs/heads/master",       // string, required
                            "url": "string"
                        },
                        "gitlab": {
                            "ref": "refs/heads/master",       // string, required
                            "url": "string"
                        },
                        "otherCodeSource": {
                            "password": "string",
                            "path": "string",
                            "url": "string",
                            "username": "string"
                        },
                        "folder": "string",           // sub folder which dependent repos will be cloned into.
                        "type": "string"              // type of code source, it must be Github or Gitlab or other.
                    }
                ]
            },
            "unitTest": {
                "command": ["string", ...], 
                "outputs": ["string", ...]
            }, 
            "codeScan": {
                "sonarqube": {
                    "name":"sonar1", // sonarqube integration name
                    "config": {
                        "sourcePath": "./", // default './'
                        "encodingStyle": "UTF-8", // default 'UTF-8'
                        "language": "Java",
                        "threshold": ""
                    }
                }
            }, 
            "package": {
                "command": ["string", ...], 
                "outputs": ["string", ...]
            }, 
            "imageBuild": {
                "buildInfos": [                       // only support one element
                    {
                        "taskName": "string",
                        "contextDir": "string", 
                        "dockerfile": "string", 
                        "dockerfilePath": "string", 
                        "imageName": "string"
                    }
                ]
            }, 
            "imageRelease": {
                "releasePolicies": [                    // image names which are not listed will be ignored.
                    {
                        "imageName": "string", 
                        "type": "string"                // type of image release policy, it must be Always or IntegrationTestSuccess.
                    }
                ]
            }, 
            "integrationTest": {
                "config": {
                    "command": ["string", ...],
                    "envVars": [
                        {
                            "name": "string", 
                            "value": "string"
                        }
                    ], 
                    "imageName": "string"
                }, 
                "services": [
                    {
                        "command": ["string", ...], 
                        "envVars": [
                            {
                                "name": "string", 
                                "value": "string"
                            }
                        ], 
                        "image": "string", 
                        "name": "string"
                    }
                ]
            }
        }
    },
    "annotations": {       // map[string]string, optional
        "key1": "value1",
    },
    "creationTime": "2017-08-23T08:40:33.764Z", 
    "lastUpdateTime": "2017-08-23T08:40:33.764Z"
}
```

Pipeline is responsible for automating the lifecycle management of an application, and can safely and reliably deploy the application from the source code to the production environment in strict accordance with a set of scientific and rational software management processes.

### PipelineRecordObject

```
{
    "id": "",                   // string, required
    "name": "",                 // string, required
    "pipelineID": "",           // string, required
    "trigger": "",              // string, optional. Can be user, scmTrigger, cronTrigger
    "performParams": {
        "ref": "refs/tags/v0.1.0",        // string, required
        "name": "newVersion",             // string, optional
        "description": "",                // string, optional
        "createScmTag": false,            // bool, optional
        "cacheDependency": false       // bool, optional
        "stages": ["codeCheckout", "unitTest", ...]     // string array, optional. Elements can be codeCheckout、unitTest、codeScan、package、imageBuild、integrationTest and imageRelease
    },
    "stageStatus": {                                    // struct, required
        "codeCheckout": {                               // struct, required
            "status": "Running|Success|Failed|Aborted", // string, required
            "commits": {                                // struct, optional
                "mainRepo": {                           // struct, optional
                    "ID": "",                           // struct, required
                    "Author": "",                       // struct, required
                    "Date": "",                         // struct, required
                    "Message": ""                       // struct, required
                },
                "depRepos": [                           // struct, optional
                    {
                        "ID": "",                       // struct, required
                        "Author": "",                   // struct, required
                        "Date": "",                     // struct, required
                        "Message": ""                   // struct, required
                    },
                ]
            },
            "startTime": "2017-08-23T08:40:33.764Z",    // time, required
            "endTime": "2017-08-23T08:40:33.764Z"       // time, optional
        },
        "package": {                                    // struct, required
            "status": "Running|Success|Failed|Aborted", // string, required
            "startTime": "2017-08-23T08:40:33.764Z",    // time, required
            "endTime": "2017-08-23T08:40:33.764Z"       // time, optional
        },
        "codeScan": {
            "status": "Running|Success|Failed", // string, required
            "startTime": "2017-08-23T08:40:33.764Z",    // time, required
            "endTime": "2017-08-23T08:40:33.764Z"       // time, optional
            "sonarqube": {
                "measures": [
                    {
                        "metric": "reliability_rating",
                        "value": "3.0"
                    },
                    {
                        "metric": "coverage",
                        "value": "3.2"
                    }
                ],
            "overviewLink": "http://sonarqube:9000/component_measures?id=5bfe0f872f6c050001da6fb0"
            }
        },
        "imageBuild": {                                 // struct, optional
            "status": "Running|Success|Failed|Aborted", // string, required
            "startTime": "2017-08-23T08:40:33.764Z",    // time, required
            "endTime": "2017-08-23T08:40:33.764Z"       // time, optional
            "tasks": [
                {
                    "name": "v1",                                   // string, required
                    "status": "Running|Success|Failed|Aborted",     // string, required
                    "image": "test:v1",                             // string, required
                    "startTime": "2017-08-23T08:40:33.764Z",        // time, required
                    "endTime": "2017-08-23T08:40:33.764Z"           // time, optional
                }
            ]
        },
        "integrationTest": {                            // struct, optional
            "status": "Running|Success|Failed|Aborted", // string, required
            "startTime": "2017-08-23T08:40:33.764Z",    // time, required
            "endTime": "2017-08-23T08:40:33.764Z"       // time, optional
        },
        "imageRelease": {                               // struct, optional
            "status": "Running|Success|Failed|Aborted", // string, required
            "images": ["xx.io/xx/server:v1", ...],      // struct, required
            "startTime": "2017-08-23T08:40:33.764Z",    // time, required
            "endTime": "2017-08-23T08:40:33.764Z"       // time, optional
            "tasks": [
                {
                    "name": "test:v1",                              // string, required
                    "status": "Running|Success|Failed|Aborted",     // string, required
                    "image": "test:v1",                             // string, required
                    "startTime": "2017-08-23T08:40:33.764Z",        // time, required
                    "endTime": "2017-08-23T08:40:33.764Z"           // time, optional
                }
            ]
        },
    },
    "status": "",               // string, required. Can be Pending, Running, Success, Failed, Aborted.
    "errorMessage": ""          // string, optional
    "startTime": "2017-08-23T08:40:33.764Z",        // time, required
    "endTime": "2017-08-23T08:40:33.764Z",          // time, optional
}
```

### ListedPipelineObject

```
{
    <PipelineObject>
    recentRecords: [<PipelineRecordObject>, ... ],
    recentSuccessRecords: [<PipelineRecordObject>, ... ],
    recentFailedRecords: [<PipelineRecordObject>, ... ]
}
```

### List pipelines

List all pipelines of one project.

**Request**

URL: `GET /api/v1/projects/{project}/pipelines[?start=&limit=&recentCount=&recentSuccessCount=&recentFailedCount=]`

Args:

| Field | Type | Detail |
| --- | --- | --- |
| start | number, optional | The starting position of the query. |
| limit | number, optional | The quantity limit for the result. |
| recentCount | number, optional | The default is 0, and the latest pipeline record is not returned |
| recentSuccessCount | number, optional | The default is 0, and the latest successful pipeline record is not returned |
| recentFailedCount | number, optional | The default is 0, and the latest failed pipeline record is not returned |

**Response**

Success:

```
200 OK

{
    "metadata": {
        "total": 0,  // number, always 
    },
    "items": [ <PipelineObject>, ... ]
}
```

### Create pipeline

Create a new pipeline.

**Request**

URL: `POST /api/v1/projects/{project}/pipelines`

Body:

```
{
    "name": "newpipeline",             // string, required
    "alias": "",                        // string, optional
    "description": ""                   // string, optional
    "build": {
        "builderImage": {
            "envVars": [
                {
                    "name": "string", 
                    "value": "string"
                }
            ], 
            "image": "string"          // string, required
        }, 
        "stages": {
            "codeCheckout": {
                "codeSources": [
                    {
                        "gitHub": {
                            "ref": "string", 
                            "url": "string"
                        }, 
                        "gitLab": {
                            "ref": "string", 
                            "url": "string"
                        }, 
                        "main": true, 
                        "otherCodeSource": {
                            "password": "string", 
                            "path": "string", 
                            "url": "string", 
                            "username": "string"
                        }, 
                        "type": "string"                // type of code source, it must be github or gitlab or other.
                    }
                ]
            }, 
            "unitTest": {             // struct, optional
                "command": ["string", ...], 
                "outputs": ["string", ...]
            }, 
            "codeScan": {             // struct, optional
                "command": ["string", ...], 
                "outputs": ["string", ...]
            }, 
            "package": {
                "command": ["string", ...],           // array, required
                "outputs": ["string", ...]            // array, required
            }, 
            "imageBuild": {
                "buildInfos": [       // array, required
                    {
                        "contextDir": "string", 
                        "dockerfile": "string", 
                        "dockerfilePath": "string", 
                        "imageName": "string"          // string, required
                    }
                ]
            }, 
            "imageRelease": {         // struct, optional
                "releasePolicy": [
                    {
                        "imageName": "string",
                        "type": "string"                // type of image release policy, it must be Always or IntegrationTestSuccess.
                    }
                ]
            }, 
            "integrationTest": {      // struct, optional
                "config": {
                    "command": ["string", ...], 
                    "imageName": "string"
                }, 
                "services": [
                    {
                        "command": ["string", ...], 
                        "envVars": [
                            {
                                "name": "string", 
                                "value": "string"
                            }
                        ], 
                        "image": "string", 
                        "name": "string"
                    }
                ]
            }
        }
    }
}
```

**Response**

Success:

```
201 Created

{
    // all <PipelineObject> fields
}
```

### Get Pipeline

Get a pipeline.

**Request**

URL: `GET /api/v1/projects/{project}/pipelines/{pipeline}`

**Response**

Success:

```
200 OK

{
    // all <PipelineObject> fields
}
```

### Update pipeline

Update the information of a pipeline.

**Request**

URL: `PUT /api/v1/projects/{project}/pipelines/{pipeline}`

Body:

```
{
    // all <PipelineObject> fields
}
```

Note:

| Field | Note |
| --- | --- |
| name | ^[a-z0-9]+((?:[._-][a-z0-9]+)*){1,29}$ |
| description | The length is limited to 100 characters |

**Response**

Success:

```
200 OK

{
    // all <PipelineObject> fields
}
```

### Delete pipeline

Delete a pipeline.

**Request**

URL: `DELETE /api/v1/projects/{project}/pipelines/{pipeline}`

**Response**

Success:

```
204 NoContent
```

### PipelinePerformParamsObject

```
{
    "ref": "master",                  // string, required. Support SCM branch and tag.
    "name": "newVersion",             // string, optional. Genereated if not specified, used as repo and image tag.
    "description": "",                // string, optional
    "createScmTag": false,            // bool, optional
    "stages": ["codeCheckout", "unitTest", ... ]        // string array, optional. If not specified, will perfrom all stages in pipeline.
}
```

Params to perform pipeline.

### List pipeline records

List all pipeline records of one pipeline.

**Request**

URL: `GET /api/v1/projects/{project}/pipelines/{pipeline}/records[?start=&limit=]`

Args:

| Name | Type | Detail |
| --- | --- | --- |
| start | number, optional | The starting position of the query. |
| limit | number, optional | The quantity limit for the result. |

**Response**

Success:

```
200 OK

{
    "metadata": {
        "total": 0,  // number, always 
    },
    "items": [ <PipelineRecordObject>, ... ]
}
```

### Create pipeline record

Create a pipeline record, which means trigger a pipeline.

**Request**

URL: `POST /api/v1/projects/{project}/pipelines/{pipeline}/records`

Body:

```
{
    "ref": "master",                                       // string, required. Support SCM branch and tag.
    "name": "v1.0.0",                                      // string, optional
    "description": "",                                     // string, optional
    "createScmTag": false,                                 // bool, optional
    "stages": ["codeCheckout", "package", ... ]            // string array, optional
}
```

Note:

| Field | Note |
| --- | --- |
| version | ^[a-z0-9]+((?:[._-][a-z0-9]+)*){1,29}$ |
| description | The length is limited to 100 characters |
| stages | Only can be some of `codeCheckout、unitTest、codeScan、package、imageBuild、integrationTest、imageRelease` |

**Response**

Success:

```
201 Created

{
    // all <PipelineRecordObject> fields
}
```

### Get pipeline record

Get the pipeline execution information.

**Request**

URL: `GET /api/v1/projects/{project}/pipelines/{pipeline}/records/{recordid}`

**Response**

Success:

```
200 OK

{
    // all <PipelineRecordObject> fields
}
```

### Delete pipeline record

Delete a pipeline record.

**Request**

URL: `DELETE /api/v1/projects/{project}/pipelines/{pipeline}/records/{recordid}`

**Response**

Success:

```
204 NoContent
```

### Update Pipeline Record Status

Update the pipeline record status. Only the pipeline status of the Running state can be changed to Aborted, which means the pipeline is terminated.

**Request**

URL: `PATCH /api/v1/projects/{project}/pipelines/{pipeline}/records/{recordid}/status`

Body:

```
{
    "status": ""  // string, required
}
```

Note:

| Field | Note |
| --- | --- |
| status | Only support Aborted status |

**Response**

Success:

```
200 Ok

{
    // all <PipelineRecordObject> fields
}
```

### Get Pipeline Record Log

Get the logs of finished pipeline records.

**Request**

URL: `GET /api/v1/projects/{project}/pipelines/{pipeline}/records/{recordid}/logs[?stage=&&download=]`

**Response**

Success:

```
200 OK

Content-Type: text/plain

Stage: Clone repository status: start
Cloning into 'code'...
Stage: Clone repository status: finish
Stage: Package status: start
$ echo hello
hello
Stage: Package status: finish
Stage: Build image status: start
Step 1 : FROM cargo.caicloud.io/caicloud/cyclone-worker:latest
 ---> 2437d0db0a28
Step 2 : ADD ./README.md /README.md
  ---> e8ca485e1ab8
......
Stage: Build image status: finish
```

| Field | Note |
| --- | --- |
| stage | Can be only one of `codeCheckout`、`unitTest`、`codeScan`、`package`、`imageBuild`、`integrationTest` and `imageRelease`, the stage must be performed in this record. Currently, `unitTest` and `codeScan` are not supported: `unitTest` is merged into `package`; `codeScan` is not implemented. If provided, only return the log of this stage; if not provided, will return all log. Not provided in default. |
| task | Sub task name. If stage is single task, the param will be omitted. If stage is multiple task, will return log of specified task. If the log is not found, will return 404. |
| download | true: download logs, the log file name is {projectName}-{pipelineName}-{recordid}-log.txt; false: directly return logs. `False` in default. |

### Get Realtime Pipeline Record Log

Get the real-time logs for running pipeline records.

**Request**

URL: `GET /api/v1/projects/{project}/pipelines/{pipeline}/records/{recordid}/logstream[?stage=&task=]`

Header:
```
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==
Sec-WebSocket-Version: 13
```

**Response**

Success:

```
101 UPGRADED
Content-Type: text/plain
Connection: Upgrade
Upgrade: tcp

Stage: Clone repository status: start
Cloning into 'code'...
Stage: Clone repository status: finish
Stage: Package status: start
$ echo hello
hello
Stage: Package status: finish
Stage: Build image status: start
Step 1 : FROM cargo.caicloud.io/caicloud/cyclone-worker:latest
 ---> 2437d0db0a28
Step 2 : ADD ./README.md /README.md
  ---> e8ca485e1ab8
......
Stage: Build image status: finish
```

Note:

The illustration of stage param is the same as that of last static log API.

### RepoObject

```
{
	"name": string,                                        // string, required. The format is {username}/{repoName}.
	"url": string                                          // string, required.
}
```

### List SCM Repos

All repositories can be accessed by this project.

**Request**

URL: `GET /api/v1/projects/{project}/repos`

**Response**

Success:

```
200 OK

{
    "metadata": {
        "total": 0,  // number, always
    },
    "items": [ <RepoObject>, ... ]
}
```

### List SCM Branches

List all branches for the repositories can be accessed by this project.

**Request**

URL: `GET /api/v1/projects/{project}/branches?repo=`

Note:

| Field | Note |
| --- | --- |
| repo | Required。Which repo to list branch for |

**Response**

Success:

```
200 OK

{
    "metadata": {
        "total": 0,  // number, always
    },
    "items": [ <branch>, ... ]
}
```

### List SCM Tags

List all tags for the repositories can be accessed by this project.

**Request**

URL: `GET /api/v1/projects/{project}/tags?repo=`

Note:

| Field | Note |
| --- | --- |
| repo | Required, repo for which the tags will be listed |

**Response**

Success:

```
200 OK

{
    "metadata": <PaginationObject>,
    "items": [ <tag>, ... ]
}
```

### Get SCM Templatetype

Get the template type of the specific repo.

**Request**

URL: `GET /api/v1/projects/{project}/templatetype?repo=`

Note:

| Field | Note |
| --- | --- |
| repo | Required, repo for which the template type will be get |


**Response**

Success:

```
200 OK

{
    "type": <type>
}
```

### Github webhook

Trigger pipeline by Github webhook.

**Request**

URL: `POST /api/v1/pipelines/{pipelineid}/githubwebhook`

Header:
```
X-GitHub-Event: ReleaseEvent|ReleaseEvent|PullRequestEvent
```

Body:

Refer to: https://developer.github.com/webhooks/#payloads

**Response**

Success:

```
200 OK
```

### Gitlab webhook

Trigger pipeline by Gitlab webhook.

**Request**

URL: `POST /api/v1/pipelines/{pipelineid}/gitlabwebhook`

Header:
```
X-GitHub-Event: Push Hook|Tag Push Hook|Merge Request Hook
```

Body:

Refer to: https://docs.gitlab.com/ce/user/project/integrations/webhooks.html

**Response**

Success:

```
200 OK
```

### SVN hooks

Trigger pipeline by SVN post-commit hooks.

**Request**

URL: `POST /api/v1/subversion/{svnrepoid}/postcommithook`

Header:
```
Content-Type:text/plain;charset=UTF-8
```

Query:
```
revision: 27 // {revision-id}
```

Body:
Output of `svnlook changed --revision $REV $REPOS`, for example:
```
U   cyclone/test.go
U   cyclone/README.md
```

**Response**

Success:

```
200 OK
```

To make post-commit hooks effective, you should [config your svn repository](#Config-your-svn-repository).

#### Config your svn repository
You can set up a post commit hook so the Subversion repository can notify cyclone whenever a change is made to that repository. To do this, put the following script in your post-commit file (in the $REPOSITORY/hooks directory):
```
REPOS="$1"
REV="$2"
TXN_NAME="$3"

UUID=`svnlook uuid $REPOS`

/usr/bin/curl --request POST --header "Content-Type:text/plain;charset=UTF-8" \
  --data "`svnlook changed --revision $REV $REPOS`" \
  {cyclone-server-address}/api/v1/subversion/$UUID/postcommithook?revision=$REV
```

Notes:

Replace `{cyclone-server-address}` by the actural cyclone server address value.

### PipelineStatusStatsObject

```
{
    "overview": {
        "total": "500",                       // int, required
        "success": "200",                     // int, required
        "failed": "200",                      // int, required
        "aborted": "200",                     // int, required
        "successRatio": "40.00%"              // ratio, required
    },
    "details": {
        {
            "timestamp": "1521555816",        // int, required
            "success": "10",                  // int, required
            "failed": "20",                   // int, required
            "aborted": "0"                    // int, required
        },
        {
            "timestamp": "1521555816",        // int, required
            "success": "10",                  // int, required
            "failed": "20",                   // int, required
            "aborted": "0"                    // int, required
        },
        {
            "timestamp": "1521555816",        // int, required
            "success": "10",                  // int, required
            "failed": "20",                   // int, required
            "aborted": "0"                    // int, required
        }
    }
}
```

Note:

| Field | Note |
| --- | --- |
| timestamp | Required, Unix timestamp format, counted as 0:0:0 in that day |

### Get Project Stats

Get the stats for all pipelines in this project.

**Request**

URL: `GET /api/v1/projects/{project}/stats?[startTime=&endTime=]`

Note:

| Field | Note |
| --- | --- |
| startTime | Required, Unix timestamp format, start from 0:0:0 |
| endTime | Required, Unix timestamp format, start from 23:59:59 |

**Response**

Success:

```
200 OK

{
    // all <PipelineStatusStatsObject> fields
}
```

### Get Pipeline Stats

Get the execution stats for the pipeline.

**Request**

URL: `GET /api/v1/projects/{project}/pipelines/{pipeline}/stats?[startTime=&endTime=]`

Note:
Illustration of startTime and endTime ditto.

**Response**

Success:

```
200 OK

{
    // all <PipelineStatusStatsObject> fields
}
```


### CloudObject

```
{
    "id": "123",                                              // string, optional.
    "type": "Docker",                                         // string, required.
    "name": "cluster1",                                       // string, required.
    "docker": {
        "host": "unix:///var/run/docker.sock",                // string, required.
        "certPath": "/etc/docker",                            // string, optional.
        "insecure": false                                     // bool, optional.
    },
    "kubernetes": {
        "host": "1.2.3.4",                                    // string, required.
        "namespace": "cyclone-worker",                        // string, required.
        "bearerToken": "asdlfjaslfjewasdfasdf"                // string, optional.
        "username": "root",                                   // string, optional.
        "password": "123456",                                 // string, optional.
        "inCluster": true,                                    // bool, optional.
        "tlsClientConfig": {                                  // optional.
            "insecure": false,                                // bool, optional.
            "caFile":"/var/run/secrets/ca.crt",               // string, optional.
            "caData":""                                       // []byte, optional.
        }
    }
    "creationTime": "2016-04-26T05:21:13.140Z",               // string, required
    "lastUpdateTime": "2016-04-26T05:21:13.140Z"              // string, required
}
```

Note:

| Field | Note |
| --- | --- |
| type | Required, cloud type, support Docker and Kubernetes. |
| namespace | Required, k8s namespace for worker |
| bearerToken | Optional, must use one of bearerToken and username/password as auth method. |

### Create cloud

Add cloud。

**Request**

URL: `POST /api/v1/clouds`

Body:

```
    // all <CloudObject> fields
```

**Response**

Success:

```
201 OK

{
    // all <CloudObject> fields
}
```

### List clouds

List all clouds.

**Request**

URL: `GET /api/v1/clouds`

**Response**

Success:

```
200 OK

{
    // all <CloudObject> fields
}
{
    "metadata": <PaginationObject>,
    "items": [ <CloudObject>, ... ]
}
```

### Get cloud

Get cloud info。

**Request**

URL: `GET /api/v1/clouds/{cloud}`

**Response**

Success:

```
200 OK

{
    // all <CloudObject> fields
}
```

### Update cloud

Update the cloud config.

**Request**

URL: `PUT /api/v1/clouds/{cloud}`

Body:

```
{
    // all <CloudObject> fields
}
```

**Response**

Success:

```
200 OK

{
    // all <CloudObject> fields
}
```

### Delete cloud

Delete the cloud.

**Request**

URL: `DELETE /api/v1/clouds/{cloud}`

**Response**

Success:

```
204 No Content
```

### WorkerInstance
```
{
    "name": "cyclone-worker-5b30a4e5488571000107261f",
    "status": "Running",
    "creationTime": "2018-06-25T16:16:37+08:00",
    "lastUpdateTime": "2018-06-25T16:16:37+08:00",
    "projectName": "project1",
    "pipelineName" "pipeline1",
    "recordID": "abcdef"
}
```

### List cyclone workers

List all cyclone workers.

**Request**

URL: `GET /api/v1/clouds/{cloud}/workers[namespace=nsvalue]`

Args:

| Name | Type | Detail |
| --- | --- | --- |
| namespace | string, required when cloud type is `Kubernetes` | Specify k8s namespace. |

**Response**

Success:

```
200 OK

{
    "metadata": {
        "total": 0,            // number, always
    },
    "items": [ <WorkerInstance>, ... ]
}
```

### ConfigTemplateObject

```
{
    "name": "python2.7",                                      // string, required.
    "type": "maven",                                          // string, required.
    "builderImage": "python:2.7-alpine",                      // string, required.
    "testCommands": "",                                       // string
    "packageCommands": "2016-04-26T05:21:13.140Z",            // string
    "customizedDockerfile": "Dockerfile contents"             // string
}
```


Note:

| Field | Note |
| --- | --- |
| type | Required, pipeline config template type, Go, Maven, Gradle, NodeJS, Python and PHP. |

### List config templates

List all cyclone pipeline config templates.

**Request**

URL: `GET /api/v1/configtemplates`

**Response**

Success:

```
200 OK

{
    "metadata": {
        "total": 0,            // number, always
    },
    "items": [ <ConfigTemplateObject>, ... ]
}
```

### IntegrationObject

```
{
    "name": "sonar1",  // can not update
    "alias": "alias",
    "type": "SonarQube",  // can not update
    "sonarqube": {
        "description": "This is my first sonar qube instance.",
        "address": "http://192.168.21.100:9000",
        "token": "f399878566d5d6a3de1759222a4b5eb15cac51de",
        "user":"" // Optional, stored for front-end to display.
    },
    "creationTime": "2017-08-23T09:44:08.653Z",
    "lastUpdateTime": "2017-08-23T09:44:08.653Z"
}
```

### Create integration

Add integration.

**Request**

URL: `POST /api/v1/integrations`

Body:

```
    // all <IntegrationObject> fields
```

**Response**

Success:

```
201 OK

{
    // all <CloudObject> fields
}
```

### List integrations

List all integrations.

**Request**

URL: `GET /api/v1/integrations`

**Response**

Success:

```
200 OK

{
    "metadata": {
        "total": 0,
    },
    "items": [ <IntegrationObject>, ... ]
}
```


### Get integration

Get integration information.

**Request**

URL: `GET /api/v1/integrations/{integration}`

**Response**

Success:

```
200 OK

{
    // all <IntegrationObject> fields
}
```

### Update integration

Update the integration information.

**Request**

URL: `PUT /api/v1/integrations/{integration}`

Body:

```
{
    // all <IntegrationObject> fields
}
```

**Response**

Success:

```
200 OK

{
    // all <IntegrationObject> fields
}
```

### Delete integration

Delete the integration.

**Request**

URL: `DELETE /api/v1/integrations/{integration}`

**Response**

Success:

```
204 No Content
```