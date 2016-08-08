package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/deis/controller-sdk-go/api"
	"github.com/deis/controller-sdk-go/apps"
	"github.com/deis/controller-sdk-go/config"
	"github.com/deis/controller-sdk-go/domains"
	"github.com/deis/workflow-cli/pkg/git"
	"github.com/deis/workflow-cli/pkg/logging"
	"github.com/deis/workflow-cli/pkg/webbrowser"
	"github.com/deis/workflow-cli/settings"
)

// AppCreate creates an app.
func AppCreate(cf, id, buildpack, remote string, noRemote bool, wOut io.Writer) error {
	s, err := settings.Load(cf)
	if err != nil {
		return err
	}

	fmt.Fprint(wOut, "Creating Application... ")
	quit := progress(wOut)
	app, err := apps.New(s.Client, id)

	quit <- true
	<-quit

	if checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}

	fmt.Fprintf(wOut, "done, created %s\n", app.ID)

	if buildpack != "" {
		configValues := api.Config{
			Values: map[string]interface{}{
				"BUILDPACK_URL": buildpack,
			},
		}
		if _, err = config.Set(s.Client, app.ID, configValues); checkAPICompatibility(s.Client, err, wOut) != nil {
			return err
		}
	}

	if !noRemote {
		if err = git.CreateRemote(s.Client.ControllerURL.Host, remote, app.ID); err != nil {
			if err.Error() == "exit status 128" {
				msg := "A git remote with the name %s already exists. To overwrite this remote run:\n"
				msg += "deis git:remote --force --remote %s --app %s"
				return fmt.Errorf(msg, remote, remote, app.ID)
			}
			return err
		}

		fmt.Fprintf(wOut, remoteCreationMsg, remote, app.ID)
	}

	if noRemote {
		fmt.Fprintf(wOut, "If you want to add a git remote for this app later, use `deis git:remote -a %s`\n", app.ID)
	}

	return nil
}

// AppsList lists apps on the Deis controller.
func AppsList(cf string, results int, wOut io.Writer) error {
	s, err := settings.Load(cf)

	if err != nil {
		return err
	}

	if results == defaultLimit {
		results = s.Limit
	}

	apps, count, err := apps.List(s.Client, results)
	if checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}

	fmt.Fprintf(wOut, "=== Apps%s", limitCount(len(apps), count))

	for _, app := range apps {
		fmt.Fprintln(wOut, app.ID)
	}
	return nil
}

// AppInfo prints info about app.
func AppInfo(cf, appID string, wOut io.Writer) error {
	s, appID, err := load(cf, appID)

	if err != nil {
		return err
	}

	app, err := apps.Get(s.Client, appID)
	if checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}

	url, err := appURL(s, appID, wOut)
	if err != nil {
		return err
	}

	if url == "" {
		url = fmt.Sprintf(noDomainAssignedMsg, appID)
	}

	fmt.Fprintf(wOut, "=== %s Application\n", app.ID)
	fmt.Fprintln(wOut, "updated: ", app.Updated)
	fmt.Fprintln(wOut, "uuid:    ", app.UUID)
	fmt.Fprintln(wOut, "created: ", app.Created)
	fmt.Fprintln(wOut, "url:     ", url)
	fmt.Fprintln(wOut, "owner:   ", app.Owner)
	fmt.Fprintln(wOut, "id:      ", app.ID)

	fmt.Fprintln(wOut)
	// print the app processes
	if err = PsList(cf, app.ID, defaultLimit, wOut); err != nil {
		return err
	}

	fmt.Fprintln(wOut)
	// print the app domains
	if err = DomainsList(cf, app.ID, defaultLimit, wOut); err != nil {
		return err
	}

	fmt.Fprintln(wOut)

	return nil
}

// AppOpen opens an app in the default webbrowser.
func AppOpen(cf, appID string, wOut io.Writer) error {
	s, appID, err := load(cf, appID)

	if err != nil {
		return err
	}

	u, err := appURL(s, appID, wOut)
	if err != nil {
		return err
	}

	if u == "" {
		return fmt.Errorf(noDomainAssignedMsg, appID)
	}

	if !(strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://")) {
		u = "http://" + u
	}

	return webbrowser.Webbrowser(u)
}

// AppLogs returns the logs from an app.
func AppLogs(cf, appID string, lines int, wOut io.Writer) error {
	s, appID, err := load(cf, appID)

	if err != nil {
		return err
	}

	logs, err := apps.Logs(s.Client, appID, lines)
	if checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}

	for _, log := range strings.Split(strings.TrimRight(logs, `\n`), `\n`) {
		logging.PrintLog(os.Stdout, log)
	}

	return nil
}

// AppRun runs a one time command in the app.
func AppRun(cf, appID, command string, wOut io.Writer) error {
	s, appID, err := load(cf, appID)

	if err != nil {
		return err
	}

	fmt.Fprintf(wOut, "Running '%s'...\n", command)

	out, err := apps.Run(s.Client, appID, command)
	if checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}

	if out.ReturnCode == 0 {
		fmt.Fprint(wOut, out.Output)
	} else {
		fmt.Fprint(os.Stderr, out.Output)
	}

	os.Exit(out.ReturnCode)
	return nil
}

// AppDestroy destroys an app.
func AppDestroy(cf, appID, confirm string, wOut io.Writer) error {
	gitSession := false

	s, err := settings.Load(cf)

	if err != nil {
		return err
	}

	if appID == "" {
		appID, err = git.DetectAppName(s.Client.ControllerURL.Host)

		if err != nil {
			return err
		}

		gitSession = true
	}

	if confirm == "" {
		fmt.Fprintf(wOut, ` !    WARNING: Potentially Destructive Action
 !    This command will destroy the application: %s
 !    To proceed, type "%s" or re-run this command with --confirm=%s

> `, appID, appID, appID)

		fmt.Scanln(&confirm)
	}

	if confirm != appID {
		return fmt.Errorf("App %s does not match confirm %s, aborting.", appID, confirm)
	}

	startTime := time.Now()
	fmt.Fprintf(wOut, "Destroying %s...\n", appID)

	if err = apps.Delete(s.Client, appID); checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}

	fmt.Fprintf(wOut, "done in %ds\n", int(time.Since(startTime).Seconds()))

	if gitSession {
		return GitRemove(cf, appID, wOut)
	}

	return nil
}

// AppTransfer transfers app ownership to another user.
func AppTransfer(cf, appID, username string, wOut io.Writer) error {
	s, appID, err := load(cf, appID)

	if err != nil {
		return err
	}

	fmt.Fprintf(wOut, "Transferring %s to %s... ", appID, username)

	err = apps.Transfer(s.Client, appID, username)
	if checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}

	fmt.Fprintln(wOut, "done")

	return nil
}

const noDomainAssignedMsg = "No domain assigned to %s"

// appURL grabs the first domain an app has and returns this.
func appURL(s *settings.Settings, appID string, wOut io.Writer) (string, error) {
	domains, _, err := domains.List(s.Client, appID, 1)
	if checkAPICompatibility(s.Client, err, wOut) != nil {
		return "", err
	}

	if len(domains) == 0 {
		return "", nil
	}

	return expandURL(s.Client.ControllerURL.Host, domains[0].Domain), nil
}

// expandURL expands an app url if necessary.
func expandURL(host, u string) string {
	if strings.Contains(u, ".") {
		// If domain is a full url.
		return u
	}

	// If domain is a subdomain, look up the controller url and replace the subdomain.
	parts := strings.Split(host, ".")
	parts[0] = u
	return strings.Join(parts, ".")
}
