package cmd

import (
	"fmt"
	"io"

	"github.com/deis/controller-sdk-go/api"
	"github.com/deis/controller-sdk-go/config"
)

// HealthchecksList lists an app's healthchecks.
func HealthchecksList(cf, appID string, wOut io.Writer) error {
	s, appID, err := load(cf, appID)

	if err != nil {
		return err
	}

	config, err := config.List(s.Client, appID)

	if err != nil {
		return err
	}

	fmt.Fprintf(wOut, "=== %s Healthchecks\n\n", appID)

	fmt.Fprintln(wOut, "--- Liveness")
	if livenessProbe, found := config.Healthcheck["livenessProbe"]; found {
		fmt.Fprintln(wOut, livenessProbe)
	} else {
		fmt.Fprintln(wOut, "No liveness probe configured.")
	}

	fmt.Fprintln(wOut, "\n--- Readiness")
	if readinessProbe, found := config.Healthcheck["readinessProbe"]; found {
		fmt.Fprintln(wOut, readinessProbe)
	} else {
		fmt.Fprintln(wOut, "No readiness probe configured.")
	}
	return nil
}

// HealthchecksSet sets an app's healthchecks.
func HealthchecksSet(cf, appID, healthcheckType string, probe *api.Healthcheck, wOut io.Writer) error {
	s, appID, err := load(cf, appID)

	if err != nil {
		return err
	}

	fmt.Fprintf(wOut, "Applying %s healthcheck... ", healthcheckType)

	quit := progress(wOut)
	configObj := api.Config{}
	configObj.Healthcheck = make(map[string]*api.Healthcheck)

	configObj.Healthcheck[healthcheckType] = probe

	_, err = config.Set(s.Client, appID, configObj)

	quit <- true
	<-quit

	if err != nil {
		return err
	}

	fmt.Fprint(wOut, "done\n\n")

	return HealthchecksList(cf, appID, wOut)
}

// HealthchecksUnset removes an app's healthchecks.
func HealthchecksUnset(cf, appID string, healthchecks []string, wOut io.Writer) error {
	s, appID, err := load(cf, appID)

	if err != nil {
		return err
	}

	fmt.Fprint(wOut, "Removing healthchecks... ")

	quit := progress(wOut)

	configObj := api.Config{}

	healthcheckMap := make(map[string]*api.Healthcheck)

	for _, healthcheck := range healthchecks {
		healthcheckMap[healthcheck] = nil
	}

	configObj.Healthcheck = healthcheckMap

	_, err = config.Set(s.Client, appID, configObj)

	quit <- true
	<-quit

	if err != nil {
		return err
	}

	fmt.Fprint(wOut, "done\n\n")

	return HealthchecksList(cf, appID, wOut)
}
