package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/deis/pkg/prettyprint"

	"github.com/deis/controller-sdk-go/api"
	"github.com/deis/controller-sdk-go/config"
)

// TagsList lists an app's tags.
func TagsList(cf, appID string, wOut io.Writer) error {
	s, appID, err := load(cf, appID)

	if err != nil {
		return err
	}

	config, err := config.List(s.Client, appID)
	if checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}

	fmt.Fprintf(wOut, "=== %s Tags\n", appID)

	tagMap := make(map[string]string)

	for key, value := range config.Tags {
		tagMap[key] = fmt.Sprintf("%v", value)
	}

	fmt.Fprint(wOut, prettyprint.PrettyTabs(tagMap, 5))

	return nil
}

// TagsSet sets an app's tags.
func TagsSet(cf, appID string, tags []string, wOut io.Writer) error {
	s, appID, err := load(cf, appID)

	if err != nil {
		return err
	}

	tagsMap, err := parseTags(tags)
	if err != nil {
		return err
	}

	fmt.Fprint(wOut, "Applying tags... ")

	quit := progress(wOut)
	configObj := api.Config{}
	configObj.Tags = tagsMap

	_, err = config.Set(s.Client, appID, configObj)
	quit <- true
	<-quit
	if checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}

	fmt.Fprint(wOut, "done\n\n")

	return TagsList(cf, appID, wOut)
}

// TagsUnset removes an app's tags.
func TagsUnset(cf, appID string, tags []string, wOut io.Writer) error {
	s, appID, err := load(cf, appID)

	if err != nil {
		return err
	}

	fmt.Fprint(wOut, "Applying tags... ")

	quit := progress(wOut)

	configObj := api.Config{}

	tagsMap := make(map[string]interface{})

	for _, tag := range tags {
		tagsMap[tag] = nil
	}

	configObj.Tags = tagsMap

	_, err = config.Set(s.Client, appID, configObj)
	quit <- true
	<-quit
	if checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}

	fmt.Fprint(wOut, "done\n\n")

	return TagsList(cf, appID, wOut)
}

func parseTags(tags []string) (map[string]interface{}, error) {
	tagMap := make(map[string]interface{})

	for _, tag := range tags {
		key, value, err := parseTag(tag)

		if err != nil {
			return nil, err
		}

		tagMap[key] = value
	}

	return tagMap, nil
}

func parseTag(tag string) (string, string, error) {
	parts := strings.Split(tag, "=")

	if len(parts) != 2 {
		return "", "", fmt.Errorf(`%s is invalid, Must be in format key=value
Examples: rack=1 evironment=production`, tag)
	}

	return parts[0], parts[1], nil
}
