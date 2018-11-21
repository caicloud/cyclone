# Tracing Application example

## Running
### Run Jaeger Backend
An all-in-one Jaeger backend is packaged as a Docker container with in-memory storage.
```
docker run -d -p6831:6831/udp -p16686:16686 jaegertracing/all-in-one:latest
```
Jaeger UI can be accessed at http://localhost:16686.

### Run Tracing Application
```
go run ./main.go

curl -H "Content-Type: application/json" -X POST -d '{"namespace":"ns1","name":"mongodb", "config": "db"}' http://localhost:8080/applications
```
