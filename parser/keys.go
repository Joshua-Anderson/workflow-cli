package parser

import (
	"io"

	"github.com/deis/workflow-cli/cmd"
	docopt "github.com/docopt/docopt-go"
)

// Keys routes key commands to the specific function.
func Keys(argv []string, wOut io.Writer, wErr io.Writer) error {
	usage := `
Valid commands for SSH keys:

keys:list        list SSH keys for the logged in user
keys:add         add an SSH key
keys:remove      remove an SSH key

Use 'deis help [command]' to learn more.
`

	switch argv[0] {
	case "keys:list":
		return keysList(argv, wOut)
	case "keys:add":
		return keyAdd(argv, wOut)
	case "keys:remove":
		return keyRemove(argv, wOut)
	default:
		if printHelp(argv, usage, wOut) {
			return nil
		}

		if argv[0] == "keys" {
			argv[0] = "keys:list"
			return keysList(argv, wOut)
		}

		PrintUsage(wErr)
		return nil
	}
}

func keysList(argv []string, wOut io.Writer) error {
	usage := addGlobalFlags(`
Lists SSH keys for the logged in user.

Usage: deis keys:list [options]

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

	return cmd.KeysList(safeGetValue(args, "--config"), results, wOut)
}

func keyAdd(argv []string, wOut io.Writer) error {
	usage := addGlobalFlags(`
Adds SSH keys for the logged in user.

Usage: deis keys:add [<key>]

Arguments:
  <key>
    a local file path to an SSH public key used to push application code.
`)

	args, err := docopt.Parse(usage, argv, true, "", false, true)

	if err != nil {
		return err
	}

	cf := safeGetValue(args, "--config")
	key := safeGetValue(args, "<key>")

	return cmd.KeyAdd(cf, key, wOut)
}

func keyRemove(argv []string, wOut io.Writer) error {
	usage := addGlobalFlags(`
Removes an SSH key for the logged in user.

Usage: deis keys:remove <key>

Arguments:
  <key>
    the SSH public key to revoke source code push access.
`)

	args, err := docopt.Parse(usage, argv, true, "", false, true)

	if err != nil {
		return err
	}

	cf := safeGetValue(args, "--config")
	key := safeGetValue(args, "<key>")

	return cmd.KeyRemove(cf, key, wOut)
}
