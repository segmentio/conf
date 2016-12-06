# conf [![CircleCI](https://circleci.com/gh/segmentio/conf.svg?style=shield)](https://circleci.com/gh/segmentio/conf) [![Go Report Card](https://goreportcard.com/badge/github.com/segmentio/conf)](https://goreportcard.com/report/github.com/segmentio/conf) [![GoDoc](https://godoc.org/github.com/segmentio/conf?status.svg)](https://godoc.org/github.com/segmentio/conf)
Go package for loading program configuration from multiple sources.

Example
-------

```go
package main

import (
    "fmt"

    "github.com/segmentio/conf"
)

func main() {
    var config struct {
        Message string `json:"m" help:"A message to print."`
    }

    // Load the configuration, either from a config file, the environment or the program arguments.
    conf.Load(&config)

    fmt.Println(config.Message)
}
```
```
$ go run ./example.go -m 'Hello World!'
Hello World!
```
