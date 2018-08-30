# Token Service    
 Token service is provided by API `GET /service/token`.  It is used to issue access token to docker registry. Anytime when you access docker registry via `docker pull`, or `docker push`, this API is called.  
  
![token_service.png](./img/token_service.png)  
  
Suppose we have a registry `cargo.caicloudprivatetest.com`, let's to see what would happen when we access it.  
  
### Case #1  

```bash
$ docker pull cargo.caicloudprivatetest.com/caicloud/busybox:latest
```  

Without login, docker registry is accessed anonymously, in this case, only `pull` permission on public project is granted. So the token service will issue an token for this access.  
  
### Case #2

```bash
$ docker push cargo.caicloudprivatetest.com/caicloud/golang:1.9.2  
``` 

Since public project can't be pushed anonymously, the token service won't issue a token for the access, so the above command will fail due to: _denied: requested access to the resource is denied_  

### Case #3  

```bash
$ docker login cargo.caicloudprivatetest.com -u admin -p 123456
$ docker push cargo.caicloudprivatetest.com/caicloud/golang:1.9.2
```
  
After login, in subsequent commands, basic auth information will be send to the token service to get token, if the user is valid and have access to the project, a valid token will be issued.  
    
## Basic Auth    

Token service accepts basic auth. User information is set in request header `Authorization`. For example,    
    
> Authorization: Basic YWRtaW46UHdkMTIzNDU2    

`YWRtaW46UHdkMTIzNDU2` is the base64 encoded user account information.    
    
For more details of basic auth, refer to [RFC 2617](https://tools.ietf.org/html/rfc2617#section-2)    
    
    
## Permission Check  
  
### General Rues  
  
- If project is public, only `pull` is allowed, unless user is registry admin  
- If project doesn't exists, no permission granted  
  
### Anonymous    
 When access registry in anonymous, only pull on public project is allowed. Token service won't issue token for other kinds of access.  
   
### Registry Admin   
If the user provided in the basic auth information is registry admin, token service will grant it full access to the registry. As long as the project exists in the registry, pull or push are all allowed.  
    
### Normal User  
  
For normal user, if it's in single tenant environment, `pull`, `push` to any existing projects are granted.  
  
In multi-tenant environment, permission check is based on roles of users. In CAuth, roles are assigned to teams,  users join teams. Two different role groups are used here, namely `all-projects`, `one-project`. `all-projects` grants permission to all projects (within given tenant), while `one-project` grants permission to one specific project. There are 3 different role types, `guest`, `user`, `owner`. Permissions of them are:

- `guest`: pull
- `user`: pull, push
- `owner`: *

To check permission of one user to one project:

1. Suppose project is in tenant `T`, get a list of related subjects:
	- teams subject: teams in tenant `T` that user is a member of
	- tenant subject: tenant T (This can be regarded as default team of the tenant, that every tenant member is in)
2. For each subject, get all roles assigned to it, with role group `all-projects` or `one-project`. Then check permission based on these roles.  Permission granted by different roles are joined. For example, if one `all-projects-guest` role grants `pull` to the project and one `one-project-user` role grants `pull, push`, then the final permission is `pull, push`.
  
### Cyclone User  
  
Cyclone users are registered by cyclone to execute pipeline. They are special users for token service. Each tenant can register one account in Cargo-Admin with user name `__cyclone__[TENANT]`, so tenant `devops` has user name `__cyclone__devops`.  
  
When cyclone use such account to access registry, token service would:  
  
- In single tenant environment, `pull`, `push` permission is granted.  
- In multi-tenant environment, if project belongs to the tenant specified by the cyclone user, `pull`, `push` permission is granted. Otherwise only `pull` is granted for public project.