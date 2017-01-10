# Function Introduction

## Relating with VCS tools

Cyclone has been integrated with a variety of VCS tools, such as git, svn, etc. After OAuth, it can pull codes from repository and create webhook. Whenever the user commits, submits a pull request or releases a version to the repository, the webhook will trigger the CI/CD workflow. 

- Creating a service relating with the VCS tools

<div align="center">
	<img src="./create_service.png" alt="create_service" width="500">
</div>

- Listing all services

<div align="center">
	<img src="./list_services.png" alt="list_services" width="500">
</div>

## Continus integration and security scanning

All of the processes in workflow are visible. "prebuild" compiles the executable files. "build" builds the published image. "integration" executes the integrated test. "publish" publishes the image and scans the vulnerabilities. "post build" does some relating operations after image published. "deploy" uses the published image to deploy a application to Kubernetes or any other container cloud platform. Cyclone would send email to notify the result of the workflow. All of the processes are shipped by container. It will wipe off the differences caused by environment. 

- Log of the workflow

<div align="center">
	<img src="./logs.png" alt="logs" width="500">
</div>

- Security scanning

<div align="center">
	<img src="./security.png" alt="security" width="500">
</div>

- Sending the build log via email 

<div align="center">
	<img src="./pagging.png" alt="pagging" width="500">
</div>

## Resource management

Cyclone separates the logic of scheduling and the building workflow. It also supports to add user worker nodes and various quota plans. 

- Setting of user resource configuration

<div align="center">
	<img src="./quota.png" alt="quota" width="500">
</div>

- Resource quota of single building

<div align="center">
	<img src="./create_version.png" alt="create_version" width="500">
</div>

## Union publishing and dependency management

Cyclone can manage multi-component united builds. It uses the graphical user interface to display and manage the dependency of the components. 

<div align="center">
	<img src="./dependency.png" alt="dependency" width="500">
</div>

## Continus delivery

Cyclone provides flexible and continuous deployment based on release policy and role control, provides upgrade and rollback policies based on container and image version control.

- various deploying plans

<div align="center">
	<img src="./deployment.png" alt="deployment" width="500">
</div>
