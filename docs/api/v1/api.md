<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Cyclone API V1](#cyclone-api-v1)
  - [API Summary](#api-summary)
    - [Project APIs](#project-apis)
    - [Pipeline APIs](#pipeline-apis)
    - [Pipeline Record APIs](#pipeline-record-apis)
  - [API Common](#api-common)
    - [Prefix](#prefix)
    - [Path Parameter Explanation](#path-parameter-explanation)
    - [Client Error Response](#client-error-response)
  - [API Details](#api-details)
    - [ProjectDataStructure](#projectdatastructure)
    - [Add a project](#add-a-project)
      - [Request](#request)
      - [Response](#response)
    - [Get all projects](#get-all-projects)
      - [Request](#request-1)
      - [Response](#response-1)
    - [Update a project](#update-a-project)
      - [Request](#request-2)
      - [Response](#response-2)
    - [Get a project](#get-a-project)
      - [Request](#request-3)
      - [Response](#response-3)
    - [Delete a project](#delete-a-project)
      - [Request](#request-4)
      - [Response](#response-4)
    - [PipelineDataStructure](#pipelinedatastructure)
    - [Add a pipeline](#add-a-pipeline)
      - [Request](#request-5)
      - [Response](#response-5)
    - [Get all pipelines](#get-all-pipelines)
      - [Request](#request-6)
      - [Response](#response-6)
    - [Update a pipeline](#update-a-pipeline)
      - [Request](#request-7)
      - [Response](#response-7)
    - [Get a pipeline](#get-a-pipeline)
      - [Request](#request-8)
      - [Response](#response-8)
    - [Delete a pipeline](#delete-a-pipeline)
      - [Request](#request-9)
      - [Response](#response-9)
    - [Trigger a pipeline](#trigger-a-pipeline)
      - [Request](#request-10)
      - [Response](#response-10)
    - [PipelineRecordDataStructure](#pipelinerecorddatastructure)
    - [Get all pipeline records](#get-all-pipeline-records)
      - [Request](#request-11)
      - [Response](#response-11)
    - [Get a pipeline record](#get-a-pipeline-record)
      - [Request](#request-12)
      - [Response](#response-12)
    - [Delete a pipeline record](#delete-a-pipeline-record)
      - [Request](#request-13)
      - [Response](#response-13)
    - [Get log of a pipeline record](#get-log-of-a-pipeline-record)
      - [Request](#request-14)
      - [Response](#response-14)

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

| Method | SubPath | Request Body | Response Body | function | Detail |
| --- | --- | --- | --- | --- | --- |
| POST | /projects | Project | Project | Add a project | WIP, [link](#add-a-project) |
| GET | /projects |     | Project list | Get all projects | WIP, [link](#get-all-projects) |
| PUT | /projects/{project} | Project | Project | Update a project | WIP, [link](#update-a-project) |
| GET | /projects/{project} |     | Project | Get a project | WIP, [link](#get-a-project) |
| DELETE | /projects/{project} |     |     | Delete a project | WIP, [link](#delete-a-project) |

### Pipeline APIs

- [Pipeline Data Structure](#PipelineDataStructure)

| Method | SubPath | Request Body | Response Body | function | Detail |
| --- | --- | --- | --- | --- | --- |
| POST | /projects/{project}/pipelines | Pipeline</br> * name: required</br> * build.builderImage: required</br> * build.stages: required| Pipeline | Add a pipeline | WIP, [link](#add-a-pipeline) |
| GET | /projects/{project}/pipelines |     | Pipeline list | Get all pipelines | WIP, [link](#get-all-pipelines) |
| PUT | /projects/{project}/pipelines/{pipeline} | Pipeline</br>| Pipeline | Update a pipeline | WIP, [link](#update-a-pipeline) |
| GET | /projects/{project}/pipelines/{pipeline} |     | Pipeline | Get a pipeline | WIP, [link](#get-a-pipeline) |
| DELETE | /projects/{project}/pipelines/{pipeline} |     |     | Delete a pipeline | WIP, [link](#delete-a-pipeline) |
| POST | /projects/{project}/pipelines/{pipeline} | version:string</br>description:string</br>createTag:bool</br>stages:[]string</br> |     | Trigger a pipeline | WIP, [link](#trigger-a-pipeline) |
| POST | /projects/{project}/pipelines/{pipeline}/gitLabWebhook |     |     | Trigger a pipeline by gitlab webhook | trigger by gitlab |
| POST | /projects/{project}/pipelines/{pipeline}/gitHubWebhook |     |     | Trigger a pipeline by github webhook | trigger by github |

### Pipeline Record APIs

- [Pipeline Record Data Structure](#PipelineRecordDataStructure)

| Method | SubPath | Request Body | Response Body | function | Detail |
| --- | --- | --- | --- | --- | --- |
| GET | /projects/{project}/pipelines/{pipeline}/records |     | Pipeline record list | Get all pipeline records | WIP, [link](#get-all-pipeline-records) |
| GET | /projects/{project}/pipelines/{pipeline}/records/{recordId} |     | Pipeline record | Get a pipeline record | WIP, [link](#get-a-pipeline-record) |
| DELETE | /projects/{project}/pipelines/{pipeline}/records/{recordId} |     |     | Delete a pipeline record | WIP, [link](#delete-a-pipeline-record) |
| GET | /projects/{project}/pipelines/{pipeline}/records/{recordId}/logs |     | string | Get log of a pipeline record | WIP, [link](#get-log-of-a-pipeline-record) |

## API Common

### Prefix

Add `/api/v1/` ahead, complete path: `/api/v1/` + `SubPath`

### Path Parameter Explanation

In path parameter, both `{project}` and `{pipiline}` are `name`, and only `{recordId}` is `ID`

### Client Error Response

```
400 Bad Request

{
    "message": "a human readable debug message",
    "code": "E0001",
    "data": <DataForTemplateRendering>
}
```

## API Details

### ProjectDataStructure

```json
{
  "createdTime": "2017-08-23T09:44:08.653Z", 
  "description": "string",                   // string, optional
  "id": "string",                            
  "name": "string",                          // string, always
  "owner": "string",                        
  "updatedTime": "2017-08-23T09:44:08.653Z"  
}
```

Project is responsible for managing a set of applications, in this system, mainly managing a number of pipelines corresponding to this set of applications.

### Add a project

Create a new project.

#### Request

Http method: `POST`

URL: `/api/v1/projects`

Body:

```
{
    "description": "",                   // string, optional
    "name": "",                          // string, required
    "owner": ""                          // string, required
}
```

#### Response

Success:

```
201 Created

{
    // all <ProjectDataStructure> fields
}
```

### Get all projects

List all projects.

#### Request

Http method: GET

URL: `/api/v1/projects[?start=&limit=]`

Args:

| Name | Type | Detail |
| --- | --- | --- |
| start | number, optional | The starting position of the query. |
| limit | number, optional | The quantity limit for the result. |

#### Response

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

### Update a project

Update the information of a project.

#### Request

Http method: PUT

URL: `/api/v1/projects/{project}`

Body:

```
{
    "description": "",                   // string, optional
    "name": "",                          // string, optional
}
```

#### Response

Success:

```
200 OK

{
    // all <ProjectDataStructure> fields
}
```

### Get a project

Get the information of a project.

#### Request

Http method: GET

URL: `/api/v1/projects/{project}`

#### Response

Success:

```
200 OK

{
    // all <ProjectDataStructure> fields
}
```

### Delete a project

Delete a project.

#### Request

Http method: DELETE

URL: `/api/v1/projects/{project}`

#### Response

Success:

```
203 NoContent
```

### PipelineDataStructure

```json
{
    "autoTrigger": {
        "cronTrigger": {
            "crons": [
                {
                    "expression": "string", 
                    "finalStage": "string"
                }
            ]
        }, 
        "scmTrigger": {
            "commitTrigger": {
                "finalStage": "string"
            }, 
            "commitWithCommentsTrigger": {
                "comments": [
                    "string"
                ], 
                "finalStage": "string"
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
                "codeSources": [
                    {
                        "gitHub": {
                            "branch": "string", 
                            "url": "string"
                        }, 
                        "gitLab": {
                            "branch": "string", 
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
            "codeScan": {
                "command": [
                    "string"
                ], 
                "outputs": [
                    "string"
                ]
            }, 
            "imageBuild": {
                "buildInfos": [
                    {
                        "contextDir": "string", 
                        "dockerfile": "string", 
                        "dockerfilePath": "string", 
                        "imageName": "string"
                    }
                ]
            }, 
            "imageRelease": {
                "releasePolicy": [
                    {
                        "image": "string", 
                        "type": "string"                // type of image release policy, it must be Always or Never or IntegrationTestSuccess.
                    }
                ]
            }, 
            "integrationTest": {
                "integrationTestSet": {
                    "command": [
                        "string"
                    ], 
                    "image": "string"
                }, 
                "services": [
                    {
                        "command": [
                            "string"
                        ], 
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
            }, 
            "package": {
                "command": [
                    "string"
                ], 
                "outputs": [
                    "string"
                ]
            }, 
            "unitTest": {
                "command": [
                    "string"
                ], 
                "outputs": [
                    "string"
                ]
            }
        }
    }, 
    "createdTime": "2017-08-23T08:40:33.764Z", 
    "description": "string", 
    "id": "string", 
    "name": "string", 
    "notification": {
        "emailNotification": {
            "emails": [
                "string"
            ]
        }, 
        "notificationPolicy": "string"
    }, 
    "owner": "string", 
    "projectID": "string", 
    "serviceID": "string", 
    "updatedTime": "2017-08-23T08:40:33.764Z"
}
```

Pipeline is responsible for automating the lifecycle management of an application, and can safely and reliably deploy the application from the source code to the production environment in strict accordance with a set of scientific and rational software management processes.

### Add a pipeline

Create a new pipeline.

#### Request

Http method: `POST`

URL: `/api/v1/projects/{project}/pipelines`

Body:

```
{
    // all <PipelineDataStructure> fields
}
```

#### Response

Success:

```
201 Created

{
    // all <PipelineDataStructure> fields
}
```

### Get all pipelines

List all pipelines of one project.

#### Request

Http method: GET

URL: `/api/v1/projects/{project}/pipelines[?start=&limit=]`

Args:

| Name | Type | Detail |
| --- | --- | --- |
| start | number, optional | The starting position of the query. |
| limit | number, optional | The quantity limit for the result. |

#### Response

Success:

```
200 OK

{
    "metadata": {
        "total": 0,            // number, always 
    },
    "items": [ <PipelineDataStructure>, ... ]
}
```

### Update a pipeline

Update the information of a pipeline.

#### Request

Http method: PUT

URL: `/api/v1/projects/{project}/pipelines/{pipeline}`

Body:

```
{
    // all <PipelineDataStructure> fields
}
```

#### Response

Success:

```
200 OK

{
    // all <PipelineDataStructure> fields
}
```

### Get a pipeline

Get the information of a pipeline.

#### Request

Http method: GET

URL: `/api/v1/projects/{project}/pipelines/{pipeline}`

#### Response

Success:

```
200 OK

{
    // all <PipelineDataStructure> fields
}
```

### Delete a pipeline

Delete a pipeline.

#### Request

Http method: DELETE

URL: `/api/v1/projects/{project}/pipelines/{pipeline}`

#### Response

Success:

```
203 NoContent
```

### Trigger a pipeline

Trigger a pipeline.

#### Request

Http method: `POST`

URL: `/api/v1/projects/{project}/pipelines`

Body:

```json

```

#### Response

Success:

```
201 Created

{
    // string
}
```

### PipelineRecordDataStructure

```json
{
    "endTime": "2017-08-23T10:52:19.610Z", 
    "id": "string", 
    "pipelineID": "string", 
    "stageStatus": {
        "codeCheckout": {
            "endTime": "2017-08-23T10:52:19.610Z", 
            "startTime": "2017-08-23T10:52:19.610Z", 
            "status": "string"
        }, 
        "codeScan": {
            "endTime": "2017-08-23T10:52:19.610Z", 
            "startTime": "2017-08-23T10:52:19.610Z", 
            "status": "string"
        }, 
        "imageBuild": {
            "endTime": "2017-08-23T10:52:19.610Z", 
            "startTime": "2017-08-23T10:52:19.610Z", 
            "status": "string"
        }, 
        "imageRelease": {
            "endTime": "2017-08-23T10:52:19.610Z", 
            "startTime": "2017-08-23T10:52:19.610Z", 
            "status": "string"
        }, 
        "integrationTest": {
            "endTime": "2017-08-23T10:52:19.610Z", 
            "startTime": "2017-08-23T10:52:19.610Z", 
            "status": "string"
        }, 
        "package": {
            "endTime": "2017-08-23T10:52:19.610Z", 
            "startTime": "2017-08-23T10:52:19.610Z", 
            "status": "string"
        }, 
        "unitTest": {
            "endTime": "2017-08-23T10:52:19.610Z", 
            "startTime": "2017-08-23T10:52:19.610Z", 
            "status": "string"
        }
    }, 
    "startTime": "2017-08-23T10:52:19.610Z", 
    "status": "string", 
    "trigger": "string", 
    "versionID": "string"
}
```

Args:

| Name | Detail |
| --- | --- |
| status | it can be "Pending" or "Running" or "Success" or "Failed" or "Aborted" |

Pipeline Record is an execution record of a pipeline that clearly records and displays detailed information about each stage of the application's lifecycle.

### Get all pipeline records

List all pipeline records of one pipeline.

#### Request

Http method: POST

URL: `/api/v1/projects/{project}/pipelines/{pipeline}/records[?start=&limit=]`

Args:

| Name | Type | Detail |
| --- | --- | --- |
| start | number, optional | The starting position of the query. |
| limit | number, optional | The quantity limit for the result. |

#### Response

Success:

```
200 OK

{
    "metadata": {
        "total": 0,            // number, always 
    },
    "items": [ <PipelineRecordDataStructure>, ... ]
}
```

### Get a pipeline record

Get the information of a pipeline.

#### Request

Http method: GET

URL: `/api/v1/projects/{project}/pipelines/{pipeline}/records/{recordId}`

#### Response

Success:

```
200 OK

{
    // all <PipelineRecordDataStructure> fields
}
```

### Delete a pipeline record

Delete a pipeline record.

#### Request

Http method: DELETE

URL: `/api/v1/projects/{project}/pipelines/{pipeline}/records/{recordId}`

#### Response

Success:

```
203 NoContent
```

### Get log of a pipeline record

Get log of a pipeline record.

#### Request

Http method: GET

URL: `/api/v1/projects/{project}/pipelines/{pipeline}/records/{recordId}/logs`

#### Response

Success:

```
200 OK

{
    // string
}
```
