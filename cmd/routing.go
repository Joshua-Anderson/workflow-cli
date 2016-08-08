package cmd

import (
	"fmt"
	"io"

	"github.com/deis/controller-sdk-go/api"
	"github.com/deis/controller-sdk-go/config"
)

// RoutingInfo provides information about the status of app routing.
func RoutingInfo(cf, appID string, wOut io.Writer) error {
	s, appID, err := load(cf, appID)

	if err != nil {
		return err
	}

	config, err := config.List(s.Client, appID)
	if checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}

	if config.Routable {
		fmt.Fprintln(wOut, "Routing is enabled.")
	} else {
		fmt.Fprintln(wOut, "Routing is disabled.")
	}
	return nil
}

// RoutingEnable enables an app from being exposed by the router.
func RoutingEnable(cf, appID string, wOut io.Writer) error {
	s, appID, err := load(cf, appID)

	if err != nil {
		return err
	}

	fmt.Fprintf(wOut, "Enabling routing for %s... ", appID)

	quit := progress(wOut)
	_, err = config.Set(s.Client, appID, api.Config{Routable: true})

	quit <- true
	<-quit

	if err != nil {
		return err
	}

	fmt.Fprint(wOut, "done\n\n")
	return nil
}

// RoutingDisable disables an app from being exposed by the router.
func RoutingDisable(cf, appID string, wOut io.Writer) error {
	s, appID, err := load(cf, appID)

	if err != nil {
		return err
	}

	fmt.Fprintf(wOut, "Disabling routing for %s... ", appID)

	quit := progress(wOut)
	_, err = config.Set(s.Client, appID, api.Config{Routable: false})

	quit <- true
	<-quit

	if err != nil {
		return err
	}

	fmt.Fprint(wOut, "done\n\n")
	return nil
}
