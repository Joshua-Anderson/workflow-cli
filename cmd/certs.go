package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"

	"github.com/deis/controller-sdk-go"
	"github.com/deis/controller-sdk-go/certs"
	"github.com/deis/workflow-cli/settings"
)

// CertsList lists certs registered with the controller.
func CertsList(cf string, results int, wOut io.Writer) error {
	s, err := settings.Load(cf)

	if err != nil {
		return err
	}

	if results == defaultLimit {
		results = s.Limit
	}

	certList, _, err := certs.List(s.Client, results)
	if checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}

	if len(certList) == 0 {
		fmt.Fprintln(wOut, "No certs")
		return nil
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(false)
	table.SetAutoFormatHeaders(false)
	table.SetHeaderLine(true)
	table.SetHeader([]string{"Name", "Common Name", "SubjectAltName", "Expires", "Fingerprint", "Domains", "Updated", "Created"})
	for _, cert := range certList {
		domains := strings.Join(cert.Domains[:], ",")
		san := strings.Join(cert.SubjectAltName[:], ",")

		// Make dates more readable
		now := time.Now()
		expires := cert.Expires.Time.Format("2 Jan 2006")
		created := cert.Created.Time.Format("2 Jan 2006")
		updated := cert.Updated.Time.Format("2 Jan 2006")

		if cert.Expires.Time.Before(now) {
			expires += " (expired)"
		} else {
			// Ghetto solution
			expires += " (in"
			year := cert.Expires.Time.Year() - now.Year()
			month := cert.Expires.Time.Month() - now.Month()
			day := cert.Expires.Time.Day() - now.Day()

			if year > 0 {
				expires += fmt.Sprintf(" %d year", year)
				if year > 1 {
					expires += "s"
				}
			} else if month > 0 {
				expires += fmt.Sprintf(" %d month", month)
				if month > 1 {
					expires += "s"
				}
			} else if day != 0 {
				// special handling on negative days
				if day < 0 {
					day *= -1
				}

				expires += fmt.Sprintf(" %d day", day)
				if day > 1 {
					expires += "s"
				}
			}
			expires += ")"
		}

		// show a shorter version of the fingerprint
		fingerprint := cert.Fingerprint[:5] + "[...]" + cert.Fingerprint[len(cert.Fingerprint)-5:]

		table.Append([]string{cert.Name, cert.CommonName, san, expires, fingerprint, domains, updated, created})
	}
	table.Render()

	return nil
}

// CertAdd adds a cert to the controller.
func CertAdd(cf string, cert string, key string, name string, wOut io.Writer) error {
	s, err := settings.Load(cf)

	if err != nil {
		return err
	}

	fmt.Fprint(wOut, "Adding SSL endpoint... ")
	quit := progress(wOut)
	err = doCertAdd(s.Client, cert, key, name, wOut)
	quit <- true
	<-quit

	if err != nil {
		return err
	}

	fmt.Fprintln(wOut, "done")
	return nil
}

func doCertAdd(c *deis.Client, cert string, key string, name string, wOut io.Writer) error {
	certFile, err := ioutil.ReadFile(cert)
	if err != nil {
		return err
	}

	keyFile, err := ioutil.ReadFile(key)
	if err != nil {
		return err
	}

	_, err = certs.New(c, string(certFile), string(keyFile), name)
	return checkAPICompatibility(c, err, wOut)
}

// CertRemove deletes a cert from the controller.
func CertRemove(cf, name string, wOut io.Writer) error {
	s, err := settings.Load(cf)
	if checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}

	fmt.Fprintf(wOut, "Removing %s... ", name)
	quit := progress(wOut)

	err = certs.Delete(s.Client, name)
	quit <- true
	<-quit
	if checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}

	fmt.Fprintln(wOut, "done")
	return nil
}

// CertInfo gets info about certficiate
func CertInfo(cf, name string, wOut io.Writer) error {
	s, err := settings.Load(cf)
	if err != nil {
		return err
	}

	cert, err := certs.Get(s.Client, name)
	if checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}

	domains := strings.Join(cert.Domains[:], ",")
	if domains == "" {
		domains = "No connected domains"
	}

	san := strings.Join(cert.SubjectAltName[:], ",")
	if san == "" {
		san = "N/A"
	}

	fmt.Fprintf(wOut, "=== %s Certificate\n", cert.Name)
	fmt.Fprintln(wOut, "Common Name(s):    ", cert.CommonName)
	fmt.Fprintln(wOut, "Expires At:        ", cert.Expires)
	fmt.Fprintln(wOut, "Starts At:         ", cert.Starts)
	fmt.Fprintln(wOut, "Fingerprint:       ", cert.Fingerprint)
	fmt.Fprintln(wOut, "Subject Alt Name:  ", san)
	fmt.Fprintln(wOut, "Issuer:            ", cert.Issuer)
	fmt.Fprintln(wOut, "Subject:           ", cert.Subject)
	fmt.Fprintln(wOut)
	fmt.Fprintln(wOut, "Connected Domains: ", domains)
	fmt.Fprintln(wOut, "Owner:             ", cert.Owner)
	fmt.Fprintln(wOut, "Created:           ", cert.Created)
	fmt.Fprintln(wOut, "Updated:           ", cert.Updated)

	return nil
}

// CertAttach attaches a certificate to a domain
func CertAttach(cf, name, domain string, wOut io.Writer) error {
	s, err := settings.Load(cf)

	if err != nil {
		return err
	}

	fmt.Fprintf(wOut, "Attaching certificate %s to domain %s... ", name, domain)
	quit := progress(wOut)

	err = certs.Attach(s.Client, name, domain)
	quit <- true
	<-quit
	if checkAPICompatibility(s.Client, err, wOut) == nil {
		fmt.Fprintln(wOut, "done")
	}

	return err
}

// CertDetach detaches a certificate from a domain
func CertDetach(cf, name, domain string, wOut io.Writer) error {
	s, err := settings.Load(cf)

	if err != nil {
		return err
	}

	fmt.Fprintf(wOut, "Detaching certificate %s from domain %s... ", name, domain)
	quit := progress(wOut)

	err = certs.Detach(s.Client, name, domain)
	quit <- true
	<-quit
	if checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}

	fmt.Fprintln(wOut, "done")
	return nil
}
