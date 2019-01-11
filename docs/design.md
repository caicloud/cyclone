# Cyclone Design

Cyclone is a powerful workflow engine and end-to-end pipeline solution implemented with native Kubernetes resources. 
Cyclone is architectured with a low-level workflow engine that is application agnostic, offering capabilities like
workflow DAG scheduling, resource lifecycle management.

## Architecture

![Cyclone arch](images/cyclone-arch.png)

### Components

* Cyclone Web: UI of Cyclone helps users to easily use Cyclone.
* Cyclone Server: API server of Cyclone provides a set of RESTful APIs to operate Cyclone resources.
* Workflow Controller: Controller of Cyclone workflow engine, controls the execution of stages in workflows.

### Kubernetes CRDs

* Stage: Minimum executable unit for workflow, it does some work on input resources and output results.
* Resource: Resource is the data used by stages, it can be the input or output of stages.
* Workflow: Executable DAG graph composed of stages.
* WorkflowRun: Running record of a workflow.
* WorkflowTrigger: Auto-trigger policy for workflow.

## Implementation

### Workflow Engine

Workflow controller is the main component of the workflow engine. It watches various kinds of Cyclone CRDs, and controls the execution of workflows.

Workflow controller creates a pod for each stage to execute, and there are 3 kinds of containers in the pod:

* Init Containers: Init containers will prepare input resources for the stage before workload containers start.
* Workload Containers: Workload containers are functional containers, which do the main work for stages.
* Sidecar Containers: There are 2 common sidecar containers.
  * Coordinator: Coordinator sidecar is in charge of log collection, artifact collection and notifying resource resolver to handle output resources.
  * Resource Resolver: Resource resolver sidecar will handle the output resources after workload containers finished.

### Resources

Each type of resource needs a resource resolver to handle their resources. Now Cyclone supports 4 types of resources:

* Image: Image resources in the registry, supports pulling and pushing operations.
* Git: Code resources in Git SCM like Github and Gitlab, only supports cloning source code.
* KV: Stages can generate some key-value pairs as the outputs, which can be used by dependent stages.
* General: General type allows users to implement handlers by themselves for other types of resources.

### Workflow Executation

Workflow is an executable DAG graph composed of stages, its stages can run serially and parallelly.
If one stage has dependencies, it can only start to run after all its dependencies have finished.
Stages can parallelly run if they have no direct or indirect dependency relationship.
