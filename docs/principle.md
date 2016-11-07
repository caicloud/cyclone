## Workflow
![flow](flow.png)
- Circle provides abundant [API](http://118.193.142.27:7099/apidocs/) for web applications.
- After relating the code repository in VCS with the service of Circle via API, commiting and releasing to VCS will notify Circle-Server by webhook
- Circle-Server will run a Circle-Worker container which base the tech of Docker-in-Docker, the container will checkout code from VCS, then execute steps according to configrations of caicloud.yml in the code repository as followed: 
  - PreBuild: compile executable file in the specified system environment
  - Build: copy the executable file to the specified system environment, packet the environment to a image and push the image to registry
  - PostBuild：run a container to execute some shells or commads which aims to do some related operations after the images published
  - Integretion: use the image built durning the build step to run a container , run the micro service containers which depend by CI, run the integretion testing
  - Deploy: use the published image to run a application in to containers cluster Platform such as kubernetes
- The process log can be pulled from Circle-Server via websocket
- Circle-Server will send the result and log of CI & CD to user by email when the progress has finished

## Architecture
![architecture](architecture.png)

Each cube represent a container
- The Api-Server component in Circle-Server provides the restful API service, if the task created by calling the API needs long time to handle, it will generate a pending event and write into etcd
- The EventManger component load pending events from etcd, watch the change of events, and send new pending event to WorkerManager
- WorkerManager call the docker API to run a Circle-Worker container, and send infomations to it via ENVs
- Circle-Worker use event ID as a token to call API and get event infomations, then run containers to execute integretion, prebuild, build and post build steps, progress log push to Log-Server and save to kafka
- Log-Server component pull log from kafka and push it to user
- The date need to be persisted save into mongo