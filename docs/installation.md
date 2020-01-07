# Installation Guide

> Cyclone has been tested with Kubernetes 1.12, 1.13 and 1.14.

## Install With Helm

### Prerequisites

Cyclone can be easily installed with [Helm](https://helm.sh/) with a version higher than **2.10**,
refer to [helm install guide](https://helm.sh/docs/using_helm/#install-helm) for Helm installation.

You can use `helm template` to generate Kubernetes manifests for Cyclone and then install it using `kubectl apply`,
or you can use `helm install` to manage Cyclone installation by Tiller.
You can [install Tiller](https://helm.sh/docs/using_helm/#initialize-helm-and-install-tiller) after Helm is ready with command:

```bash
$ helm init --history-max 200
```

### Customizable Installation

The simplest way to install Cyclone is using Helm chart. By default, images are pulled from DockerHub, so make sure DockerHub is accessible from your cluster.

```bash
$ helm install --name cyclone --namespace cyclone-system ./helm/cyclone
```

If you want to use your own private registry, you can configure it as:

```bash
$ helm install --name cyclone --namespace cyclone-system --set imageRegistry.registry=cargo.caicloud.xyz,imageRegistry.project=release ./helm/cyclone
```

For more detailed configuration, please use values file, [default values file](../helm/cyclone/values.yaml) is a good reference on how to write it.

```bash
$ helm install --name cyclone --namespace cyclone-system -f <path-to-your-values-file> ./helm/cyclone
```

If you want to release after change charts, you can upgrade with command:

```bash
$ helm upgrade cyclone ./helm/cyclone
```

If you want to uninstall Cyclone, you can clean up it with command:

```bash
$ helm delete --purge cyclone
```

### Configuration

#### Common Configurations

| Parameter | Description | Default |
| --------- | --------- | --------- |
| `imageRegistry.registry` | Image registry where to pull images | `docker.io` |
| `imageRegistry.project` | Project in the registry where to pull Cyclone component images | `caicloud` |
| `imageRegistry.libraryProject` | Project in the registry where to pull common images, like `busybox` | `library` |
| `serverAddress` | Address of the Cyclone Server (provides Restful APIs) | `cyclone-server.default.svc.cluster.local::7099` |
| `systemNamespace` | Namespace where Cyclone will be installed | `default` |

#### Workflow Engine Configurations

| Parameter | Description | Default |
| --------- | --------- | --------- |
| `engine.images` | Images used in workflow engine, for example, `engine.images.gc` (default `alpine:3.8`) defines image used for garbage collection. | `alpine:3.8` for `engine.images.gc`, `docker:18.03-dind` for `dind`|
| `engine.gc.enabled` | Whether enable garbage collection | `true` |
| `engine.gc.delaySeconds` | Time to wait before cleaning up execution of a workflow. It gives users chance to check the execution details (for example, pods, data on PVC) when execution finished | `300` |
| `engine.gc.retry` | How many times to retry (include the initial one, so `1` means no retry) when performing GC | `1` |
| `engine.limits.maxWorkflowRuns` | Maximum number of execution records to keep for each workflow | `50` |
| `engine.resourceRequirement` | Default resource requirements that would be applied to each stage, if non specified when execute a workflow | CPU: 50m/100m, Memory: 128Mi/256Mi |
| `engine.notification.url` | URL where to notify workflow execution result, it's Cyclone server by default | `http://cyclone-server.default.svc.cluster.local::7099/apis/v1alpha1/notifications` |
| `engine.developMode` | Whether it's in develop mode, in develop mode, ImagePullPolicy would be `Always` in the engine | `true` |
| `engine.executionContext.namespace` | If no execution context（cluster, namespace, PVC） specified when run workflow, use this namespace in control cluster | `cyclone-system` |
| `engine.executionContext.pvc` | If no execution context specified when run workflow, use this PVC | `cyclone-pvc-system` |

#### Cyclone Server Configurations

| Parameter | Description | Default |
| --------- | --------- | --------- |
| `server.listenAddress` | Address where Cyclone server will serve | `0.0.0.0` |
| `server.listenPort` | Port that Cyclone server will serve on | `7099` |
| `server.nodePort` | Node port that Cyclone server will expose its service from the cluster | `30011` |
| `server.init.defaultTenant` | Whether init default tenant `system` in the installation | `true` |
| `server.init.templates` | Whether init stage templates | `true` |
| `server.openControlCluster` | When a tenant created, whether to open control cluster for the tenant to run workflow, to open control cluster, PVC would be created | `true` |
| `server.pvc.storageClass` | Default StorageClass used to create PVC | Empty string |
| `server.pvc.size` | Default PVC size if not specified | `10Gi` |
| `server.resourceRequirement` | Resource quota applied to namespace create for tenant to run workflow | CPU: 1/2, Memory: 2Gi/4Gi |
| `server.storageWatcher.reportUrl` | URL to report PVC usage, it's Cyclone server by default | `http://cyclone-server.default.svc.cluster.local::7099/apis/v1alpha1/storage/usages` |
| `server.storageWatcher.intervalSeconds` | Time interval to report PVC usage | `30` |
| `server.storageWatcher.resourceRequirements` | Resource requirements applied to the storage watcher pod | CPU: 50m/100m, Memory: 32Mi/64Mi |

#### Cyclone Web Configurations 

| Parameter | Description | Default |
| --------- | --------- | --------- |
| `web.replicas` | Replicas of cyclone web | `1` |
| `web.listenPort` | Port Cyclone web will listen on | `80` |
| `web.nodePort` | Node port that Cyclone web will expose its service | `30022` |

## Build Images

If you'd like to build images from the source code, please do it as:

```bash
$ docker login <registry> -u <user> -p <pwd>
$ make push REGISTRIES=<registry>/<project>
```

Here <registry>/<project> specifies the registry and project where to push your images, for example, `cargo.caicloud.xyz/release`.

Then install Cyclone with Helm.

```bash
$ helm install --name cyclone --namespace cyclone-system --set imageRegistry.registry=<registry>,imageRegistry.project=<project> ./helm/cyclone
```
