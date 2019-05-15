# Stage Execution Result

In addition to stage output resources (data pushed to external systems, like docker image), a stage can also have execution result. Execution results are key-value data that would be saved in WorkflowRun status.

For example,

```yaml
apiVersion: cyclone.dev/v1alpha1
kind: WorkflowRun
metadata:
  ...
spec:
  ...
status:
  ...
  stages:
    sonarcube-stage:
      outputs:
      - key: overall
        value: Passed
```

## How It Works

Execution result is collected from result file `/__result__` in workload container. So it you want some results to be saved to WorkflowRun status, you should write them to the result file `/__result__`.

The result file is a plain text file with line format `<key>:<value>`. Here is a simple example stage with execution results generated:

```yaml
apiVersion: cyclone.dev/v1alpha1
kind: Stage
metadata:
  name: example
spec:
  pod:
    spec:
      containers:
      - name: main
        image: busybox:1.30.0
        command:
        - /bin/sh
        - -c
        - echo "overall:Passed" >> /__result__
```


