package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"

	"github.com/deis/controller-sdk-go/builds"
)

// BuildsList lists an app's builds.
func BuildsList(cf, appID string, results int, wOut io.Writer) error {
	s, appID, err := load(cf, appID)

	if err != nil {
		return err
	}

	if results == defaultLimit {
		results = s.Limit
	}

	builds, count, err := builds.List(s.Client, appID, results)
	if checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}

	fmt.Fprintf(wOut, "=== %s Builds%s", appID, limitCount(len(builds), count))

	for _, build := range builds {
		fmt.Fprintln(wOut, build.UUID, build.Created)
	}
	return nil
}

// BuildsCreate creates a build for an app.
func BuildsCreate(cf, appID, image, procfile string, wOut io.Writer) error {
	s, appID, err := load(cf, appID)

	if err != nil {
		return err
	}

	procfileMap := make(map[string]string)

	if procfile != "" {
		if procfileMap, err = parseProcfile([]byte(procfile)); err != nil {
			return err
		}
	} else if _, err := os.Stat("Procfile"); err == nil {
		contents, err := ioutil.ReadFile("Procfile")
		if err != nil {
			return err
		}

		if procfileMap, err = parseProcfile(contents); err != nil {
			return err
		}
	}

	fmt.Fprint(wOut, "Creating build... ")
	quit := progress(wOut)
	_, err = builds.New(s.Client, appID, image, procfileMap)
	quit <- true
	<-quit
	if checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}

	fmt.Fprintln(wOut, "done")

	return nil
}

func parseProcfile(procfile []byte) (map[string]string, error) {
	procfileMap := make(map[string]string)
	return procfileMap, yaml.Unmarshal(procfile, &procfileMap)
}
