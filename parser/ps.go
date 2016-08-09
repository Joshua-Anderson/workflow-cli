package parser

import (
	"github.com/deis/workflow-cli/cmd"
	docopt "github.com/docopt/docopt-go"
)

// Ps routes ps commands to their specific function.
func Ps(argv []string) error {
	usage := `
Valid commands for processes:

ps:list        list application processes
ps:restart     restart an application or its process types
ps:scale       scale processes (e.g. web=4 worker=2)

Use 'deis help [command]' to learn more.
`

	switch argv[0] {
	case "ps:list":
		return psList(argv)
	case "ps:restart":
		return psRestart(argv)
	case "ps:scale":
		return psScale(argv)
	default:
		if printHelp(argv, usage) {
			return nil
		}

		if argv[0] == "ps" {
			argv[0] = "ps:list"
			return psList(argv)
		}

		PrintUsage()
		return nil
	}
}

func psList(argv []string) error {
	usage := addGlobalFlags(`
Lists processes servicing an application.

Usage: deis ps:list [options]

Options:
  -a --app=<app>
    the uniquely identifiable name for the application.
`)

	args, err := docopt.Parse(usage, argv, true, "", false, true)
	if err != nil {
		return err
	}

	// The 1000 is fake for now until API understands limits
	return cmd.PsList(safeGetValue(args, "--config"), safeGetValue(args, "--app"), 1000)
}

func psRestart(argv []string) error {
	usage := addGlobalFlags(`
Restart an application, a process type or a specific process.

Usage: deis ps:restart [<type>] [options]

Arguments:
  <type>
    the process name as defined in your Procfile, such as 'web' or 'worker'.
    To restart a particular process, use 'web-asdfg' or 'app-v2-web-asdfg'.

Options:
  -a --app=<app>
    the uniquely identifiable name for the application.
`)

	args, err := docopt.Parse(usage, argv, true, "", false, true)

	if err != nil {
		return err
	}

	cf := safeGetValue(args, "--config")
	apps := safeGetValue(args, "--app")
	tp := safeGetValue(args, "<type>")
	return cmd.PsRestart(cf, apps, tp)
}

func psScale(argv []string) error {
	usage := addGlobalFlags(`
Scales an application's processes by type.

Usage: deis ps:scale <type>=<num>... [options]

Arguments:
  <type>
    the process name as defined in your Procfile, such as 'web' or 'worker'.
    Note that Dockerfile apps have a default 'cmd' process type.
  <num>
    the number of processes.

Options:
  -a --app=<app>
    the uniquely identifiable name for the application.
`)

	args, err := docopt.Parse(usage, argv, true, "", false, true)

	if err != nil {
		return err
	}

	cf := safeGetValue(args, "--config")
	apps := safeGetValue(args, "--app")
	return cmd.PsScale(cf, apps, args["<type>=<num>"].([]string))
}
