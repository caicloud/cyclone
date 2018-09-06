# Metrics

It's time to expose some metrics to help understand and diagnose our service! Nirvana has out-of-box support for
instrumentation. To enable exposing request metrics, just add one more configuration:

```go
config := nirvana.NewDefaultConfig("", 8080).
	Configure(
		metrics.Path("/metrics"),
	)
```

The actual configuration is done with `metrics` plugin. `plugin` is another concept in Nirvana - we can always
add more functionalities to Nirvana via plugin, and each plugin can be individually enabled or disabled. How
plugins are implemented depends on plugin author. For example, some plugins are simply static configuration,
while some are more complex middlewares. All plugins are registered into config. The server will install them
when the server starts.

Now if we start our server, we'll see a wealth of information exposed as [prometheus](https://prometheus.io) format.
The metrics are exposed via `/metrics` endpoint.

```
$ go run ./examples/getting-started/metrics/echo.go
```

Use ab (ApacheBench) to simulate some user load; in another terminal, run:

```
ab -n 1000 -H 'Content-type: application/json' 'http://localhost:8080/echo?msg=testtesttest'
```

Once done, let's checkout some default metrics from metrics plugin. The metric `nirvana_request_count` tells
us how many requests we've seen in total. Since we use `-n 1000`, there will be 1000 requests to `/echo` endpoint.

```
$ curl http://localhost:8080/metrics 2>&1 | grep nirvana_request_count
# HELP nirvana_request_count Counter of server requests broken out for each verb, API resource, client, and HTTP response contentType and code.
# TYPE nirvana_request_count counter
nirvana_request_count{client="ApacheBench/2.3",code="200",contentType="application/json",method="GET",path="/echo"} 1000
```

The metric `nirvana_request_latencies` shows distribution of our service latencies. We've added a random sleep
between [0, 100) in our service; therefore, p90 is around 90m.

```
$ curl http://localhost:8080/metrics 2>&1 | grep "nirvana_request_latencies"
# HELP nirvana_request_latencies Response latency distribution in microseconds for each verb, resource and client.
# TYPE nirvana_request_latencies histogram
nirvana_request_latencies_bucket{method="GET",path="/echo",le="125000"} 1000
nirvana_request_latencies_bucket{method="GET",path="/echo",le="250000"} 1000
nirvana_request_latencies_bucket{method="GET",path="/echo",le="500000"} 1000
nirvana_request_latencies_bucket{method="GET",path="/echo",le="1e+06"} 1000
nirvana_request_latencies_bucket{method="GET",path="/echo",le="2e+06"} 1000
nirvana_request_latencies_bucket{method="GET",path="/echo",le="4e+06"} 1000
nirvana_request_latencies_bucket{method="GET",path="/echo",le="8e+06"} 1000
nirvana_request_latencies_bucket{method="GET",path="/echo",le="+Inf"} 1000
nirvana_request_latencies_sum{method="GET",path="/echo"} 48533
nirvana_request_latencies_count{method="GET",path="/echo"} 1000
nirvana_request_latencies_bucket{method="GET",path="/metrics",le="125000"} 1
nirvana_request_latencies_bucket{method="GET",path="/metrics",le="250000"} 1
nirvana_request_latencies_bucket{method="GET",path="/metrics",le="500000"} 1
nirvana_request_latencies_bucket{method="GET",path="/metrics",le="1e+06"} 1
nirvana_request_latencies_bucket{method="GET",path="/metrics",le="2e+06"} 1
nirvana_request_latencies_bucket{method="GET",path="/metrics",le="4e+06"} 1
nirvana_request_latencies_bucket{method="GET",path="/metrics",le="8e+06"} 1
nirvana_request_latencies_bucket{method="GET",path="/metrics",le="+Inf"} 1
nirvana_request_latencies_sum{method="GET",path="/metrics"} 0
nirvana_request_latencies_count{method="GET",path="/metrics"} 1
# HELP nirvana_request_latencies_summary Response latency summary in microseconds for each verb and resource.
# TYPE nirvana_request_latencies_summary summary
nirvana_request_latencies_summary{method="GET",path="/echo",quantile="0.5"} 53
nirvana_request_latencies_summary{method="GET",path="/echo",quantile="0.9"} 89
nirvana_request_latencies_summary{method="GET",path="/echo",quantile="0.99"} 98
nirvana_request_latencies_summary_sum{method="GET",path="/echo"} 48533
nirvana_request_latencies_summary_count{method="GET",path="/echo"} 1000
nirvana_request_latencies_summary{method="GET",path="/metrics",quantile="0.5"} 0
nirvana_request_latencies_summary{method="GET",path="/metrics",quantile="0.9"} 0
nirvana_request_latencies_summary{method="GET",path="/metrics",quantile="0.99"} 0
nirvana_request_latencies_summary_sum{method="GET",path="/metrics"} 0
nirvana_request_latencies_summary_count{method="GET",path="/metrics"} 1
```

See user guide for more information about metrics plugin (and others). For full example code, see [metrics](./examples/getting-started/metrics).
