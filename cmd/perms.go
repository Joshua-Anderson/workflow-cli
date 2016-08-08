package cmd

import (
	"fmt"
	"io"

	"github.com/deis/controller-sdk-go/perms"
	"github.com/deis/workflow-cli/pkg/git"
	"github.com/deis/workflow-cli/settings"
)

// PermsList prints which users have permissions.
func PermsList(cf, appID string, admin bool, results int, wOut io.Writer) error {
	s, appID, err := permsLoad(cf, appID, admin)

	if err != nil {
		return err
	}

	var users []string
	var count int

	if admin {
		if results == defaultLimit {
			results = s.Limit
		}
		users, count, err = perms.ListAdmins(s.Client, results)
	} else {
		users, err = perms.List(s.Client, appID)
	}

	if checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}

	if admin {
		fmt.Fprintf(wOut, "=== Administrators%s", limitCount(len(users), count))
	} else {
		fmt.Fprintf(wOut, "=== %s's Users\n", appID)
	}

	for _, user := range users {
		fmt.Fprintln(wOut, user)
	}

	return nil
}

// PermCreate adds a user to an app or makes them an administrator.
func PermCreate(cf, appID string, username string, admin bool, wOut io.Writer) error {

	s, appID, err := permsLoad(cf, appID, admin)

	if err != nil {
		return err
	}

	if admin {
		fmt.Fprintf(wOut, "Adding %s to system administrators... ", username)
		err = perms.NewAdmin(s.Client, username)
	} else {
		fmt.Fprintf(wOut, "Adding %s to %s collaborators... ", username, appID)
		err = perms.New(s.Client, appID, username)
	}

	if checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}

	fmt.Fprintln(wOut, "done")

	return nil
}

// PermDelete removes a user from an app or revokes admin privileges.
func PermDelete(cf, appID, username string, admin bool, wOut io.Writer) error {

	s, appID, err := permsLoad(cf, appID, admin)

	if err != nil {
		return err
	}

	if admin {
		fmt.Fprintf(wOut, "Removing %s from system administrators... ", username)
		err = perms.DeleteAdmin(s.Client, username)
	} else {
		fmt.Fprintf(wOut, "Removing %s from %s collaborators... ", username, appID)
		err = perms.Delete(s.Client, appID, username)
	}

	if checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}

	fmt.Fprintln(wOut, "done")

	return nil
}

func permsLoad(cf, appID string, admin bool) (*settings.Settings, string, error) {
	s, err := settings.Load(cf)

	if err != nil {
		return nil, "", err
	}

	if !admin && appID == "" {
		appID, err = git.DetectAppName(s.Client.ControllerURL.Host)

		if err != nil {
			return nil, "", err
		}
	}

	return s, appID, err
}
