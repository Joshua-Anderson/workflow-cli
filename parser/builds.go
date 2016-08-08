package parser

import (
	"io"

	"github.com/deis/workflow-cli/cmd"
	docopt "github.com/docopt/docopt-go"
)

// Builds routes build commands to their specific function.
func Builds(argv []string, wOut io.Writer, wErr io.Writer) error {
	usage := `
Valid commands for builds:

builds:list        list build history for an application
builds:create      imports an image and deploys as a new release

Use 'deis help [command]' to learn more.
`

	switch argv[0] {
	case "builds:list":
		return buildsList(argv, wOut)
	case "builds:create":
		return buildsCreate(argv, wOut)
	default:
		if printHelp(argv, usage, wOut) {
			return nil
		}

		if argv[0] == "builds" {
			argv[0] = "builds:list"
			return buildsList(argv, wOut)
		}

		PrintUsage(wErr)
		return nil
	}
}

func buildsList(argv []string, wOut io.Writer) error {
	usage := addGlobalFlags(`
Lists build history for an application.

Usage: deis builds:list [options]

Options:
  -a --app=<app>
    the uniquely identifiable name for the application.
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

	return cmd.BuildsList(safeGetValue(args, "--config"), safeGetValue(args, "--app"), results, wOut)
}

func buildsCreate(argv []string, wOut io.Writer) error {
	usage := addGlobalFlags(`
Creates a new build of an application. Imports an <image> and deploys it to Deis
as a new release. If a Procfile is present in the current directory, it will be used
as the default process types for this application.

Usage: deis builds:create <image> [options]

Arguments:
  <image>
    A fully-qualified docker image, either from Docker Hub (e.g. deis/example-go:latest)
    or from an in-house registry (e.g. myregistry.example.com:5000/example-go:latest).
    This image must include the tag.

Options:
  -a --app=<app>
    The uniquely identifiable name for the application.
  -p --procfile=<procfile>
    A YAML string used to supply a Procfile to the application.
`)

	args, err := docopt.Parse(usage, argv, true, "", false, true)

	if err != nil {
		return err
	}

	app := safeGetValue(args, "--app")
	image := safeGetValue(args, "<image>")
	procfile := safeGetValue(args, "--procfile")
	cf := safeGetValue(args, "--config")

	return cmd.BuildsCreate(cf, app, image, procfile, wOut)
}
