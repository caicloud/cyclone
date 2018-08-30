
# Cargo-Admin  
  
Cargo-Admin is a service based on Harbor,  it manages docker images distributions for Compass. A set of REST APIs are provided. Cargo-Admin can manage multiple registries, and provide new features in addition to Harbor:  
  
- Multi-tenant support  
- Fine-grained token service for registry authorization  
- Image upload (by tarball or Dockerfile) and image copy

From overall view, Cargo-Admin consists two services: admin service and token service. The former manages projects and images and achieves tenant separation, while the latter provide token service for registries. The structure of Cargo-Admin is depicted in the following figure.

![structure.png](./docs/img/structure.png)
  
## Glossary & Concept  
  
**Harbor** : An enterprise-class container registry server based on Docker Distribution made by vmware. (Github: [vmware/harbor](https://github.com/vmware/harbor))  
  
**Cargo**: Cargo is a new deployment solution of Habor, with health check and high available accounted. Cargo will configure Habor to use token service provided by Cargo-Admin. Sometimes, we use _cargo_ and _harbor_ interchangeably.  
  
**Registry**: Container registry,  which stores and lets you distribute Docker images. We use *registry* and *harbor* interchangeably in cargo-admin.  
  
**Project**: Same concept to that in Harbor, it's a group of repositories. In other places, this is often refer to as `owner`, for example in docker hub, for repo `prom/busybox` , `prom` stands for `prometheus`, it's `owner` of the repo. And for official images like `busybox`, there is default owner (or project in our concept) `library`. Before version 2.7, we once use _workspace_ for this concept, and it's still appears in other places like `Cyclone`.  
  
**Repository**: _A repository potentially holds multiple variants of an image._ Give `prom/busybox:latest` as an example, `prom/busybox` is a repository, which can holds multiple images like `prom/busybox:latest`, `prom/busybox:v1.0`, etc. We can tag a repository to get an image. You can understand this from `docker images` results. Each line represent an image, and first column of each image is called _REPOSITORY_, different tags to a same repository makes different images.  
```  
$ docker images  
REPOSITORY        TAG               IMAGE ID             CREATED             SIZE  
prom/busybox     latest            e0b7a9ef61c6        23 hours ago         1.15MB  
prom/busybox     v1.0              6d7976764092        23 hours ago         1.15MB  
```  
  
**Image**: Docker images are the basis of [containers](https://docs.docker.com/glossary/?term=container). An Image is an ordered collection of root filesystem changes and the corresponding execution parameters for use within a container runtime. An image typically contains a union of layered filesystems stacked on top of each other. An image does not have state and it never changes. With registry, project, repository concept in mind, one full image name looks like the following. It consists of `domain` (corresponds to registry), `project`, `repo`, and `tag`.  
> cargo.caicloudprivatetest.com/caicloud/busybox:latest  
  
**Replication**: It's a concept introduced from Harbor. Replication is used to replicate images of one project from one registry to another registry. Replication in Cargo-Admin corresponds to replication policy in Harbor. It can have different trigger types: manually, schedule, immediately.  
- Manual: Replicate the repositories manually when needed. The deletion operations are not replicated.  
- Immediate: When a new repository is pushed to the project, it is replicated to the remote registry immediately. We also call this trigger type `On Push` in Cargo-Admin.  
- Scheduled: Replicate the repositories daily or weekly. The deletion operations are not replicated.  
  
**Target**: It's a concept used in replication, a target is another registry where images will be replicated to. It's also concept introduced from Harbor, targets are managed in harbor for replication.  
  
**Token**: Authorization service of registry access is provided by Cargo-Admin. When we push/pull an image from registry, an _service/token_ request is forwarded to Cargo-Admin with account information to obtain a valid access token. Then the token is used to access registry.  
  
## Configure

There two services in Cargo-Admin, so two different configure files are defined:
  
  _cargo-admin-config.json_
```json  
{  
   "address": "0.0.0.0:8080",  
   "system_tenant": "system-tenant",  
   "mongo": {  
      "addrs": "127.0.0.1:27017",  
      "db": "cargo_admin",  
      "mode": "strong"  
  },  
  "default_public_projects": [  
      {  
         "name": "library",  
         "if_exists": "force",  
         "harbor": "default"  
      },  
      {  
         "name": "release",  
         "if_exists": "force",  
         "harbor": "default"  
       }  
  ],  
  "default_registry": {  
      "host": "https://cargo.caicloudprivatetest.com",  
      "username": "admin",  
      "password": "Pwd123456"  
   }  
}
```
**address**: Endpoint of this cargo-admin service
  
**mongo**: Mongo database to be used
  
**system_tenant**: System tenant name, usually it's `system-tenant`
  
**default_public_projects**: Defines several default public projects, e.g. `library`, `release`  
  
**default_registry**: Define the default registry to manage.

_cargo-token-config.json_
```json
{  
   "address": "0.0.0.0:8081",  
   "system_tenant": "system-tenant",  
   "private_key": "./private_key.pem",  
   "cauth_addr": "http://dex-cauth:8080",  
   "token_expiration": 24,  
   "mongo": {  
      "addrs": "127.0.0.1:27017",  
      "db": "cargo_admin",  
      "mode": "strong"  
  }  
}
```
**address**: Endpoint of this cargo-token service
  
**mongo**: Mongo database to be used
  
**system_tenant**: System tenant name, usually it's `system-tenant`
  
**private_key**: Private key used to generate token
  
**cauth_addr**: Address of CAuth service, it is used for AuthZ

**token_expiration**: Expiration time in hours of the token generated
  
## Make  
  
Refer to Makefile for more details about make commands.  
  
To build:  
  
```  
$ make build-linux  
```  
  
To create images:  
  
```  
$ make container VERSION=v0.4.0  
```  
  
Target `container` already includes `build-linux` target, so to build image, this command is sufficient.  Two images will be built: `cargo-admin`, `cargo-token`.
  
## Run Locally  
  
Create a configure file, for example, _~/Tmp/cargo-admin-config.json_. Make sure the mongo and default registry is available. You can start a local mongo service simply by,  
  
```bash
$ docker run -d -p 27017:27017 --name mongo mongo  
```  
  
Run Cargo-Admin,  
  
```bash
$ docker run -it --rm  -v ~/Tmp/cargo-admin-config.json:/app/etc/cargo-admin-config.json -p 8080:8080 cargo.caicloudprivatetest.com/caicloud/cargo-admin:v0.4.0  
```  
  
  Then you can try it out with:
```bash
$ curl -XGET "http://localhost:8080/api/v2/registries/default/projects?includePublic=true"
```

Please note that, since cargo-admin is run in container, you should configure the Mongo address to the host IP, instead of `127.0.0.1`. Also, if you want to try image upload features, please mount `/var/run/dind/docker.sock` correctly, such as `-v /var/run/docker.sock:/var/run/dind/docker.sock`

You can run cargo-token in the similar way, as long as you have configured it correctly.
  
## Token Service  
  
For what is token service and how permission is granted, please refer to document `./docs/token_service.md`