package client

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/dns3l/dns3l-go/api"
	"github.com/stretchr/testify/assert"
)

func TestGetServerInfo(t *testing.T) {
	time := time.Date(2021, 06, 03, 12, 00, 00, 0, time.UTC)
	srv := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/info", r.URL.EscapedPath())
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			assert.NoError(t, asJSON(&api.ServerInfo{
				Version: &api.ServerInfoVersion{
					Daemon: "v1.2.3",
					API:    "v4.5.6",
				},
				Contact: &api.ServerInfoContact{
					URL: "https://dns3l.example.com",
					EMail: []string{
						"kilgore.trout@example.com",
					},
				},
				Auth: &api.ServerInfoAuth{
					GroupsPrefix:      "Dom2_Dns3l_",
					GroupsDomainDelim: ".",
				},
				Renewal: &api.ServerInfoRenewal{
					LastRun:    &time,
					Successful: 12,
					Failed:     0,
				},
			}, w))
		}))
	defer srv.Close()
	srvurl, _ := url.Parse(srv.URL)
	client := NewUnauthenticated(srvurl, srv.Client())
	srvinfo, err := client.GetServerInfo(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, srvinfo)
}
