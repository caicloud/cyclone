# http2curl
:triangular_ruler: Convert Golang's http.Request to CURL command line

[![CircleCI](https://circleci.com/gh/moul/http2curl.svg?style=shield)](https://circleci.com/gh/moul/http2curl)
[![GoDoc](https://godoc.org/moul.io/http2curl?status.svg)](https://godoc.org/moul.io/http2curl)
[![License](https://img.shields.io/badge/license-Apache--2.0%20%2F%20MIT-%2397ca00.svg)](https://github.com/moul/http2curl/blob/master/COPYRIGHT)
[![GitHub release](https://img.shields.io/github/release/moul/http2curl.svg)](https://github.com/moul/http2curl/releases)
[![Go Report Card](https://goreportcard.com/badge/moul.io/http2curl)](https://goreportcard.com/report/moul.io/http2curl)
[![CodeFactor](https://www.codefactor.io/repository/github/moul/http2curl/badge)](https://www.codefactor.io/repository/github/moul/http2curl)
[![codecov](https://codecov.io/gh/moul/http2curl/branch/master/graph/badge.svg)](https://codecov.io/gh/moul/http2curl)
[![GolangCI](https://golangci.com/badges/github.com/moul/http2curl.svg)](https://golangci.com/r/github.com/moul/http2curl)
[![Sourcegraph](https://sourcegraph.com/github.com/moul/http2curl/-/badge.svg)](https://sourcegraph.com/github.com/moul/http2curl?badge)
[![Sourcegraph](https://sourcegraph.com/moul.io/http2curl/-/badge.svg)](https://sourcegraph.com/moul.io/http2curl?badge)
[![Made by Manfred Touron](https://img.shields.io/badge/made%20by-Manfred%20Touron-blue.svg?style=flat)](https://manfred.life/)


To do the reverse, check out [mholt/curl-to-go](https://github.com/mholt/curl-to-go).

## Example

```go
import (
    "http"
    "moul.io/http2curl"
)

data := bytes.NewBufferString(`{"hello":"world","answer":42}`)
req, _ := http.NewRequest("PUT", "http://www.example.com/abc/def.ghi?jlk=mno&pqr=stu", data)
req.Header.Set("Content-Type", "application/json")

command, _ := http2curl.GetCurlCommand(req)
fmt.Println(command)
// Output: curl -X PUT -d "{\"hello\":\"world\",\"answer\":42}" -H "Content-Type: application/json" http://www.example.com/abc/def.ghi?jlk=mno&pqr=stu
```

## Install

```php
$ go get moul.io/http2curl
```

## Usages

- https://github.com/parnurzeal/gorequest
- https://github.com/scaleway/scaleway-cli
- https://github.com/nmonterroso/cowsay-slackapp
- https://github.com/moul/as-a-service
- https://github.com/gavv/httpexpect
- https://github.com/smallnest/goreq

## License

© 2019 [Manfred Touron](https://manfred.life)

Licensed under the [Apache License, Version 2.0](https://www.apache.org/licenses/LICENSE-2.0) ([`LICENSE-APACHE`](LICENSE-APACHE)) or the [MIT license](https://opensource.org/licenses/MIT) ([`LICENSE-MIT`](LICENSE-MIT)), at your option. See the [`COPYRIGHT`](COPYRIGHT) file for more details.

`SPDX-License-Identifier: (Apache-2.0 OR MIT)`
