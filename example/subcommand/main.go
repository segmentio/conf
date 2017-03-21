package main

import (
	"fmt"
	"os"

	"github.com/segmentio/conf"
)

var config struct {
	Baz string `conf:"baz" help:"some flag"`
}

func main() {
	cmd, args := conf.LoadWith(&config, conf.Loader{
		Name: "root",
		Args: os.Args[1:],
		Commands: []conf.Command{
			{"cmd", "child command"},
		},
	})

	if cmd == "cmd" {
		var cmdconf struct {
			Foo string `conf:"foo" help:"some flag"`
		}

		conf.LoadWith(&cmdconf, conf.Loader{
			Name: "cmd",
			Args: args,
		})

		fmt.Printf("hello from `cmd`, root_conf: %+v cmd_conf: %+v\n", config, cmdconf)
		return
	}
}
