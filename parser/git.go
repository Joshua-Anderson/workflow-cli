package parser

import (
	"fmt"
	"io"

	"github.com/deis/workflow-cli/cmd"
	docopt "github.com/docopt/docopt-go"
)

// Git routes git commands to their specific function.
func Git(argv []string, wOut io.Writer, wErr io.Writer) error {
	usage := `
Valid commands for git:

git:remote          Adds git remote of application to repository
git:remove          Removes git remote of application from repository

Use 'deis help [command]' to learn more.
`

	switch argv[0] {
	case "git:remote":
		return gitRemote(argv, wOut)
	case "git:remove":
		return gitRemove(argv, wOut)
	case "git":
		fmt.Fprint(wOut, usage)
		return nil
	default:
		PrintUsage(wErr)
		return nil
	}
}

func gitRemote(argv []string, wOut io.Writer) error {
	usage := addGlobalFlags(`
Adds git remote of application to repository

Usage: deis git:remote [options]

Options:
  -a --app=<app>
    the uniquely identifiable name for the application.
  -r --remote=REMOTE
    name of remote to create. [default: deis]
  -f --force
    overwrite remote of the given name if it already exists.
`)

	args, err := docopt.Parse(usage, argv, true, "", false, true)

	if err != nil {
		return err
	}

	cf := safeGetValue(args, "--config")
	app := safeGetValue(args, "--app")
	remote := safeGetValue(args, "--remote")
	force := args["--force"].(bool)

	return cmd.GitRemote(cf, app, remote, force, wOut)
}

func gitRemove(argv []string, wOut io.Writer) error {
	usage := addGlobalFlags(`
Removes git remotes of application from repository.

Usage: deis git:remove [options]

Options:
  -a --app=<app>
    the uniquely identifiable name for the application.
`)

	args, err := docopt.Parse(usage, argv, true, "", false, true)

	if err != nil {
		return err
	}

	return cmd.GitRemove(safeGetValue(args, "--config"), safeGetValue(args, "--app"), wOut)
}
