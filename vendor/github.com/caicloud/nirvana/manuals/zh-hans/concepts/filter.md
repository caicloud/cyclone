# 过滤器

包路径: `github.com/caicloud/nirvana/service`

在某些场景下，我们需要从源头对请求进行处理和过滤。Nirvana 提供了 Filter 机制，可以在收到一个请求的时候，立刻进行处理。并根据 Filter 的返回值来确定是直接丢弃该请求还是继续处理。

```go
// Filter can filter request. It has the highest priority in a request
// lifecycle. It runs before router matching.
// If a filter return false, that means the request should be filtered.
// If a filter want to filter a request, it should handle the request
// by itself.
type Filter func(resp http.ResponseWriter, req *http.Request) bool
```

Filter 在整个 Nirvana 框架中处于最高优先级。Filter 返回 false 则表示请求不应该被继续处理，立刻丢弃。

## Nirvana 提供的一些 Filters

### RedirectTrailingSlash

这个过滤器判断 URL Path 尾部是不是存在 `/`，如果存在就重定向到没有 `/` 的路径上。

### FillLeadingSlash

这个过滤器判断 URL Path 首部有没有 `/`，如果没有就加上 `/`。

### ParseRequestForm

这个过滤器只针对 `application/x-www-form-urlencoded` 和 `multipart/form-data`，然后 Parse 这两种类型的请求体，并转换为 Form 和 File。

