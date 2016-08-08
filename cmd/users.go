package cmd

import (
	"fmt"
	"io"

	"github.com/deis/controller-sdk-go/users"
	"github.com/deis/workflow-cli/settings"
)

// UsersList lists users registered with the controller.
func UsersList(cf string, results int, wOut io.Writer) error {
	s, err := settings.Load(cf)

	if err != nil {
		return err
	}

	if results == defaultLimit {
		results = s.Limit
	}

	users, count, err := users.List(s.Client, results)
	if checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}

	fmt.Fprintf(wOut, "=== Users%s", limitCount(len(users), count))

	for _, user := range users {
		fmt.Fprintln(wOut, user.Username)
	}
	return nil
}
