package cmd

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/deis/controller-sdk-go/releases"
)

// ReleasesList lists an app's releases.
func ReleasesList(cf, appID string, results int, wOut io.Writer) error {
	s, appID, err := load(cf, appID)

	if err != nil {
		return err
	}

	if results == defaultLimit {
		results = s.Limit
	}

	releases, count, err := releases.List(s.Client, appID, results)
	if checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}

	fmt.Fprintf(wOut, "=== %s Releases%s", appID, limitCount(len(releases), count))

	w := new(tabwriter.Writer)

	w.Init(os.Stdout, 0, 8, 1, '\t', 0)
	for _, r := range releases {
		fmt.Fprintf(w, "v%d\t%s\t%s\n", r.Version, r.Created, r.Summary)
	}
	w.Flush()
	return nil
}

// ReleasesInfo prints info about a specific release.
func ReleasesInfo(cf, appID string, version int, wOut io.Writer) error {
	s, appID, err := load(cf, appID)

	if err != nil {
		return err
	}

	r, err := releases.Get(s.Client, appID, version)
	if checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}

	fmt.Fprintf(wOut, "=== %s Release v%d\n", appID, version)
	if r.Build != "" {
		fmt.Fprintln(wOut, "build:   ", r.Build)
	}
	fmt.Fprintln(wOut, "config:  ", r.Config)
	fmt.Fprintln(wOut, "owner:   ", r.Owner)
	fmt.Fprintln(wOut, "created: ", r.Created)
	fmt.Fprintln(wOut, "summary: ", r.Summary)
	fmt.Fprintln(wOut, "updated: ", r.Updated)
	fmt.Fprintln(wOut, "uuid:    ", r.UUID)

	return nil
}

// ReleasesRollback rolls an app back to a previous release.
func ReleasesRollback(cf, appID string, version int, wOut io.Writer) error {
	s, appID, err := load(cf, appID)

	if err != nil {
		return err
	}

	if version == -1 {
		fmt.Fprint(wOut, "Rolling back one release... ")
	} else {
		fmt.Fprintf(wOut, "Rolling back to v%d... ", version)
	}

	quit := progress(wOut)
	newVersion, err := releases.Rollback(s.Client, appID, version)
	quit <- true
	<-quit
	if checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}

	fmt.Fprintf(wOut, "done, v%d\n", newVersion)

	return nil
}
