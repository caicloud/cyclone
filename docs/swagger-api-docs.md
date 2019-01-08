# Cyclone Server APIs

`Cyclone Server` provides a set of APIs to interact with Cyclone, you can access the API docs from `Cyclone Web` by accessing:

```
http://<cyclone-web>:<port>/swagger
```

For example, `http://localhost:3000/swagger` when run `Cyclone Web` locally.

## Generate API Specification

`Cyclone Server` is built with [Nirvana](https://github.com/caicloud/nirvana) framework, we can generate API specifications from source code conveniently by following steps:

1. Install nirvana cli and its dependencies:

```
$ go get -u github.com/caicloud/nirvana/cmd/nirvana
$ go get -u github.com/golang/dep/cmd/dep
```

2. Run following command Cyclone root directory:

```
nirvana api --output ./ pkg/server/apis --serve=":8088"
```

This will generate the API specification file and render it with [ReDoc](https://github.com/Rebilly/ReDoc).

You can access the APIs with nice UI interface through `http://localhost:8088`.

And the API specification file is located at `./api.v1alpha1.json`
