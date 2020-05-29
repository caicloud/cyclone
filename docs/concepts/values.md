# Values

Cyclone supports 3 different kinds of values, which can be used as stage arguments.

- Concrete value
- Ref value
- Generative value

```yaml
- apiVersion: cyclone.dev/v1alpha1
  kind: Stage
  metadata:
    name: stg
  spec:
    pod:
      inputs:
        arguments:
        - name: v1
          value: master
        - name: v2
          value: $(random:5)
        - name: v3
          value: $(timenow)
        - name: v4
          value: ${variables.key}
```

## Concrete Value

Concrete values are values that can be used directly, only `string` is supported, for example `v1.0`, `Pwd123456`.

## Ref Value

Ref values are reference type values that refer to values in elsewhere, for example `k8s secret`. Valid ref values must be in format `${<types>....}`.

For the moment, there are 3 kinds of reference values supported:

### Secret

```
${secrets.<ns>:<secret>/<jsonpath>/...}
```

Secret value refers to a value in a k8s secret. For a secret named 'secret' under namespace 'ns':
```json
{
  "apiVersion": "v1",
  "data": {
    "key": "KEY",
    "json": "{\"user\":{\"id\": \"111\"}}"
  },
  "kind": "Secret",
  "type": "Opaque",
  "metadata": {
    ...
  }
}
```

`${secrets.ns:secret/data.key}` will resolve to `KEY`
`${secrets.ns:secret/data.json/user.id}` will resolve to `111`

### Global Variables

```
${variables.<key>}
```

Global variable value refers to value defined in workflowrun.

```yaml
apiVersion: cyclone.dev/v1alpha1
kind: WorkflowRun
metadata:
  name: wfr
  namespace: default
spec:
  globalVariables:
  - name: tag
    value: v1.0
  - ...
  ...
```

`${variables.tag}` will resolve to `v1.0`

### Stage Outputs

```
${stages.<stage>.outputs.<key>}
```

Stage output refers to values output by another stage in the same workflow run.

Note: it's not fully implemented yet, but will come soon.

```yaml
apiVersion: cyclone.dev/v1alpha1
kind: WorkflowRun
metadata:
  name: wfr
  namespace: default
spec:
  ...
status:
  stages:
    stg1:
      outputs:
      - name: result
        value: 100
      - name: total
        value: 5
      ...
  ...
```

`${stages.stg1.outputs.total}` will resolves to `5`.

## Generative value

Generative values are values that will be generated in runtime, for example, randomly generated string, timestamp. Generative value type has format `$(<type>:<params>)`.

There are generative value type supported now:

### Random String

```
$(random:<length>)
```

For example, `$(random:5)`.

### Now Time

```
$(timenow:<format>)
```

Timestamp when generate the value, for example, `$(timenow)`, `$(timenow:RFC3339)`