#!/usr/bin/python

import requests
import json
import time
import multiprocessing

#host = 'http://127.0.0.1:7099/'
#host = 'http://43.254.54.38:7099/'
#host = 'http://118.193.143.243/'
host = 'http://139.196.115.69:7099/'
#host = 'https://fornax-canary.caicloud.io/'

headers = {
    "Accept": "application/json",
    "Content-type": "application/json;charset=utf-8",
}

data = {
    "name": "hello",
    "description": "test",
    "username": "test",
    "repository": {
        "url": "https://github.com/DanielWangZhanggui/cpu-load.git",
        "vcs": "git"
    },
    "build_path": ""
}

def health_check():
    url = host + 'api/v0.1/healthcheck'
    r = requests.get(url, headers=headers)
    print r.status_code, r.text
	
def get_event(event_id):
    url = host + 'api/v0.1/events/{event_id}'
    _headers = {
        "Accept": "application/json",
        "Content-type": "application/json;charset=utf-8",
        "token": event_id,
    }
    r = requests.get(url.format(**{'event_id': event_id}), headers=_headers)
    print r.status_code, r.text
	
def set_event(event_id, result, message):
    url = host + 'api/v0.1/events-result/{event_id}'
    _headers = {
        "Accept": "application/json",
        "Content-type": "application/json;charset=utf-8",
        "token": event_id,
    }
    _data = {
        "result" : result,
        "error_msg" : message
    }
    r = requests.post(url.format(**{'event_id': event_id}), headers=_headers, data=json.dumps(_data))
    print r.status_code, r.text

def create_service(user_id, service_name):
    _data = {
        "name": service_name,
        "description": "test",
        "username": "test",
        "repository": {
            #"url": "/home/superxi/gopath/src/github.com/caicloud/console-web",
            "url": "https://github.com/superxi911/cpu-load.git", 
            "vcs": "git", 
            #"webhook": "github"
            #"url": "svn://118.193.185.187/svn-demo/trunk", 
            #"vcs": "svn",
            #"username": "superxi",
            #"password": "xxxpwd"
        },
        "build_path": ""
    }
    url = host + 'api/v0.1/{user_id}/services'
    r = requests.post(url.format(**{'user_id': user_id}), headers=headers, data=json.dumps(_data))
    print r.status_code, r.text

def set_service(user_id, service_id):
    _data = {
        "description": "test",
        "repository": {
            "webhook": "github"
        },
        "profile": {
            "profiles": [{"email": "xxx@caicloud.io", "cellphone": "10086"}],
            "setting": "sendwhenfinished"
        },
        "deploy_plans": [
            {
                "plan_name": "plan1",
                "config": {
                    "cluster_id": "iii",
                    "cluster_name": "nnn",
                    "partition": "ppp",
                    "application": "aaa",
                    "containers": ["con1", "con2"]
                }
            },
            {
                "plan_name": "plan2",
                "config": {
                    "cluster_id": "iii",
                    "cluster_name": "nnn",
                    "partition": "ppp",
                    "application": "aaa",
                    "containers": ["con1", "con2"]
                }
            }
        ]
    }
    url = host + 'api/v0.1/{user_id}/services/{service_id}'
    r = requests.put(url.format(**{'user_id': user_id,'service_id': service_id}), headers=headers, data=json.dumps(_data))
    print r.status_code, r.text

def delete_service(user_id, service_id):
    url = host + 'api/v0.1/{user_id}/services/{service_id}'
    r = requests.delete(url.format(**{'user_id': user_id,'service_id': service_id}), headers=headers)
    print r.status_code, r.text

def get_services(user_id):
    url = host + 'api/v0.1/{user_id}/services'
    r = requests.get(url.format(**{'user_id': user_id}), headers=headers)
    print r.status_code, r.text

def get_service(user_id, service_id):
    url = host + 'api/v0.1/{user_id}/services/{service_id}'
    r = requests.get(url.format(**{'user_id': user_id,'service_id': service_id}), headers=headers)
    print r.status_code, r.text

def create_version(uid, service_id):
    url = host + 'api/v0.1/{uid}/versions'

    _data = {
        "name": "d89u8i5klsmhknh0lkbhchyajngidud5291by",
        "description": "v3",
        "service_id": service_id,
        #"operation": "publish"
        "operation": "integrationpublish"
    }

    r = requests.post(url.format(uid=uid), headers=headers, data=json.dumps(_data))
    print r.status_code, r.text

def get_version(user_id, service_id, version_id):
    url = host + 'api/v0.1/{user_id}/versions/{version_id}'
    r = requests.get(url.format(**{'user_id': user_id,'service_id': service_id,'version_id': version_id}), headers=headers)
    print r.status_code, r.text

def get_versions(user_id, service_id):
    url = host + 'api/v0.1/{user_id}/services/{service_id}/versions'
    r = requests.get(url.format(**{'user_id': user_id,'service_id': service_id}), headers=headers)
    print r.status_code, r.text

def cancel_build(user_id, version_id):
    url = host + 'api/v0.1/{user_id}/versions/{version_id}/cancelbuild'
    r = requests.post(url.format(**{'user_id': user_id,'version_id': version_id}), headers=headers)
    print r.status_code, r.text

def worker(num):
    """thread worker function"""
    print 'Worker:', num
    while 1:
        create_service('superxi', 'test_service', 'https://github.com/superxi911/console-web.git')
        time.sleep(0.1)
    return

def create_project(user_id, project_name):
    _data = {
        "name": project_name,
        "description": "test",
    }
    url = host + 'api/v0.1/{user_id}/projects'
    r = requests.post(url.format(**{'user_id': user_id}), headers=headers, data=json.dumps(_data))
    print r.status_code, r.text

def get_project(user_id, project_id):
    url = host + 'api/v0.1/{user_id}/projects/{project_id}'
    r = requests.get(url.format(**{'user_id': user_id, 'project_id': project_id}), headers=headers)
    print r.status_code, r.text

def get_projects(user_id):
    url = host + 'api/v0.1/{user_id}/projects'
    r = requests.get(url.format(**{'user_id': user_id}), headers=headers)
    print r.status_code, r.text

def set_project(user_id, project_id):
    url = host + 'api/v0.1/{user_id}/projects/{project_id}'
    _data = {
        "services" : [
            {
                "service_id" : "288664e6-55ea-41ef-afda-c8bf7bc8d8e9", 
                "depend" : [
                    {"service_id" : "94c936c6-ebc5-43bb-bf38-2772f431ec4e"}
                ]
            }, 
            {
                "service_id" : "288664e6-55ea-41ef-afda-c8bf7bc8d8e9", 
                "depend" : [
                    {"service_id" : "94c936c6-ebc5-43bb-bf38-2772f431ec4e"}
                ]
            }, 
            {
                "service_id" : "288664e6-55ea-41ef-afda-c8bf7bc8d8e9", 
                "depend" : [
                    {"service_id" : "94c936c6-ebc5-43bb-bf38-2772f431ec4e"}
                ]
            }
        ]
    }
    r = requests.put(url.format(**{'user_id': user_id,'project_id': project_id}), headers=headers, data=json.dumps(_data))
    print r.status_code, r.text

def delete_project(user_id, project_id):
    url = host + 'api/v0.1/{user_id}/projects/{project_id}'
    r = requests.delete(url.format(**{'user_id': user_id,'project_id': project_id}), headers=headers)
    print r.status_code, r.text

def create_project_version(user_id, project_id, version_name):
    _data = {
        "project_id": project_id,
        "policy": "manual", 
        "name": version_name, 
        "description": "test",
    }
    url = host + 'api/v0.1/{user_id}/versions_project'
    r = requests.post(url.format(**{'user_id': user_id}), headers=headers, data=json.dumps(_data))
    print r.status_code, r.text

def get_project_versions(user_id, project_id):
    url = host + 'api/v0.1/{user_id}/projects/{project_id}/versions'
    r = requests.get(url.format(**{'user_id': user_id,'project_id': project_id}), headers=headers)
    print r.status_code, r.text

def get_project_version(user_id, projectversion_id):
    url = host + 'api/v0.1/{user_id}/projectversions/{projectversion_id}'
    r = requests.get(url.format(**{'user_id': user_id,'projectversion_id': projectversion_id}), headers=headers)
    print r.status_code, r.text

def create_worker_node(node_name, docker_host):
    _data = {
        "name": node_name,
        "description": "test",
        "ip": "127.0.0.1",
        "docker_host": docker_host,
        "type": "system",
        "total_resource": {
            "memory": 5368709120, 
            "cpu": 10240,
        },
    }
    url = host + 'api/v0.1/system_worker_nodes'
    r = requests.post(url, headers=headers, data=json.dumps(_data))
    print r.status_code, r.text

def get_worker_node(node_id):
    url = host + 'api/v0.1/system_worker_nodes/{node_id}'
    r = requests.get(url.format(**{'node_id': node_id}), headers=headers)
    print r.status_code, r.text

def get_worker_nodes():
    url = host + 'api/v0.1/system_worker_nodes'
    r = requests.get(url, headers=headers)
    print r.status_code, r.text

def delete_worker_node(node_id):
    url = host + 'api/v0.1/system_worker_nodes/{node_id}'
    r = requests.delete(url.format(**{'node_id': node_id}), headers=headers)
    print r.status_code, r.text

if __name__ == '__main__':

    #health_check()
	
    #get_event('63e3a836-bdf9-4106-8c9a-a068b5e3a987')
    #set_event('63e3a836-bdf9-4106-8c9a-a068b5e3a987', 'success', 'well done')
	
    #create_service('superxi', 'test')
    service_id = 'f61a8986-5ddf-4d51-8eb3-a6cde8ed6260'
    #set_service('superxi', service_id)
    #get_service('superxi', '871e6da4-5a5d-4fd8-bb8e-166f817ff2c9')
    get_services('superxi')
    #delete_service("superxi", service_id)

    create_version('superxi', service_id)
    #get_versions('superxi', service_id)
    version_id = '9d148439-4c56-47e3-9712-23587e9b5c41'
    #cancel_build('superxi', version_id)
    #get_version('superxi', service_id, version_id)
    #for i in range(5):
    #    p = multiprocessing.Process(target=worker, args=(i,))
    #    p.start()

    #create_project("superxi", "test")
    #delete_project("superxi", "bf6afce1-be09-40ab-91d5-68a3129ba5c1")
    #set_project("superxi", "601780f2-266d-4072-9ff5-7b6b4ab4452b")
    #create_project_version("superxi", "1e7e3b93-a03b-4c18-88ff-ffdaddc65024", "v1")
    #get_projects("superxi")
    #get_project("superxi", "a0298c65-5237-4ec1-8af7-f4056e290adf")
    #get_project_versions("superxi", "e22144ee-ba66-4424-a405-054ddd749830")
    #get_project_version("superxi", "5d0a9b99-d75f-4848-89eb-7ae5b88dc20d")
	
    #create_worker_node("test1", "unix:///var/run/docker.sock")
    #create_worker_node("test2", "tcp://120.26.103.107:2375")
    #get_worker_node("e43c0207-f8f3-4513-bcff-c991c50dee74")
    #delete_worker_node("8211f1fc-2c40-42c2-90e9-b8d16f2d6336")
    #get_worker_nodes()
    pass
