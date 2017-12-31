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
  - [API Common](#api-common)
    - [Path Parameter Explanation](#path-parameter-explanation)
  - [API Details](#api-details)
    - [ProjectDataStructure](#projectdatastructure)
    - [List projects](#list-projects)
    - [Create project](#create-project)
    - [Get project](#get-project)
    - [Update project](#update-project)
    - [Delete project](#delete-project)
    - [PipelineDataStructure](#pipelinedatastructure)
    - [PipelineRecordDataStructure](#pipelinerecorddatastructure)
    - [ListedPipelineDataStructure](#listedpipelinedatastructure)
    - [List pipelines](#list-pipelines)
    - [Create pipeline](#create-pipeline)
    - [Get Pipeline](#get-pipeline)
    - [Update pipeline](#update-pipeline)
    - [Delete pipeline](#delete-pipeline)
    - [PipelinePerformParamsDataStructure](#pipelineperformparamsdatastructure)
    - [List pipeline records](#list-pipeline-records)
    - [Create pipeline record](#create-pipeline-record)
    - [Get pipeline record](#get-pipeline-record)
    - [Delete pipeline record](#delete-pipeline-record)
    - [Update Pipeline Record Status](#update-pipeline-record-status)
    - [Get Pipeline Record Log](#get-pipeline-record-log)
    - [Get Realtime Pipeline Record Log](#get-realtime-pipeline-record-log)
    - [RepoObjectDataStructure](#repoobjectdatastructure)
    - [List SCM Repos](#list-scm-repos)
    - [List SCM Branches](#list-scm-branches)

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

- [Project Data Structure](#ProjectDataStructure)

| API | Path | Detail |
| --- | --- | --- |
| List | GET `/api/v1/projects` | WIP, [link](#list-projects) |
| Create | POST `/api/v1/projects` | WIP, [link](#create-project) |
| Get | GET `/api/v1/projects/{project}` | WIP, [[link](#get-project) |
| Update | PATCH `/api/v1/projects/{project}` | WIP, [link](#update-project) |
| Delete | DELETE `/api/v1/projects/{project}` | WIP, [link](#delete-project) |

### Pipeline APIs

- [Pipeline Data Structure](#PipelineDataStructure)
- [Pipeline Record  Data Structure](#pipelinerecorddatastructure)
- [Listed Pipeline  Data Structure](#listedpipelinedatastructure)

| API | Path | Detail |
| --- | --- | --- |
| List | GET `/api/v1/projects/{project}/pipelines` | WIP, [link](#list-pipelines) |
| Create | POST `/api/v1/projects/{project}/pipelines` | WIP, [link](#create-pipeline) |
| Get | GET `/api/v1/projects/{project}/pipelines/{pipeline}` | WIP, [link](#get-pipeline) |
| Update | PUT `/api/v1/projects/{project}/pipelines/{pipeline}` | WIP, [link](#update-pipeline) |
| Delete | DELETE `/api/v1/projects/{project}/pipelines/{pipeline}` | WIP, [link](#delete-pipeline) |

### Pipeline Record APIs

- [Pipeline Record Data Structure](#PipelineRecordDataStructure)

| API | Path | Detail |
| --- | --- | --- |
| List | GET `/api/v1/projects/{project}/pipelines/{pipeline}/records` | WIP, [link](#list-pipeline-records) |
| Create | POST `/api/v1/projects/{project}/pipelines/{pipeline}/records` | WIP, [link](#create-pipeline-record) |
| Get | GET `/api/v1/projects/{project}/pipelines/{pipeline}/records/{recordID}` | WIP, [link](#get-pipeline-record) |
| Delete | DELETE `/api/v1/projects/{project}/pipelines/{pipeline}/records/{recordID}` | WIP, [link](#delete-pipeline-record) |
| Update Status | PATCH `/api/v1/projects/{project}/pipelines/{pipeline}/records/{recordID}/status` | WIP, [link](#update-pipeline-record-status) |

### Pipeline Record Logs API

| API | Path | Detail |
| --- | --- | --- |
| Get | GET `/api/v1/projects/{project}/pipelines/{pipeline}/records/{recordID}/logs` | WIP, [link](#get-pipeline-record-log) |
| Get | GET `/api/v1/projects/{project}/pipelines/{pipeline}/records/{recordID}/logstream` | WIP, [link](#get-realtime-pipeline-record-log) |

### SCM API

- [Repo Object Data Structure](#RepoObjectDataStructure)

| API | Path | Detail |
| --- | --- | --- |
| List | GET `/api/v1/workspaces/{workspace}/repos` | WIP, [link](#list-scm-repos) |
| List | GET `/api/v1/workspaces/{workspace}/branches?repo=` | WIP, [link](#list-scm-branches) |

## API Common

### Path Parameter Explanation

In path parameter, both `{project}` and `{pipiline}` are `name`, and only `{recordId}` is `ID`

## API Details

### ProjectDataStructure

```
{
  "name": "string",                             // name of the project, should be unique
  "description": "string",                      // description of the project
  "owner": "string"                             // owner of the project
  "scm": {                                      // required
      "type": "",                               // string, required. Only support "Gitlab" and "Github".
      "server": "",                             // string, required
      "username": "",                           // string, optional
      "password": "",                           // string, optional
      "token": "",                              // string, optional
  },
  "creationTime": "2017-08-23T09:44:08.653Z",   // created time of the project
  "lastUpdateTime": "2017-08-23T09:44:08.653Z", // updated time of the project
}
```

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
    "items": [ <ProjectDataStructure>, ... ]
}
```

### Create project

Create a new project.

**Request**

URL: `POST /api/v1/projects`

Body:

```
{
    "description": "",                   // string, optional
    "name": "",                          // string, required
    "scm": {                             // required
      "type": "",                        // string, required. Only support "Gitlab" and "Github".
      "server": "",                      // string, required
      "username": "",                    // string, optional
      "password": "",                    // string, optional
      "token": "",                       // string, optional
    }
}
```

Note:

| Field | Note |
| --- | --- |
| name | ^[a-z0-9]+((?:[._-][a-z0-9]+)*){1,29}$ |
| description | The length is limited to 100 characters |
| type | Supports `Gitlab` and `Github` |

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
    // all <ProjectDataStructure> fields
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
    // all <ProjectDataStructure> fields
}
```

### Update project

Update the information of a project.

**Request**

URL: `PATCH /api/v1/projects/{project}`

Body:

```
{
    "description": "New description.",  // string, required
}
```

Note:

| Field | Note |
| --- | --- |
| description | The length is limited to 100 characters |

**Response**

Success:

```
200 OK

{
    // all <ProjectDataStructure> fields
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

### PipelineDataStructure

```
{
    "id": "string", 
    "name": "string", 
    "alias": "string", 
    "description": "string", 
    "owner": "string", 
    "projectID": "string", 
    "serviceID": "string",                  // workaround, remove soon
    "notification": {                       // not implemented
        "emailNotification": {
            "emails": [
                "string"
            ]
        }, 
        "notificationPolicy": "string"
    }, 
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
            "commitTrigger": {
                "stages": ["string", ...]
            }, 
            "commitWithCommentsTrigger": {
                "comments": ["string", ...], 
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
                "codeSources": [                      // only support one element
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
                        "type": "string"              // type of code source, it must be github or gitlab or other.
                    }
                ]
            }, 
            "unitTest": {
                "command": ["string", ...], 
                "outputs": ["string", ...]
            }, 
            "codeScan": {
                "command": ["string", ...], 
                "outputs": ["string", ...]
            }, 
            "package": {
                "command": ["string", ...], 
                "outputs": ["string", ...]
            }, 
            "imageBuild": {
                "buildInfos": [                       // only support one element
                    {
                        "contextDir": "string", 
                        "dockerfile": "string", 
                        "dockerfilePath": "string", 
                        "imageName": "string"
                    }
                ]
            }, 
            "imageRelease": {
                "releasePolicy": [                    // image names which are not listed will be ignored.
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
    "creationTime": "2017-08-23T08:40:33.764Z", 
    "lastUpdateTime": "2017-08-23T08:40:33.764Z"
}
```

Pipeline is responsible for automating the lifecycle management of an application, and can safely and reliably deploy the application from the source code to the production environment in strict accordance with a set of scientific and rational software management processes.

### PipelineRecordDataStructure

```
{
    "id": "",                   // string, required
    "name": "",                 // string, required
    "pipelineID": "",           // string, required
    "versionID": "",            // string, required
    "trigger": "",              // string, optional. Can be user, scmTrigger, cronTrigger
    "performParams": {
        "ref": "master",                  // string, required
        "name": "newVersion",             // string, optional
        "description": "",                // string, optional
        "createScmTag": false,            // bool, optional
        "stages": ["codeCheckout", "unitTest", ...]     // string array, optional. Elements can be codeCheckout、unitTest、codeScan、package、imageBuild、integrationTest and imageRelease
    },
    "status": "",               // string, required. Can be Pending, Running, Success, Failed, Aborted.
    "startTime": "2017-08-23T08:40:33.764Z",        // time, required
    "endTime": "2017-08-23T08:40:33.764Z",          // time, optional
}
```

### ListedPipelineDataStructure

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
    "items": [ <PipelineDataStructure>, ... ]
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
    // all <PipelineDataStructure> fields
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
    // all <PipelineDataStructure> fields
}
```

### Update pipeline

Update the information of a pipeline.

**Request**

URL: `PUT /api/v1/projects/{project}/pipelines/{pipeline}`

Body:

```
{
    // all <PipelineDataStructure> fields
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
    // all <PipelineDataStructure> fields
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

### PipelinePerformParamsDataStructure

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
    "items": [ <PipelineRecordDataStructure>, ... ]
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
    // all <PipelineRecordDataStructure> fields
}
```

### Get pipeline record

Get the pipeline execution information.

**Request**

URL: `GET /api/v1/projects/{project}/pipelines/{pipeline}/records/{recordID}`

**Response**

Success:

```
200 OK

{
    // all <PipelineRecordDataStructure> fields
}
```

### Delete pipeline record

Delete a pipeline record.

**Request**

URL: `DELETE /api/v1/projects/{project}/pipelines/{pipeline}/records/{recordID}`

**Response**

Success:

```
204 NoContent
```

### Update Pipeline Record Status

Update the pipeline record status. Only the pipeline status of the Running state can be changed to Aborted, which means the pipeline is terminated.

**Request**

URL: `PATCH /api/v1/projects/{project}/pipelines/{pipeline}/records/{recordID}/status`

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
    // all <PipelineRecordDataStructure> fields
}
```

### Get Pipeline Record Log

Get the logs of finished pipeline records.

**Request**

URL: `GET /api/v1/projects/{project}/pipelines/{pipeline}/records/{recordID}/logs[?download=]`

**Response**

Success:

```
200 OK

Content-Type: text/plain

step: clone repository state: start
Cloning into 'code'...
step: clone repository state: finish
step: Parse Yaml state: start
step: Parse Yaml state: finish
step: Pre Build state: start
$ echo hello
hello
step: Pre Build state: finish
step: Build image state: start
Step 1 : FROM cargo.caicloud.io/caicloud/cyclone-worker:latest
 ---> 2437d0db0a28
Step 2 : ADD ./README.md /README.md
  ---> e8ca485e1ab8
......
```

Note:

This API can be called only after the pipeline records finish.

| Field | Note |
| --- | --- |
| download | true: download logs, the log file name is {projectName}-{pipelineName}-{recordId}-log.txt; false: directly return logs. `False` in default. |

### Get Realtime Pipeline Record Log

Get the real-time logs for running pipeline records.

**Request**

URL: `GET /api/v1/projects/{project}/pipelines/{pipeline}/records/{recordID}/logstream`

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

step: clone repository state: start
Cloning into 'code'...
step: clone repository state: finish
step: Parse Yaml state: start
step: Parse Yaml state: finish
step: Pre Build state: start
$ echo hello
hello
step: Pre Build state: finish
step: Build image state: start
Step 1 : FROM cargo.caicloud.io/caicloud/cyclone-worker:latest
 ---> 2437d0db0a28
Step 2 : ADD ./README.md /README.md
  ---> e8ca485e1ab8
......
```

### RepoObjectDataStructure

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
