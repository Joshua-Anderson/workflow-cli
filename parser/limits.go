package parser

import (
	"io"

	"github.com/deis/workflow-cli/cmd"
	docopt "github.com/docopt/docopt-go"
)

// Limits routes limits commands to their specific function
func Limits(argv []string, wOut io.Writer, wErr io.Writer) error {
	usage := `
Valid commands for limits:

limits:list        list resource limits for an app
limits:set         set resource limits for an app
limits:unset       unset resource limits for an app

Use 'deis help [command]' to learn more.
`

	switch argv[0] {
	case "limits:list":
		return limitsList(argv, wOut)
	case "limits:set":
		return limitSet(argv, wOut)
	case "limits:unset":
		return limitUnset(argv, wOut)
	default:
		if printHelp(argv, usage, wOut) {
			return nil
		}

		if argv[0] == "limits" {
			argv[0] = "limits:list"
			return limitsList(argv, wOut)
		}

		PrintUsage(wErr)
		return nil
	}
}

func limitsList(argv []string, wOut io.Writer) error {
	usage := addGlobalFlags(`
Lists resource limits for an application.

Usage: deis limits:list [options]

Options:
  -a --app=<app>
    the uniquely identifiable name of the application.
`)

	args, err := docopt.Parse(usage, argv, true, "", false, true)

	if err != nil {
		return err
	}

	return cmd.LimitsList(safeGetValue(args, "--config"), safeGetValue(args, "--app"), wOut)
}

func limitSet(argv []string, wOut io.Writer) error {
	usage := addGlobalFlags(`
Sets resource limits for an application.

A resource limit is a finite resource within a pod which we can apply
restrictions through Kubernetes. This limit is applied to each individual
pod, so setting a memory limit of 1G for an application means that each
pod gets 1G of memory.

Usage: deis limits:set [options] <type>=<limit>...

Arguments:
  <type>
    the process type as defined in your Procfile, such as 'web' or 'worker'.
    Note that Dockerfile apps have a default 'cmd' process type.
  <limit>
    The limit to apply to the process type. By default, this is set to --memory.
    You can only set one type of limit per call.

    With --memory, units are represented in Bytes (B), Kilobytes (K), Megabytes
    (M), or Gigabytes (G). For example, 'deis limit:set cmd=1G' will restrict all
    "cmd" processes to a maximum of 1 Gigabyte of memory each.

    With --cpu, units are represented in the number of CPUs. For example,
    'deis limit:set --cpu cmd=1' will restrict all "cmd" processes to a
    maximum of 1 CPU. Alternatively, you can also use milli units to specify the
    number of CPU shares the pod can use. For example, 'deis limits:set --cpu cmd=500m'
    will restrict all "cmd" processes to half of a CPU.

Options:
  -a --app=<app>
    the uniquely identifiable name for the application.
  -c --cpu
    limits CPU.
  -m --memory
    limits memory. [default: true]
`)

	args, err := docopt.Parse(usage, argv, true, "", false, true)

	if err != nil {
		return err
	}

	cf := safeGetValue(args, "--config")
	app := safeGetValue(args, "--app")
	limits := args["<type>=<limit>"].([]string)
	limitType := "memory"

	if args["--cpu"].(bool) {
		limitType = "cpu"
	}

	return cmd.LimitsSet(cf, app, limits, limitType, wOut)
}

func limitUnset(argv []string, wOut io.Writer) error {
	usage := addGlobalFlags(`
Unsets resource limits for an application.

Usage: deis limits:unset [options] [--memory | --cpu] <type>...

Arguments:
  <type>
    the process type as defined in your Procfile, such as 'web' or 'worker'.
    Note that Dockerfile apps have a default 'cmd' process type.

Options:
  -a --app=<app>
    the uniquely identifiable name for the application.
  -c --cpu
    limits cpu shares.
  -m --memory
    limits memory. [default: true]
`)

	args, err := docopt.Parse(usage, argv, true, "", false, true)

	if err != nil {
		return err
	}

	cf := safeGetValue(args, "--config")
	app := safeGetValue(args, "--app")
	limits := args["<type>"].([]string)
	limitType := "memory"

	if args["--cpu"].(bool) {
		limitType = "cpu"
	}

	return cmd.LimitsUnset(cf, app, limits, limitType, wOut)
}
