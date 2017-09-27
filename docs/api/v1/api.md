<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Cyclone API V1](#cyclone-api-v1)
  - [Endpoint](#endpoint)
  - [API Summary](#api-summary)
    - [Project APIs](#project-apis)
    - [Pipeline APIs](#pipeline-apis)
    - [Pipeline Record APIs](#pipeline-record-apis)
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

## API Common

### Path Parameter Explanation

In path parameter, both `{project}` and `{pipiline}` are `name`, and only `{recordId}` is `ID`

## API Details

### ProjectDataStructure

```
{
  "creationTime": "2017-08-23T09:44:08.653Z",   // created time of the project
  "description": "string",                      // description of the project
  "lastUpdateTime": "2017-08-23T09:44:08.653Z", // updated time of the project
  "name": "string",                             // name of the project, should be unique
  "owner": "string"                             // owner of the project
}
```

Project is responsible for managing a set of applications, in this system, mainly managing a number of pipelines corresponding to this set of applications.

### List projects

List all projects.

**Request**

Http method: `GET`

URL: `/api/v1/projects[?start=&limit=]`

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

Http method: `POST`

URL: `/api/v1/projects`

Body:

```
{
    "description": "",                   // string, optional
    "name": "",                          // string, required
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
201 Created

{
    // all <ProjectDataStructure> fields
}
```

### Get project

Get the information of a project.

**Request**

Http method: `GET`

URL: `/api/v1/projects/{project}`

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

Http method: `PATCH`

URL: `/api/v1/projects/{project}`

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

Http method: `DELETE`

URL: `/api/v1/projects/{project}`

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

Http method: `GET`

URL: `/api/v1/workspaces/{workspace}/pipelines[?start=&limit=&recentCount=&recentSuccessCount=&recentFailedCount=]`

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

Http method: `POST`

URL: `/api/v1/projects/{project}/pipelines`

Body:

```
{
    "name": "newworkspace",             // string, required
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

Http method: `PUT`

URL: `/api/v1/workspaces/{workspace}/pipelines/{pipeline}`

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

Http method: `PUT`

URL: `/api/v1/projects/{project}/pipelines/{pipeline}`

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

Http method: `DELETE`

URL: `/api/v1/projects/{project}/pipelines/{pipeline}`

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

Http method: `POST`

URL: `/api/v1/projects/{project}/pipelines/{pipeline}/records[?start=&limit=]`

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

Careate a pipeline record, which means trigger a pipeline.

**Request**

URL: `POST /api/v1/workspaces/{workspace}/pipelines/{pipeline}/records`

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

Http method: `GET`

URL: `/api/v1/projects/{project}/pipelines/{pipeline}/records/{recordId}`

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

Http method: `DELETE`

URL: `/api/v1/projects/{project}/pipelines/{pipeline}/records/{recordId}`

**Response**

Success:

```
204 NoContent
```

### Update Pipeline Record Status

Update the pipeline record status. Only the pipeline status of the Running state can be changed to Aborted, which means the pipeline is terminated.

**Request**

Http method: `PATCH`

URL: `/api/v1/workspaces/{workspace}/pipelines/{pipeline}/records/{recordID}/status`

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
    // all <WorkspaceObject> fields
}
```