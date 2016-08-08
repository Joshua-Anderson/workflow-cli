package parser

import (
	"io"

	"github.com/deis/workflow-cli/cmd"
	docopt "github.com/docopt/docopt-go"
)

// Users routes user commands to the specific function.
func Users(argv []string, wOut io.Writer, wErr io.Writer) error {
	usage := `
Valid commands for users:

users:list        list all registered users

Use 'deis help [command]' to learn more.
`

	switch argv[0] {
	case "users:list":
		return usersList(argv, wOut)
	default:
		if printHelp(argv, usage, wOut) {
			return nil
		}

		if argv[0] == "users" {
			argv[0] = "users:list"
			return usersList(argv, wOut)
		}

		PrintUsage(wErr)
		return nil
	}
}

func usersList(argv []string, wOut io.Writer) error {
	usage := addGlobalFlags(`
Lists all registered users.
Requires admin privilages.

Usage: deis users:list [options]

Options:
  -l --limit=<num>
    the maximum number of results to display, defaults to config setting
`)

	args, err := docopt.Parse(usage, argv, true, "", false, true)

	if err != nil {
		return err
	}

	results, err := responseLimit(safeGetValue(args, "--limit"))

	if err != nil {
		return err
	}

	return cmd.UsersList(safeGetValue(args, "--config"), results, wOut)
}
