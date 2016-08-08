package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/deis/workflow-cli/cli"
	"github.com/deis/workflow-cli/parser"
	docopt "github.com/docopt/docopt-go"
)

// main exits with the return value of Command(os.Args[1:]), deferring all logic to
// a func we can test.
func main() {
	os.Exit(Command(os.Args[1:], os.Stdout, os.Stderr))
}

// Command routes deis commands to their proper parser.
func Command(argv []string, wOut io.Writer, wErr io.Writer) int {
	usage := `
The Deis command-line client issues API calls to a Deis controller.

Usage: deis <command> [<args>...]

Option flags::

  -h --help     display help information
  -v --version  display client version
  -c --config   (optional) path to configuration file. Equilivent to
                setting $DEIS_PROFILE. Defaults to ~/.deis/config.json.
                If not set to a filepath, will assume location ~/.deis/<value>.json

Auth commands, use 'deis help auth' to learn more::

  register      register a new user with a controller
  login         login to a controller
  logout        logout from the current controller

Subcommands, use 'deis help [subcommand]' to learn more::

  apps          manage applications used to provide services
  builds        manage builds created using 'git push'
  certs         manage SSL endpoints for an app
  config        manage environment variables that define app config
  domains       manage and assign domain names to your applications
  git           manage git for applications
  healthchecks  manage healthchecks for applications
  keys          manage ssh keys used for 'git push' deployments
  limits        manage resource limits for your application
  perms         manage permissions for applications
  ps            manage processes inside an app container
  registry      manage private registry information for your application
  releases      manage releases of an application
  routing       manage routability of an application
  tags          manage tags for application containers
  users         manage users
  version       display client version

Shortcut commands, use 'deis shortcuts' to see all::

  create        create a new application
  destroy       destroy an application
  info          view information about the current app
  logs          view aggregated log info for the app
  open          open a URL to the app in a browser
  pull          imports an image and deploys as a new release
  run           run a command in an ephemeral app container
  scale         scale processes by type (web=2, worker=1)

Use 'git push deis master' to deploy to an application.
`
	// Reorganize some command line flags and commands.
	command, argv := parseArgs(argv)
	// Give docopt an optional final false arg so it doesn't call os.Exit().
	_, err := docopt.Parse(usage, []string{command}, false, "", true, false)

	if err != nil {
		fmt.Fprintln(wErr, err)
		return 1
	}

	if len(argv) == 0 {
		fmt.Fprintln(wErr, "Usage: deis <command> [<args>...]")
		return 1
	}

	// Dispatch the command, passing the argv through so subcommands can
	// re-parse it according to their usage strings.
	switch command {
	case "apps":
		err = parser.Apps(argv, wOut, wErr)
	case "auth":
		err = parser.Auth(argv, wOut, wErr)
	case "builds":
		err = parser.Builds(argv, wOut, wErr)
	case "certs":
		err = parser.Certs(argv, wOut, wErr)
	case "config":
		err = parser.Config(argv, wOut, wErr)
	case "domains":
		err = parser.Domains(argv, wOut, wErr)
	case "git":
		err = parser.Git(argv, wOut, wErr)
	case "healthchecks":
		err = parser.Healthchecks(argv, wOut, wErr)
	case "help":
		fmt.Fprint(wOut, usage)
		return 0
	case "keys":
		err = parser.Keys(argv, wOut, wErr)
	case "limits":
		err = parser.Limits(argv, wOut, wErr)
	case "perms":
		err = parser.Perms(argv, wOut, wErr)
	case "ps":
		err = parser.Ps(argv, wOut, wErr)
	case "registry":
		err = parser.Registry(argv, wOut, wErr)
	case "releases":
		err = parser.Releases(argv, wOut, wErr)
	case "routing":
		err = parser.Routing(argv, wOut, wErr)
	case "shortcuts":
		err = parser.Shortcuts(argv, wOut, wErr)
	case "tags":
		err = parser.Tags(argv, wOut, wErr)
	case "users":
		err = parser.Users(argv, wOut, wErr)
	case "version":
		err = parser.Version(argv, wOut)
	default:
		env := os.Environ()
		extCmd := "deis-" + command

		binary, err := exec.LookPath(extCmd)
		if err != nil {
			parser.PrintUsage(wErr)
			return 1
		}

		cmdArgv := []string{extCmd}

		cmdSplit := strings.Split(argv[0], command+":")

		if len(cmdSplit) > 1 {
			argv[0] = cmdSplit[1]
		}

		cmdArgv = append(cmdArgv, argv...)

		err = syscall.Exec(binary, cmdArgv, env)
		if err != nil {
			parser.PrintUsage(wErr)
			return 1
		}
	}
	if err != nil {
		fmt.Fprintf(wErr, "Error: %v\n", err)
		return 1
	}
	return 0
}

// parseArgs returns the provided args with "--help" as the last arg if need be,
// expands shortcuts and formats commands to be properly routed.
func parseArgs(argv []string) (string, []string) {
	if len(argv) == 1 {
		if argv[0] == "--help" || argv[0] == "-h" {
			// rearrange "deis --help" as "deis help"
			argv[0] = "help"
		} else if argv[0] == "--version" || argv[0] == "-v" {
			// rearrange "deis --version" as "deis version"
			argv[0] = "version"
		}
	}

	if len(argv) > 1 {
		// Rearrange "deis help <command>" to "deis <command> --help".
		if argv[0] == "help" || argv[0] == "--help" || argv[0] == "-h" {
			argv = append(argv[1:], "--help")
		}
	}

	if len(argv) > 0 {
		argv[0] = replaceShortcut(argv[0])

		index := strings.Index(argv[0], ":")

		if index != -1 {
			command := argv[0]
			return command[:index], argv
		}

		return argv[0], argv
	}

	return "", argv
}

func replaceShortcut(command string) string {
	expandedCommand := cli.Shortcuts[command]
	if expandedCommand == "" {
		return command
	}

	return expandedCommand
}
