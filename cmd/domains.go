package cmd

import (
	"fmt"
	"io"

	"github.com/deis/controller-sdk-go/domains"
)

// DomainsList lists domains registered with an app.
func DomainsList(cf, appID string, results int, wOut io.Writer) error {
	s, appID, err := load(cf, appID)

	if err != nil {
		return err
	}

	if results == defaultLimit {
		results = s.Limit
	}

	domains, count, err := domains.List(s.Client, appID, results)
	if checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}

	fmt.Fprintf(wOut, "=== %s Domains%s", appID, limitCount(len(domains), count))

	for _, domain := range domains {
		fmt.Fprintln(wOut, domain.Domain)
	}
	return nil
}

// DomainsAdd adds a domain to an app.
func DomainsAdd(cf, appID, domain string, wOut io.Writer) error {
	s, appID, err := load(cf, appID)

	if err != nil {
		return err
	}

	fmt.Fprintf(wOut, "Adding %s to %s... ", domain, appID)

	quit := progress(wOut)
	_, err = domains.New(s.Client, appID, domain)
	quit <- true
	<-quit
	if checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}

	fmt.Fprintln(wOut, "done")
	return nil
}

// DomainsRemove removes a domain registered with an app.
func DomainsRemove(cf, appID, domain string, wOut io.Writer) error {
	s, appID, err := load(cf, appID)

	if err != nil {
		return err
	}

	fmt.Fprintf(wOut, "Removing %s from %s... ", domain, appID)

	quit := progress(wOut)
	err = domains.Delete(s.Client, appID, domain)
	quit <- true
	<-quit
	if checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}

	fmt.Fprintln(wOut, "done")
	return nil
}
