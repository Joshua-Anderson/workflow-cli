package parser

import (
	"github.com/deis/workflow-cli/cmd"
	docopt "github.com/docopt/docopt-go"
)

// Config routes config commands to their specific function.
func Config(argv []string) error {
	usage := `
Valid commands for config:

config:list        list environment variables for an app
config:set         set environment variables for an app
config:unset       unset environment variables for an app
config:pull        extract environment variables to .env
config:push        set environment variables from .env

Use 'deis help [command]' to learn more.
`

	switch argv[0] {
	case "config:list":
		return configList(argv)
	case "config:set":
		return configSet(argv)
	case "config:unset":
		return configUnset(argv)
	case "config:pull":
		return configPull(argv)
	case "config:push":
		return configPush(argv)
	default:
		if printHelp(argv, usage) {
			return nil
		}

		if argv[0] == "config" {
			argv[0] = "config:list"
			return configList(argv)
		}

		PrintUsage()
		return nil
	}
}

func configList(argv []string) error {
	usage := addGlobalFlags(`
Lists environment variables for an application.

Usage: deis config:list [options]

Options:
  --oneline
    print output on one line.
  -a --app=<app>
    the uniquely identifiable name of the application.
`)

	args, err := docopt.Parse(usage, argv, true, "", false, true)

	if err != nil {
		return err
	}

	cf := safeGetValue(args, "--config")
	app := safeGetValue(args, "--app")
	oneline := args["--oneline"].(bool)

	return cmd.ConfigList(cf, app, oneline)
}

func configSet(argv []string) error {
	usage := addGlobalFlags(`
Sets environment variables for an application.

Usage: deis config:set <var>=<value> [<var>=<value>...] [options]

Arguments:
  <var>
    the uniquely identifiable name for the environment variable.
  <value>
    the value of said environment variable.

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

	return cmd.ConfigSet(cf, app, args["<var>=<value>"].([]string))
}

func configUnset(argv []string) error {
	usage := addGlobalFlags(`
Unsets an environment variable for an application.

Usage: deis config:unset <key>... [options]

Arguments:
  <key>
    the variable to remove from the application's environment.

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

	return cmd.ConfigUnset(cf, app, args["<key>"].([]string))
}

func configPull(argv []string) error {
	usage := addGlobalFlags(`
Extract all environment variables from an application for local use.

The environmental variables can be piped into a file, 'deis config:pull > file',
or stored locally in a file named .env. This file can be
read by foreman to load the local environment for your app.

Usage: deis config:pull [options]

Options:
  -a --app=<app>
    The application that you wish to pull from
  -i --interactive
    Prompts for each value to be overwritten
  -o --overwrite
    Allows you to have the pull overwrite keys in .env
`)

	args, err := docopt.Parse(usage, argv, true, "", false, true)

	if err != nil {
		return err
	}

	app := safeGetValue(args, "--app")
	interactive := args["--interactive"].(bool)
	overwrite := args["--overwrite"].(bool)
	cf := safeGetValue(args, "--config")

	return cmd.ConfigPull(cf, app, interactive, overwrite)
}

func configPush(argv []string) error {
	usage := addGlobalFlags(`
Sets environment variables for an application.

This file can be read by foreman
to load the local environment for your app. The file should be piped via
stdin, 'deis config:push < .env', or using the --path option.

Usage: deis config:push [options]

Options:
  -a --app=<app>
    the uniquely identifiable name for the application.
  -p <path>, --path=<path>
    a path leading to an environment file [default: .env]
`)

	args, err := docopt.Parse(usage, argv, true, "", false, true)

	if err != nil {
		return err
	}

	cf := safeGetValue(args, "--config")
	app := safeGetValue(args, "--app")
	path := safeGetValue(args, "--path")

	return cmd.ConfigPush(cf, app, path)
}
