go-slugify
==============

[![Build Status](https://travis-ci.org/mozillazg/go-slugify.svg?branch=master)](https://travis-ci.org/mozillazg/go-slugify)
[![Coverage Status](https://coveralls.io/repos/mozillazg/go-slugify/badge.svg?branch=master)](https://coveralls.io/r/mozillazg/go-slugify?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/mozillazg/go-slugify)](https://goreportcard.com/report/github.com/mozillazg/go-slugify)
[![GoDoc](https://godoc.org/github.com/mozillazg/go-slugify?status.svg)](https://godoc.org/github.com/mozillazg/go-slugify)

Make Pretty Slug.


Installation
------------

```
go get -u github.com/mozillazg/go-slugify
```

Install CLI tool:

```
go get -u github.com/mozillazg/go-slugify/slugify
$ slugify "北京kožušček,abc"
bei-jing-kozuscek-abc
```


Documentation
--------------

API documentation can be found here:
https://godoc.org/github.com/mozillazg/go-slugify


Usage
------

```go
package main

import (
	"fmt"
	"github.com/mozillazg/go-slugify"
)

func main() {
	s := "北京kožušček,abc"
	fmt.Println(slugify.Slugify(s))
	// Output: bei-jing-kozuscek-abc
}
```
