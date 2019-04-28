# Installation Guide

## From Source Code

### Build and push images

Here <registry>/<project> specifies the registry and project where to push your images, for example, test.caicloudprivatetest.com/release.

```bash
$ docker login <registry> -u <user> -p <pwd>
$ make push REGISTRIES=<registry>/<project>
```

### Deploy to Kubernetes cluster
Please make sure kubectl is installed and appropriately configured. Here AUTH is the credential to the docker registry. PVC is a PersistentVolumeClaim that needs to be prepared in k8s.

```bash
$ make deploy REGISTRIES=<registry>/<project> AUTH=<user>:<pwd> PVC=<pvc>
```