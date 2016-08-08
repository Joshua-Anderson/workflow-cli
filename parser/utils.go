package parser

import (
	"fmt"
	"io"
	"log"
	"strconv"
)

func safeGetValue(args map[string]interface{}, key string) string {
	if args[key] == nil {
		return ""
	}
	return args[key].(string)
}

func safeGetInt(args map[string]interface{}, key string) int {
	if args[key] == nil {
		return 0
	}
	retVal, err := strconv.Atoi(args[key].(string))
	if err != nil {
		log.Fatalf("could not convert %s to int: %v", args[key], err)
	}
	return retVal
}

func responseLimit(limit string) (int, error) {
	if limit == "" {
		return -1, nil
	}

	return strconv.Atoi(limit)
}

// PrintUsage runs if no matching command is found.
func PrintUsage(wErr io.Writer) {
	fmt.Fprintln(wErr, "Found no matching command, try 'deis help'")
	fmt.Fprintln(wErr, "Usage: deis <command> [<args>...]")
}

func printHelp(argv []string, usage string, wOut io.Writer) bool {
	if len(argv) > 1 {
		if argv[1] == "--help" || argv[1] == "-h" {
			fmt.Fprint(wOut, usage)
			return true
		}
	}

	return false
}

func addGlobalFlags(input string) string {
	return input + `  -c --config=<config>
    path to configuration file. Equilivent to
    setting $DEIS_PROFILE. Defaults to ~/.deis/config.json.
    If value is not a filepath, will assume location ~/.deis/<value>.json`
}
