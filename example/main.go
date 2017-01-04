// This example program can be used to test the features provided by the
// github.com/segmentio/conf package.
//
// Passing configuration via the program arguments:
//
//	$ go run ./example/main.go -msg 'hello! (from the arguments)'
//	[main] hello! (from the arguments)
//
// Passing configuration via the environment variables:
//
//	$ MAIN_MSG='hello! (from the environment)' go run ./example/main.go
//	[main] hello! (from the environment)
//
// Passing configuration via a configuration file:
//
//	$ go run ./example/main.go -config-file ./example/config.yml
//	[main] hello ${USER}! (from the config file)
//
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/segmentio/conf"
)

func main() {
	config := struct {
		Message string `conf:"msg" help:"The message to print out."`
	}{
		Message: "default",
	}
	conf.Load(&config)
	fmt.Printf("[%s] %s\n", filepath.Base(os.Args[0]), config.Message)
}
