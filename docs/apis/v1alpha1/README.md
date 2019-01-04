# Use Nirvana to generate cyclone-server API docs

Cyclone-server used [Nirvana](https://github.com/caicloud/nirvana) framework, so we can generate API docs as follow steps:
1. Install nirvana cli and its dependences:
```
$ go get -u github.com/caicloud/nirvana/cmd/nirvana
$ go get -u github.com/golang/dep/cmd/dep
```

2. Exec command in cyclone root dir:
```
nirvana api pkg/server/apis --serve=":8081"
```

3. You can access Cyclone-server API in your browser by link:
```
http://localhost:8081
```