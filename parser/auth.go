package parser

import (
	"fmt"
	"io"

	"github.com/deis/workflow-cli/cmd"
	docopt "github.com/docopt/docopt-go"
)

// Auth routes auth commands to the specific function.
func Auth(argv []string, wOut io.Writer, wErr io.Writer) error {
	usage := `
Valid commands for auth:

auth:register          register a new user
auth:login             authenticate against a controller
auth:logout            clear the current user session
auth:passwd            change the password for the current user
auth:whoami            display the current user
auth:cancel            remove the current user account
auth:regenerate        regenerate user tokens

Use 'deis help [command]' to learn more.
`

	switch argv[0] {
	case "auth:register":
		return authRegister(argv, wOut)
	case "auth:login":
		return authLogin(argv, wOut)
	case "auth:logout":
		return authLogout(argv, wOut)
	case "auth:passwd":
		return authPasswd(argv, wOut)
	case "auth:whoami":
		return authWhoami(argv, wOut)
	case "auth:cancel":
		return authCancel(argv, wOut)
	case "auth:regenerate":
		return authRegenerate(argv, wOut)
	case "auth":
		fmt.Fprint(wOut, usage)
		return nil
	default:
		PrintUsage(wErr)
		return nil
	}
}

func authRegister(argv []string, wOut io.Writer) error {
	usage := addGlobalFlags(`
Registers a new user with a Deis controller.

Usage: deis auth:register <controller> [options]

Arguments:
  <controller>
    fully-qualified controller URI, e.g. 'http://deis.local3.deisapp.com/'

Options:
  --username=<username>
    provide a username for the new account.
  --password=<password>
    provide a password for the new account.
  --email=<email>
    provide an email address.
  --ssl-verify=false
    disables SSL certificate verification for API requests
`)

	args, err := docopt.Parse(usage, argv, true, "", false, true)

	if err != nil {
		return err
	}

	controller := safeGetValue(args, "<controller>")
	username := safeGetValue(args, "--username")
	password := safeGetValue(args, "--password")
	email := safeGetValue(args, "--email")
	cf := safeGetValue(args, "--config")
	sslVerify := false

	if args["--ssl-verify"] != nil && args["--ssl-verify"].(string) == "true" {
		sslVerify = true
	}

	return cmd.Register(cf, controller, username, password, email, sslVerify, wOut)
}

func authLogin(argv []string, wOut io.Writer) error {
	usage := addGlobalFlags(`
Logs in by authenticating against a controller.

Usage: deis auth:login <controller> [options]

Arguments:
  <controller>
    a fully-qualified controller URI, e.g. "http://deis.local3.deisapp.com/".

Options:
  --username=<username>
    provide a username for the account.
  --password=<password>
    provide a password for the account.
  --ssl-verify=false
    disables SSL certificate verification for API requests
`)

	args, err := docopt.Parse(usage, argv, true, "", false, true)

	if err != nil {
		return err
	}

	controller := safeGetValue(args, "<controller>")
	username := safeGetValue(args, "--username")
	password := safeGetValue(args, "--password")
	cf := safeGetValue(args, "--config")
	sslVerify := false

	if args["--ssl-verify"] != nil && args["--ssl-verify"].(string) == "true" {
		sslVerify = true
	}

	return cmd.Login(cf, controller, username, password, sslVerify, wOut)
}

func authLogout(argv []string, wOut io.Writer) error {
	usage := addGlobalFlags(`
Logs out from a controller and clears the user session.

Usage: deis auth:logout

Options:
`)

	args, err := docopt.Parse(usage, argv, true, "", false, true)
	if err != nil {
		return err
	}

	return cmd.Logout(safeGetValue(args, "--config"), wOut)
}

func authPasswd(argv []string, wOut io.Writer) error {
	usage := addGlobalFlags(`
Changes the password for the current user.

Usage: deis auth:passwd [options]

Options:
  --password=<password>
    the current password for the account.
  --new-password=<new-password>
    the new password for the account.
  --username=<username>
    the account's username.
`)

	args, err := docopt.Parse(usage, argv, true, "", false, true)

	if err != nil {
		return err
	}

	cf := safeGetValue(args, "--config")
	username := safeGetValue(args, "--username")
	password := safeGetValue(args, "--password")
	newPassword := safeGetValue(args, "--new-password")

	return cmd.Passwd(cf, username, password, newPassword, wOut)
}

func authWhoami(argv []string, wOut io.Writer) error {
	usage := addGlobalFlags(`
Displays the currently logged in user.

Usage: deis auth:whoami [options]

Options:
  --all
    fetch a more detailed description about the user.
`)

	args, err := docopt.Parse(usage, argv, true, "", false, true)

	if err != nil {
		return err
	}

	return cmd.Whoami(safeGetValue(args, "--config"), args["--all"].(bool), wOut)
}

func authCancel(argv []string, wOut io.Writer) error {
	usage := addGlobalFlags(`
Cancels and removes the current account.

Usage: deis auth:cancel [options]

Options:
  --username=<username>
    provide a username for the account.
  --password=<password>
    provide a password for the account.
  --yes
    force "yes" when prompted.
`)

	args, err := docopt.Parse(usage, argv, true, "", false, true)

	if err != nil {
		return err
	}

	username := safeGetValue(args, "--username")
	password := safeGetValue(args, "--password")
	cf := safeGetValue(args, "--config")
	yes := args["--yes"].(bool)

	return cmd.Cancel(cf, username, password, yes, wOut)
}

func authRegenerate(argv []string, wOut io.Writer) error {
	usage := addGlobalFlags(`
Regenerates auth token, defaults to regenerating token for the current user.

Usage: deis auth:regenerate [options]

Options:
  -u --username=<username>
    specify user to regenerate. Requires admin privilages.
  --all
    regenerate token for every user. Requires admin privilages.
`)

	args, err := docopt.Parse(usage, argv, true, "", false, true)

	if err != nil {
		return err
	}

	username := safeGetValue(args, "--username")
	cf := safeGetValue(args, "--config")
	all := args["--all"].(bool)

	return cmd.Regenerate(cf, username, all, wOut)
}
