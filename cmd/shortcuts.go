package cmd

import (
	"fmt"
	"io"
	"sort"

	"github.com/deis/workflow-cli/cli"
)

// ShortcutsList displays all relevant shortcuts for the CLI.
func ShortcutsList(wOut io.Writer) error {
	fmt.Fprintln(wOut, sortShortcuts())

	return nil
}

func sortShortcuts() string {
	var (
		strBuilder string
		keys       []string
	)

	// NOTE(bacongobbler): go does not guarantee an iteration order when iterating over a map,
	// so to work around this we can sort the keys and iterate using the key array
	for k := range cli.Shortcuts {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		strBuilder += fmt.Sprintf("%s -> %s\n", k, cli.Shortcuts[k])
	}

	return strBuilder
}
