package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/deis/pkg/prettyprint"

	"github.com/deis/controller-sdk-go/api"
	"github.com/deis/controller-sdk-go/config"
)

// RegistryList lists an app's registry information.
func RegistryList(cf, appID string, wOut io.Writer) error {
	s, appID, err := load(cf, appID)

	if err != nil {
		return err
	}

	config, err := config.List(s.Client, appID)
	if checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}

	fmt.Fprintf(wOut, "=== %s Registry\n", appID)

	registryMap := make(map[string]string)

	for key, value := range config.Registry {
		registryMap[key] = fmt.Sprintf("%v", value)
	}

	fmt.Fprint(wOut, prettyprint.PrettyTabs(registryMap, 5))

	return nil
}

// RegistrySet sets an app's registry information.
func RegistrySet(cf, appID string, item []string, wOut io.Writer) error {
	s, appID, err := load(cf, appID)

	if err != nil {
		return err
	}

	registryMap, err := parseInfos(item)
	if err != nil {
		return err
	}

	fmt.Fprint(wOut, "Applying registry information... ")

	quit := progress(wOut)
	configObj := api.Config{}
	configObj.Registry = registryMap

	_, err = config.Set(s.Client, appID, configObj)
	quit <- true
	<-quit
	if checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}

	fmt.Fprint(wOut, "done\n\n")

	return RegistryList(cf, appID, wOut)
}

// RegistryUnset removes an app's registry information.
func RegistryUnset(cf, appID string, items []string, wOut io.Writer) error {
	s, appID, err := load(cf, appID)

	if err != nil {
		return err
	}

	fmt.Fprint(wOut, "Applying registry information... ")

	quit := progress(wOut)

	configObj := api.Config{}

	registryMap := make(map[string]interface{})

	for _, key := range items {
		registryMap[key] = nil
	}

	configObj.Registry = registryMap

	_, err = config.Set(s.Client, appID, configObj)
	quit <- true
	<-quit
	if checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}

	fmt.Fprint(wOut, "done\n\n")

	return RegistryList(cf, appID, wOut)
}

func parseInfos(items []string) (map[string]interface{}, error) {
	registryMap := make(map[string]interface{})

	for _, item := range items {
		key, value, err := parseInfo(item)

		if err != nil {
			return nil, err
		}

		registryMap[key] = value
	}

	return registryMap, nil
}

func parseInfo(item string) (string, string, error) {
	parts := strings.SplitN(item, "=", 2)

	if len(parts) != 2 {
		return "", "", fmt.Errorf(`%s is invalid. Must be in format key=value
Examples: username=bob password=s3cur3pw1`, item)
	}

	if parts[0] != "username" && parts[0] != "password" {
		return "", "", fmt.Errorf(`%s is invalid. Valid keys are "username" or "password"`, parts[0])
	}

	return parts[0], parts[1], nil
}
