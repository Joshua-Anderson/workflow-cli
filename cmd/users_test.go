package cmd

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/arschles/assert"
	"github.com/deis/workflow-cli/pkg/testutil"
)

func TestUsersList(t *testing.T) {
	t.Parallel()
	cf, server, err := testutil.NewTestServerAndClient()
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()

	server.Mux.HandleFunc("/v2/users/", func(w http.ResponseWriter, r *http.Request) {
		testutil.SetHeaders(w)
		fmt.Fprintf(w, `{
    "count": 2,
    "next": null,
    "previous": null,
    "results": [
        {
            "id": 1,
            "last_login": "2014-10-19T22:01:00.601Z",
            "is_superuser": true,
            "username": "test",
            "first_name": "test",
            "last_name": "testerson",
            "email": "test@example.com",
            "is_staff": true,
            "is_active": true,
            "date_joined": "2014-10-19T22:01:00.601Z",
            "groups": [],
            "user_permissions": []
        },
        {
            "id": 2,
            "last_login": "2014-10-19T22:01:00.601Z",
            "is_superuser": false,
            "username": "jkirk",
            "first_name": "james",
            "last_name": "kirk",
            "email": "jkirk@starfleet.com",
            "is_staff": true,
            "is_active": true,
            "date_joined": "2014-10-19T22:01:00.601Z",
            "groups": [],
            "user_permissions": []
        }
    ]
}`)
	})

	var b bytes.Buffer

	err = UsersList(cf, -1, &b)
	assert.NoErr(t, err)

	assert.Equal(t, b.String(), `=== Users
test
jkirk
`, "output")
}
