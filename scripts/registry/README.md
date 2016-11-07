# Run a local registry with docker auth

To start, run
```
./start.sh
```

To stop, run
```
./stop.sh
```

The following commends are used to login and push images to local repository:
```
docker login -u admin -p admin_password localhost:5000
docker tag busybox localhost:5000/busybox
docker push localhost:5000/busybox
```

Note: the registry filesystem is using local file system with path specified by REGISTRY_DATA in start.sh
