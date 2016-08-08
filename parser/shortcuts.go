package parser

import (
	"io"

	"github.com/deis/workflow-cli/cmd"
	docopt "github.com/docopt/docopt-go"
)

// Shortcuts displays all relevant shortcuts for the CLI.
func Shortcuts(argv []string, wOut io.Writer, wErr io.Writer) error {
	usage := `
Valid commands for shortcuts:

shortcuts:list       list all relevant shortcuts for the CLI

Use 'deis help [command]' to learn more.
`

	switch argv[0] {
	case "shortcuts:list":
		return shortcutsList(argv, wOut)
	default:
		if printHelp(argv, usage, wOut) {
			return nil
		}

		if argv[0] == "shortcuts" {
			argv[0] = "shortcuts:list"
			return shortcutsList(argv, wOut)
		}

		PrintUsage(wErr)
		return nil
	}
}

func shortcutsList(argv []string, wOut io.Writer) error {
	usage := `
Lists all relevant shortcuts for the CLI

Usage: deis shortcuts:list
`

	_, err := docopt.Parse(usage, argv, true, "", false, true)

	if err != nil {
		return err
	}

	return cmd.ShortcutsList(wOut)
}
