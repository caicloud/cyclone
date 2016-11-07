## Cyclone e2e Test

### HOWTO

run:

```
./run-e2e.sh
```

### Notice

If you are using docker machine, don't forget to forward port `28017`, `9092` and `2379` to host.

Because cyclone is master/slave architecture, so cyclone-worker needs to know the address of cyclone master, you need to override `CYCLONE_SERVER_HOST` to let worker could communicate with master.
