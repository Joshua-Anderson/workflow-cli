package cmd

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/arschles/assert"
	"github.com/deis/controller-sdk-go/api"
	"github.com/deis/workflow-cli/pkg/testutil"
)

func TestParseConfig(t *testing.T) {
	t.Parallel()

	_, err := parseConfig([]string{"FOO=bar", "CAR star"})
	assert.ExistsErr(t, err, "config")

	actual, err := parseConfig([]string{"FOO=bar"})
	assert.NoErr(t, err)
	assert.Equal(t, actual, map[string]interface{}{"FOO": "bar"}, "map")
}

// TODO(joshua-anderson): The controller should be saving everything as a string
// and the CLI should be sending everything as a string. This didn't use to be the
// case due to a bug in V1. Consider changing to map[string]string?
func TestFormatConfig(t *testing.T) {
	t.Parallel()

	testMap := map[string]interface{}{
		"TEST":  "testing",
		"NCC":   1701,
		"TRUE":  false,
		"FLOAT": 12.34,
	}

	testOut := formatConfig(testMap)
	assert.Equal(t, testOut, `FLOAT=12.34
NCC=1701
TEST=testing
TRUE=false
`, "output")
}

func TestConfigList(t *testing.T) {
	t.Parallel()
	cf, server, err := testutil.NewTestServerAndClient()
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()

	server.Mux.HandleFunc("/v2/apps/foo/config/", func(w http.ResponseWriter, r *http.Request) {
		testutil.SetHeaders(w)
		fmt.Fprintf(w, `{
    "owner": "jkirk",
    "app": "foo",
    "values": {
        "TEST":  "testing",
        "NCC":   "1701",
        "TRUE":  "false",
        "FLOAT": "12.34"
    },
    "memory": {},
    "cpu": {},
    "tags": {},
    "registry": {},
    "routable": true,
    "created": "2014-01-01T00:00:00UTC",
    "updated": "2014-01-01T00:00:00UTC",
    "uuid": "de1bf5b5-4a72-4f94-a10c-d2a3741cdf75"
}`)
	})

	var b bytes.Buffer

	err = ConfigList(cf, "foo", false, &b)
	assert.NoErr(t, err)

	assert.Equal(t, b.String(), `=== foo Config
FLOAT      12.34
NCC        1701
TEST       testing
TRUE       false
`, "output")
	b.Reset()

	err = ConfigList(cf, "foo", true, &b)
	assert.NoErr(t, err)

	assert.Equal(t, b.String(), "FLOAT=12.34 NCC=1701 TEST=testing TRUE=false\n", "output")
}

func TestConfigSet(t *testing.T) {
	t.Parallel()
	cf, server, err := testutil.NewTestServerAndClient()
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()

	server.Mux.HandleFunc("/v2/apps/foo/config/", func(w http.ResponseWriter, r *http.Request) {
		testutil.SetHeaders(w)
		if r.Method == "POST" {
			testutil.AssertBody(t, api.Config{
				Values: map[string]interface{}{
					"TRUE": "false",
				},
			}, r)
		}

		fmt.Fprintf(w, `{
	"owner": "jkirk",
	"app": "foo",
	"values": {
			"TEST":  "testing",
			"NCC":   "1701",
			"TRUE":  "false",
			"FLOAT": "12.34"
	},
	"memory": {},
	"cpu": {},
	"tags": {},
	"registry": {},
	"routable": true,
	"created": "2014-01-01T00:00:00UTC",
	"updated": "2014-01-01T00:00:00UTC",
	"uuid": "de1bf5b5-4a72-4f94-a10c-d2a3741cdf75"
}`)
	})

	var b bytes.Buffer

	err = ConfigSet(cf, "foo", []string{"TRUE=false"}, &b)
	assert.NoErr(t, err)

	assert.Equal(t, testutil.StripProgress(b.String()), `Creating config... done

=== foo Config
FLOAT      12.34
NCC        1701
TEST       testing
TRUE       false
`, "output")
}
