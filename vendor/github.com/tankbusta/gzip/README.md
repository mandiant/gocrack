# GZIP gin's middleware

[![Build Status](https://travis-ci.org/gin-contrib/gzip.svg)](https://travis-ci.org/gin-contrib/gzip)
[![codecov](https://codecov.io/gh/gin-contrib/gzip/branch/master/graph/badge.svg)](https://codecov.io/gh/gin-contrib/gzip)
[![Go Report Card](https://goreportcard.com/badge/github.com/gin-contrib/gzip)](https://goreportcard.com/report/github.com/gin-contrib/gzip)
[![GoDoc](https://godoc.org/github.com/gin-contrib/gzip?status.svg)](https://godoc.org/github.com/gin-contrib/gzip)
[![Join the chat at https://gitter.im/gin-gonic/gin](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/gin-gonic/gin)

Gin middleware to enable `GZIP` support.

## Usage

### Start using it

Download and install it:

```sh
$ go get github.com/gin-contrib/gzip
```

Import it in your code:

```go
import "github.com/gin-contrib/gzip"
```

### Canonical example:

```go
package main

import (
	"fmt"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong "+fmt.Sprint(time.Now().Unix()))
	})

	// Listen and Server in 0.0.0.0:8080
	r.Run(":8080")
}

```
