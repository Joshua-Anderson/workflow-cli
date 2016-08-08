package parser

import (
	"io"

	"github.com/deis/workflow-cli/cmd"
	docopt "github.com/docopt/docopt-go"
)

// Domains routes domain commands to their specific function.
func Domains(argv []string, wOut io.Writer, wErr io.Writer) error {
	usage := `
Valid commands for domains:

domains:add           bind a domain to an application
domains:list          list domains bound to an application
domains:remove        unbind a domain from an application

Use 'deis help [command]' to learn more.
`

	switch argv[0] {
	case "domains:add":
		return domainsAdd(argv, wOut)
	case "domains:list":
		return domainsList(argv, wOut)
	case "domains:remove":
		return domainsRemove(argv, wOut)
	default:
		if printHelp(argv, usage, wOut) {
			return nil
		}

		if argv[0] == "domains" {
			argv[0] = "domains:list"
			return domainsList(argv, wOut)
		}

		PrintUsage(wErr)
		return nil
	}
}

func domainsAdd(argv []string, wOut io.Writer) error {
	usage := addGlobalFlags(`
Binds a domain to an application.

Usage: deis domains:add <domain> [options]

Arguments:
  <domain>
    the domain name to be bound to the application, such as 'domain.deisapp.com'.

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
	domain := safeGetValue(args, "<domain>")

	return cmd.DomainsAdd(cf, app, domain, wOut)
}

func domainsList(argv []string, wOut io.Writer) error {
	usage := addGlobalFlags(`
Lists domains bound to an application.

Usage: deis domains:list [options]

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

	cf := safeGetValue(args, "--config")
	app := safeGetValue(args, "--app")

	return cmd.DomainsList(cf, app, results, wOut)
}

func domainsRemove(argv []string, wOut io.Writer) error {
	usage := addGlobalFlags(`
Unbinds a domain for an application.

Usage: deis domains:remove <domain> [options]

Arguments:
  <domain>
    the domain name to be removed from the application.

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
	domain := safeGetValue(args, "<domain>")

	return cmd.DomainsRemove(cf, app, domain, wOut)
}
