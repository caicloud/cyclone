package main

import (
	"fmt"
	"os"
	"sort"

	"gopkg.in/urfave/cli.v1"
)

// newCliApp create a new server cli app
func newCliApp() *cli.App {

	app := cli.NewApp()

	app.Name = "cyclone-worker"

	opts := NewWorkerOptions()
	opts.AddFlags(app)

	app.Action = func(c *cli.Context) error {
		worker := opts.NewWorker()
		return worker.Run()
	}

	// sort flags by name
	sort.Sort(cli.FlagsByName(app.Flags))

	return app
}

func main() {
	app := newCliApp()
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
