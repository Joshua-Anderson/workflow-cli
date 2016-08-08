package parser

import (
	"io"

	"github.com/deis/workflow-cli/cmd"
	docopt "github.com/docopt/docopt-go"
)

// Tags routes tags commands to their specific function
func Tags(argv []string, wOut io.Writer, wErr io.Writer) error {
	usage := `
Valid commands for tags:

tags:list        list tags for an app
tags:set         set tags for an app
tags:unset       unset tags for an app

Use 'deis help [command]' to learn more.
`

	switch argv[0] {
	case "tags:list":
		return tagsList(argv, wOut)
	case "tags:set":
		return tagsSet(argv, wOut)
	case "tags:unset":
		return tagsUnset(argv, wOut)
	default:
		if printHelp(argv, usage, wOut) {
			return nil
		}

		if argv[0] == "tags" {
			argv[0] = "tags:list"
			return tagsList(argv, wOut)
		}

		PrintUsage(wErr)
		return nil
	}
}

func tagsList(argv []string, wOut io.Writer) error {
	usage := addGlobalFlags(`
Lists tags for an application.

Usage: deis tags:list [options]

Options:
  -a --app=<app>
    the uniquely identifiable name of the application.
`)

	args, err := docopt.Parse(usage, argv, true, "", false, true)

	if err != nil {
		return err
	}

	return cmd.TagsList(safeGetValue(args, "--config"), safeGetValue(args, "--app"), wOut)
}

func tagsSet(argv []string, wOut io.Writer) error {
	usage := addGlobalFlags(`
Sets tags for an application.

A tag is a key/value pair used to tag an application's containers and is passed to the
scheduler. This is often used to restrict workloads to specific hosts matching the
scheduler-configured metadata.

Usage: deis tags:set [options] <key>=<value>...

Arguments:
  <key> the tag key, for example: "environ" or "rack"
  <value> the tag value, for example: "prod" or "1"

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
	tags := args["<key>=<value>"].([]string)

	return cmd.TagsSet(cf, app, tags, wOut)
}

func tagsUnset(argv []string, wOut io.Writer) error {
	usage := addGlobalFlags(`
Unsets tags for an application.

Usage: deis tags:unset [options] <key>...

Arguments:
  <key> the tag key to unset, for example: "environ" or "rack"

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
	tags := args["<key>"].([]string)

	return cmd.TagsUnset(cf, app, tags, wOut)
}
