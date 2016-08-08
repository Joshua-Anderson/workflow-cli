package parser

import (
	"io"

	"github.com/deis/workflow-cli/cmd"
	docopt "github.com/docopt/docopt-go"
)

// Routing displays all relevant commands for `deis routing`.
func Routing(argv []string, wOut io.Writer, wErr io.Writer) error {
	usage := `
Valid commands for routing:

routing:info       view routability of an application
routing:enable     enable routing for an app
routing:disable    disable routing for an app

Use 'deis help [command]' to learn more.
`

	switch argv[0] {
	case "routing:info":
		return routingInfo(argv, wOut)
	case "routing:enable":
		return routingEnable(argv, wOut)
	case "routing:disable":
		return routingDisable(argv, wOut)
	default:
		if printHelp(argv, usage, wOut) {
			return nil
		}

		if argv[0] == "routing" {
			argv[0] = "routing:info"
			return routingInfo(argv, wOut)
		}

		PrintUsage(wErr)
		return nil
	}
}

func routingInfo(argv []string, wOut io.Writer) error {
	usage := addGlobalFlags(`
Prints info about the current application's routability.

Usage: deis routing:info [options]

Options:
  -a --app=<app>
    the uniquely identifiable name for the application.
`)

	args, err := docopt.Parse(usage, argv, true, "", false, true)

	if err != nil {
		return err
	}

	return cmd.RoutingInfo(safeGetValue(args, "--config"), safeGetValue(args, "--app"), wOut)
}

func routingEnable(argv []string, wOut io.Writer) error {
	usage := addGlobalFlags(`
Enables routability for an app.

Usage: deis routing:enable [options]

Options:
  -a --app=<app>
    the uniquely identifiable name of the application.
`)

	args, err := docopt.Parse(usage, argv, true, "", false, true)

	if err != nil {
		return err
	}

	return cmd.RoutingEnable(safeGetValue(args, "--config"), safeGetValue(args, "--app"), wOut)
}

func routingDisable(argv []string, wOut io.Writer) error {
	usage := addGlobalFlags(`
Disables routability for an app.

Usage: deis routing:disable [options]

Options:
  -a --app=<app>
    the uniquely identifiable name of the application.
`)

	args, err := docopt.Parse(usage, argv, true, "", false, true)

	if err != nil {
		return err
	}

	return cmd.RoutingDisable(safeGetValue(args, "--config"), safeGetValue(args, "--app"), wOut)
}
