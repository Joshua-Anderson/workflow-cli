package cmd

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/deis/controller-sdk-go"
	"github.com/deis/controller-sdk-go/api"
	"github.com/deis/controller-sdk-go/ps"
)

// PsList lists an app's processes.
func PsList(cf, appID string, results int, wOut io.Writer) error {
	s, appID, err := load(cf, appID)
	if err != nil {
		return err
	}

	if results == defaultLimit {
		results = s.Limit
	}

	processes, _, err := ps.List(s.Client, appID, results)
	if checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}

	printProcesses(appID, processes, wOut)

	return nil
}

// PsScale scales an app's processes.
func PsScale(cf, appID string, targets []string, wOut io.Writer) error {
	s, appID, err := load(cf, appID)

	if err != nil {
		return err
	}

	targetMap := make(map[string]int)
	regex := regexp.MustCompile("^([a-z0-9]+)=([0-9]+)$")

	for _, target := range targets {
		if regex.MatchString(target) {
			captures := regex.FindStringSubmatch(target)
			targetMap[captures[1]], err = strconv.Atoi(captures[2])

			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("'%s' does not match the pattern 'type=num', ex: web=2\n", target)
		}
	}

	fmt.Fprintf(wOut, "Scaling processes... but first, %s!\n", drinkOfChoice())
	startTime := time.Now()
	quit := progress(wOut)

	err = ps.Scale(s.Client, appID, targetMap)
	quit <- true
	<-quit
	if checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}

	fmt.Fprintf(wOut, "done in %ds\n", int(time.Since(startTime).Seconds()))

	processes, _, err := ps.List(s.Client, appID, s.Limit)
	if err != nil {
		return err
	}

	printProcesses(appID, processes, wOut)
	return nil
}

// PsRestart restarts an app's processes.
func PsRestart(cf, appID, target string, wOut io.Writer) error {
	s, appID, err := load(cf, appID)

	if err != nil {
		return err
	}

	psType, psName := "", ""
	if target != "" {
		psType, psName = parseType(target, appID)
	}

	fmt.Fprintf(wOut, "Restarting processes... but first, %s!\n", drinkOfChoice())
	startTime := time.Now()
	quit := progress(wOut)

	processes, err := ps.Restart(s.Client, appID, psType, psName)
	quit <- true
	<-quit
	if err == deis.ErrPodNotFound {
		return fmt.Errorf("Could not find proccess type %s in app %s", psType, appID)
	} else if checkAPICompatibility(s.Client, err, wOut) != nil {
		return err
	}

	if len(processes) == 0 {
		fmt.Fprintln(wOut, "Could not find any processes to restart")
	} else {
		fmt.Fprintf(wOut, "done in %ds\n", int(time.Since(startTime).Seconds()))
		printProcesses(appID, processes, wOut)
	}

	return nil
}

func printProcesses(appID string, processes []api.Pods, wOut io.Writer) {
	psMap := ps.ByType(processes)

	fmt.Fprintf(wOut, "=== %s Processes\n", appID)

	for psType, procs := range psMap {
		fmt.Fprintf(wOut, "--- %s:\n", psType)

		for _, proc := range procs {
			fmt.Fprintf(wOut, "%s %s (%s)\n", proc.Name, proc.State, proc.Release)
		}
	}
}

func parseType(target string, appID string) (string, string) {
	psType, psName := "", ""

	if strings.Contains(target, "-") {
		replaced := strings.Replace(target, appID+"-", "", 1)
		parts := strings.Split(replaced, "-")
		// the API requires the type, for now
		// regex matches against how Deployment pod name is constructed
		regex := regexp.MustCompile("[0-9]{8,10}-[a-z0-9]{5}$")
		if regex.MatchString(replaced) {
			psType = parts[0]
		} else {
			psType = parts[1]
		}
		// process name is the full pod
		psName = target
	} else {
		psType = target
	}

	return psType, psName
}
