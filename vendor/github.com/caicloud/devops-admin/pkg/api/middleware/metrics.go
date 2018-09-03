/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/emicklei/go-restful"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	requestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_request_count",
			Help: "Counter of api requests broken out for each version, verb, API resource, client, and HTTP response contentType and code.",
		},
		[]string{"version", "verb", "resource", "client", "contentType", "code"},
	)
	requestLatencies = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "api_request_latencies",
			Help: "Response latency distribution in microseconds for each version, verb, resource and client.",
			// Use buckets ranging from 125 ms to 8 seconds.
			Buckets: prometheus.ExponentialBuckets(125000, 2.0, 7),
		},
		[]string{"version", "verb", "resource"},
	)
	requestLatenciesSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "api_request_latencies_summary",
			Help: "Response latency summary in microseconds for each verb and resource.",
			// Make the sliding window of 1h.
			MaxAge: time.Hour,
		},
		[]string{"version", "verb", "resource"},
	)
)

func init() {
	prometheus.MustRegister(requestCounter)
	prometheus.MustRegister(requestLatencies)
	prometheus.MustRegister(requestLatenciesSummary)
}

// Extract the resource from the path
// Third part of URL (/api/v1/<resource>) and ignore potential subresources
func mapUrlToResource(url string) *string {
	parts := strings.Split(url, "/")
	if len(parts) < 3 {
		return &url
	}
	return &parts[3]
}

func addMetrics(version, verb, resource, client, content, code string, startTime time.Time) {
	// Convert to microseconds
	elapsed := time.Since(startTime).Seconds() * 1000

	requestCounter.WithLabelValues(version, verb, resource, client, content, code).Inc()
	requestLatencies.WithLabelValues(version, verb, resource).Observe(elapsed)
	requestLatenciesSummary.WithLabelValues(version, verb, resource).Observe(elapsed)
}

func APIMetrics(version string) restful.FilterFunction {
	return func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		startTime := time.Now()
		verb := req.Request.Method
		resource := mapUrlToResource(req.SelectedRoutePath())
		client := getHTTPClient(req.Request)
		chain.ProcessFilter(req, resp)
		code := strconv.Itoa(resp.StatusCode())
		contentType := resp.Header().Get("Content-Type")
		if resource != nil {
			addMetrics(version, verb, *resource, client, contentType, code, startTime)
		}
	}
}

func getHTTPClient(req *http.Request) string {
	if userAgent, ok := req.Header["User-Agent"]; ok {
		if len(userAgent) > 0 {
			return userAgent[0]
		}
	}
	return "unknown"
}
