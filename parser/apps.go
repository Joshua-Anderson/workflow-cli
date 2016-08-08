package parser

import (
	"io"
	"strconv"
	"strings"

	"github.com/deis/workflow-cli/cmd"
	docopt "github.com/docopt/docopt-go"
)

// Apps routes app commands to their specific function.
func Apps(argv []string, wOut io.Writer, wErr io.Writer) error {
	usage := `
Valid commands for apps:

apps:create        create a new application
apps:list          list accessible applications
apps:info          view info about an application
apps:open          open the application in a browser
apps:logs          view aggregated application logs
apps:run           run a command in an ephemeral app container
apps:destroy       destroy an application
apps:transfer      transfer app ownership to another user

Use 'deis help [command]' to learn more.
`

	switch argv[0] {
	case "apps:create":
		return appCreate(argv, wOut)
	case "apps:list":
		return appsList(argv, wOut)
	case "apps:info":
		return appInfo(argv, wOut)
	case "apps:open":
		return appOpen(argv, wOut)
	case "apps:logs":
		return appLogs(argv, wOut)
	case "apps:run":
		return appRun(argv, wOut)
	case "apps:destroy":
		return appDestroy(argv, wOut)
	case "apps:transfer":
		return appTransfer(argv, wOut)
	default:
		if printHelp(argv, usage, wOut) {
			return nil
		}

		if argv[0] == "apps" {
			argv[0] = "apps:list"
			return appsList(argv, wOut)
		}

		PrintUsage(wErr)
		return nil
	}
}

func appCreate(argv []string, wOut io.Writer) error {
	usage := addGlobalFlags(`
Creates a new application.

- if no <id> is provided, one will be generated automatically.

Usage: deis apps:create [<id>] [options]

Arguments:
  <id>
    a uniquely identifiable name for the application. No other app can already
    exist with this name.

Options:
  --no-remote
    do not create a 'deis' git remote.
  -b --buildpack BUILDPACK
    a buildpack url to use for this app
  -r --remote REMOTE
    name of remote to create. [default: deis]
`)

	args, err := docopt.Parse(usage, argv, true, "", false, true)

	if err != nil {
		return err
	}

	id := safeGetValue(args, "<id>")
	buildpack := safeGetValue(args, "--buildpack")
	remote := safeGetValue(args, "--remote")
	cf := safeGetValue(args, "--config")
	noRemote := args["--no-remote"].(bool)

	return cmd.AppCreate(cf, id, buildpack, remote, noRemote, wOut)
}

func appsList(argv []string, wOut io.Writer) error {
	usage := addGlobalFlags(`
Lists applications visible to the current user.

Usage: deis apps:list [options]

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

	cf := safeGetValue(args, "--config")

	return cmd.AppsList(cf, results, wOut)
}

func appInfo(argv []string, wOut io.Writer) error {
	usage := addGlobalFlags(`
Prints info about the current application.

Usage: deis apps:info [options]

Options:
  -a --app=<app>
    the uniquely identifiable name for the application.
`)

	args, err := docopt.Parse(usage, argv, true, "", false, true)

	if err != nil {
		return err
	}

	app := safeGetValue(args, "--app")
	cf := safeGetValue(args, "--config")

	return cmd.AppInfo(cf, app, wOut)
}

func appOpen(argv []string, wOut io.Writer) error {
	usage := addGlobalFlags(`
Opens a URL to the application in the default browser.

Usage: deis apps:open [options]

Options:
  -a --app=<app>
    the uniquely identifiable name for the application.
`)

	args, err := docopt.Parse(usage, argv, true, "", false, true)

	if err != nil {
		return err
	}

	app := safeGetValue(args, "--app")
	cf := safeGetValue(args, "--config")

	return cmd.AppOpen(cf, app, wOut)
}

func appLogs(argv []string, wOut io.Writer) error {
	usage := addGlobalFlags(`
Retrieves the most recent log events.

Usage: deis apps:logs [options]

Options:
  -a --app=<app>
    the uniquely identifiable name for the application.
  -n --lines=<lines>
    the number of lines to display
`)

	args, err := docopt.Parse(usage, argv, true, "", false, true)

	if err != nil {
		return err
	}

	app := safeGetValue(args, "--app")
	cf := safeGetValue(args, "--config")

	linesStr := safeGetValue(args, "--lines")
	var lines int

	if linesStr == "" {
		lines = -1
	} else {
		lines, err = strconv.Atoi(linesStr)

		if err != nil {
			return err
		}
	}

	return cmd.AppLogs(cf, app, lines, wOut)
}

func appRun(argv []string, wOut io.Writer) error {
	usage := addGlobalFlags(`
Runs a command inside an ephemeral app container. Default environment is
/bin/bash.

Usage: deis apps:run [options] [--] <command>...

Arguments:
  <command>
    the shell command to run inside the container.

Options:
  -a --app=<app>
    the uniquely identifiable name for the application.
`)

	args, err := docopt.Parse(usage, argv, true, "", false, true)

	if err != nil {
		return err
	}

	app := safeGetValue(args, "--app")
	cf := safeGetValue(args, "--config")
	command := strings.Join(args["<command>"].([]string), " ")

	return cmd.AppRun(cf, app, command, wOut)
}

func appDestroy(argv []string, wOut io.Writer) error {
	usage := addGlobalFlags(`
Destroys an application.

Usage: deis apps:destroy [options]

Options:
  -a --app=<app>
    the uniquely identifiable name for the application.
  --confirm=<app>
    skips the prompt for the application name. <app> is the uniquely identifiable
    name for the application.
`)

	args, err := docopt.Parse(usage, argv, true, "", false, true)

	if err != nil {
		return err
	}

	app := safeGetValue(args, "--app")
	confirm := safeGetValue(args, "--confirm")
	cf := safeGetValue(args, "--config")

	return cmd.AppDestroy(cf, app, confirm, wOut)
}

func appTransfer(argv []string, wOut io.Writer) error {
	usage := addGlobalFlags(`
Transfer app ownership to another user.

Usage: deis apps:transfer <username> [options]

Arguments:
  <username>
    the user that the app will be transfered to.

Options:
  -a --app=<app>
    the uniquely identifiable name for the application.
`)

	args, err := docopt.Parse(usage, argv, true, "", false, true)

	if err != nil {
		return err
	}

	cf := safeGetValue(args, "--config")
	app := safeGetValue(args, "--app")
	user := safeGetValue(args, "<username>")

	return cmd.AppTransfer(cf, app, user, wOut)
}
