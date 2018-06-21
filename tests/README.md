## Cyclone e2e Test

### HOWTO

- if you are using docker machine, run:

```
./scripts/e2e-docker.sh
```

- if you have a k8s cluster, run:
```
./scripts/e2e-k8s.sh
```

### Notice

If you are using docker machine, don't forget to forward port `27017` to host.

Because cyclone is master/slave architecture, so cyclone-worker needs to know the address of cyclone master, you need to override `CYCLONE_SERVER` to let worker could communicate with master.
