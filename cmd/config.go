package cmd

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/deis/pkg/prettyprint"

	"github.com/deis/controller-sdk-go/api"
	"github.com/deis/controller-sdk-go/config"
)

// ConfigList lists an app's config.
func ConfigList(cf, appID string, oneLine bool, wOut io.Writer) error {
	s, appID, err := load(cf, appID)

	if err != nil {
		return err
	}

	config, err := config.List(s.Client, appID)
	if checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}

	var keys []string
	for k := range config.Values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	if oneLine {
		cPs := sortedConfig(config.Values)
		for i, cP := range cPs {
			sep := " "
			if i == len(cPs)-1 {
				sep = "\n"
			}
			fmt.Fprintf(wOut, "%s=%v%s", cP.Key, cP.Value, sep)
		}
	} else {
		fmt.Fprintf(wOut, "=== %s Config\n", appID)

		configMap := make(map[string]string)

		// config.Values is type interface, so it needs to be converted to a string
		for _, key := range keys {
			configMap[key] = fmt.Sprintf("%v", config.Values[key])
		}

		fmt.Fprint(wOut, prettyprint.PrettyTabs(configMap, 6))
	}

	return nil
}

// ConfigSet sets an app's config variables.
func ConfigSet(cf, appID string, configVars []string, wOut io.Writer) error {
	s, appID, err := load(cf, appID)

	if err != nil {
		return err
	}

	configMap, err := parseConfig(configVars)

	if err != nil {
		return err
	}

	value, ok := configMap["SSH_KEY"]

	if ok {
		sshKey := value.(string)

		if _, err = os.Stat(value.(string)); err == nil {
			contents, err := ioutil.ReadFile(value.(string))

			if err != nil {
				return err
			}

			sshKey = string(contents)
		}

		sshRegex := regexp.MustCompile("^-.+ .SA PRIVATE KEY-*")

		if !sshRegex.MatchString(sshKey) {
			return fmt.Errorf("Could not parse SSH private key:\n %s", sshKey)
		}

		configMap["SSH_KEY"] = base64.StdEncoding.EncodeToString([]byte(sshKey))
	}

	// NOTE(bacongobbler): check if the user is using the old way to set healthchecks. If so,
	// send them a deprecation notice.
	for key := range configMap {
		if strings.Contains(key, "HEALTHCHECK_") {
			fmt.Fprintln(wOut, `Hey there! We've noticed that you're using 'deis config:set HEALTHCHECK_URL'
to set up healthchecks. This functionality has been deprecated. In the future, please use
'deis healthchecks' to set up application health checks. Thanks!`)
		}
	}

	fmt.Fprint(wOut, "Creating config... ")

	quit := progress(wOut)
	configObj := api.Config{Values: configMap}
	_, err = config.Set(s.Client, appID, configObj)
	quit <- true
	<-quit
	if checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}
	fmt.Fprint(wOut, "done\n\n")

	return ConfigList(cf, appID, false, wOut)
}

// ConfigUnset removes a config variable from an app.
func ConfigUnset(cf, appID string, configVars []string, wOut io.Writer) error {
	s, appID, err := load(cf, appID)

	if err != nil {
		return err
	}

	fmt.Fprint(wOut, "Removing config... ")

	quit := progress(wOut)

	configObj := api.Config{}

	valuesMap := make(map[string]interface{})

	for _, configVar := range configVars {
		valuesMap[configVar] = nil
	}

	configObj.Values = valuesMap

	_, err = config.Set(s.Client, appID, configObj)
	quit <- true
	<-quit
	if checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}

	fmt.Fprint(wOut, "done\n\n")

	return ConfigList(cf, appID, false, wOut)
}

// ConfigPull pulls an app's config to a file.
func ConfigPull(cf, appID string, interactive bool, overwrite bool, wOut io.Writer) error {
	s, appID, err := load(cf, appID)

	if err != nil {
		return err
	}

	configVars, err := config.List(s.Client, appID)
	if checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}

	stat, err := os.Stdout.Stat()

	if err != nil {
		return err
	}

	if (stat.Mode() & os.ModeCharDevice) == 0 {
		fmt.Fprint(wOut, formatConfig(configVars.Values))
		return nil
	}

	filename := ".env"

	if !overwrite {
		if _, err := os.Stat(filename); err == nil {
			return fmt.Errorf("%s already exists, pass -o to overwrite", filename)
		}
	}

	if interactive {
		contents, err := ioutil.ReadFile(filename)

		if err != nil {
			return err
		}
		localConfigVars := strings.Split(string(contents), "\n")

		configMap, err := parseConfig(localConfigVars[:len(localConfigVars)-1])
		if err != nil {
			return err
		}

		for key, value := range configVars.Values {
			localValue, ok := configMap[key]

			if ok {
				if value != localValue {
					var confirm string
					fmt.Fprintf(wOut, "%s: overwrite %s with %s? (y/N) ", key, localValue, value)

					fmt.Scanln(&confirm)

					if strings.ToLower(confirm) == "y" {
						configMap[key] = value
					}
				}
			} else {
				configMap[key] = value
			}
		}

		return ioutil.WriteFile(filename, []byte(formatConfig(configMap)), 0755)
	}

	return ioutil.WriteFile(filename, []byte(formatConfig(configVars.Values)), 0755)
}

// ConfigPush pushes an app's config from a file.
func ConfigPush(cf, appID, fileName string, wOut io.Writer) error {
	stat, err := os.Stdin.Stat()

	if err != nil {
		return err
	}

	var contents []byte

	if (stat.Mode() & os.ModeCharDevice) == 0 {
		buffer := new(bytes.Buffer)
		buffer.ReadFrom(os.Stdin)
		contents = buffer.Bytes()
	} else {
		contents, err = ioutil.ReadFile(fileName)

		if err != nil {
			return err
		}
	}

	file := strings.Split(string(contents), "\n")
	config := []string{}

	for _, configVar := range file {
		if len(configVar) > 0 {
			config = append(config, configVar)
		}
	}

	return ConfigSet(cf, appID, config, wOut)
}

func parseConfig(configVars []string) (map[string]interface{}, error) {
	configMap := make(map[string]interface{})

	regex := regexp.MustCompile(`^([A-z_]+[A-z0-9_]*)=([\s\S]+)$`)
	for _, config := range configVars {
		// Skip config that starts with an comment
		if config[0] == '#' {
			continue
		}

		if regex.MatchString(config) {
			captures := regex.FindStringSubmatch(config)
			configMap[captures[1]] = captures[2]
		} else {
			return nil, fmt.Errorf("'%s' does not match the pattern 'key=var', ex: MODE=test\n", config)
		}
	}

	return configMap, nil
}

// configPair is used for sorting configuration variables by removing them from the
// unsortible maps.
type configPair struct {
	Key   string
	Value interface{}
}

type configPairs []configPair

func (cPs configPairs) Len() int           { return len(cPs) }
func (cPs configPairs) Swap(i, j int)      { cPs[i], cPs[j] = cPs[j], cPs[i] }
func (cPs configPairs) Less(i, j int) bool { return cPs[i].Key < cPs[j].Key }

func sortedConfig(configVars map[string]interface{}) configPairs {
	var cPs configPairs

	for key, value := range configVars {
		cPs = append(cPs, configPair{Key: key, Value: value})
	}

	sort.Sort(cPs)
	return cPs
}

func formatConfig(configVars map[string]interface{}) string {
	var formattedConfig string

	cPs := sortedConfig(configVars)

	for _, cP := range cPs {
		formattedConfig += fmt.Sprintf("%s=%v\n", cP.Key, cP.Value)
	}

	return formattedConfig
}
