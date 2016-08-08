package parser

import (
	"fmt"
	"io"

	"github.com/deis/workflow-cli/version"
	docopt "github.com/docopt/docopt-go"
)

// Version displays the client version
func Version(argv []string, wOut io.Writer) error {
	usage := `
Displays the client version.

Usage: deis version

Use 'deis help [command]' to learn more.
`
	if _, err := docopt.Parse(usage, argv, true, "", false, true); err != nil {
		return err
	}

	fmt.Fprintln(wOut, version.Version)

	return nil
}
