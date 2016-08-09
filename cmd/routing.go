package cmd

import (
	"fmt"

	"github.com/deis/controller-sdk-go/api"
	"github.com/deis/controller-sdk-go/config"
)

// RoutingInfo provides information about the status of app routing.
func RoutingInfo(cf, appID string) error {
	s, appID, err := load(cf, appID)

	if err != nil {
		return err
	}

	config, err := config.List(s.Client, appID)
	if checkAPICompatibility(s.Client, err) != nil {
		return err
	}

	if config.Routable {
		fmt.Println("Routing is enabled.")
	} else {
		fmt.Println("Routing is disabled.")
	}
	return nil
}

// RoutingEnable enables an app from being exposed by the router.
func RoutingEnable(cf, appID string) error {
	s, appID, err := load(cf, appID)

	if err != nil {
		return err
	}

	fmt.Printf("Enabling routing for %s... ", appID)

	quit := progress()
	_, err = config.Set(s.Client, appID, api.Config{Routable: true})

	quit <- true
	<-quit

	if err != nil {
		return err
	}

	fmt.Print("done\n\n")
	return nil
}

// RoutingDisable disables an app from being exposed by the router.
func RoutingDisable(cf, appID string) error {
	s, appID, err := load(cf, appID)

	if err != nil {
		return err
	}

	fmt.Printf("Disabling routing for %s... ", appID)

	quit := progress()
	_, err = config.Set(s.Client, appID, api.Config{Routable: false})

	quit <- true
	<-quit

	if err != nil {
		return err
	}

	fmt.Print("done\n\n")
	return nil
}
