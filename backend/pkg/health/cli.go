// Copyright 2026 Northern.tech AS
//
// Licensed under the Apache License, Version 2.0 (see LICENSE).

package health

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli"
)

// Command returns a urfave/cli.Command named "healthcheck". urlFn is
// evaluated when the subcommand runs so the caller can read the service's
// resolved config (listen address) rather than baking it in at registration
// time. A failure prints the error to stderr and exits with status 1 so
// that Docker records the probe as failed.
func Command(urlFn func(c *cli.Context) string) cli.Command {
	return cli.Command{
		Name:   "healthcheck",
		Usage:  "Probe the service's internal health endpoint and exit 0 if healthy.",
		Action: makeAction(urlFn, os.Exit),
	}
}

// makeAction is separated so tests can inject an exit function.
func makeAction(urlFn func(c *cli.Context) string, exit func(int)) cli.ActionFunc {
	return func(c *cli.Context) error {
		url := urlFn(c)
		if err := Probe(context.Background(), url, DefaultTimeout); err != nil {
			fmt.Fprintln(os.Stderr, err)
			exit(1)
			return nil
		}
		return nil
	}
}
