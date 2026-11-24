package client

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/dns3l/dns3l-go/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type closeTrackingBody struct {
	io.Reader
	closed bool
}

func (b *closeTrackingBody) Close() error {
	b.closed = true
	return nil
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestInternalResponsesAreClosed(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr bool
		call    func(Client) error
	}{
		{
			name: "JSON response",
			body: `{}`,
			call: func(c Client) error {
				_, err := c.GetStatus(context.Background())
				return err
			},
		},
		{
			name: "all PEM resources",
			body: `{}`,
			call: func(c Client) error {
				_, err := c.GetAllResources(context.Background(), "ca1", "cert")
				return err
			},
		},
		{
			name: "certificate bytes",
			body: "certificate PEM",
			call: func(c Client) error {
				_, err := c.GetResourceBytes(context.Background(), "ca1", "cert", ResourceTypeCert)
				return err
			},
		},
		{
			name:    "certificate parser failure",
			body:    "invalid certificate PEM",
			wantErr: true,
			call: func(c Client) error {
				_, err := c.GetResourceCert(context.Background(), "ca1", "cert")
				return err
			},
		},
		{
			name:    "key parser failure",
			body:    "invalid key PEM",
			wantErr: true,
			call: func(c Client) error {
				_, err := c.GetResourceKey(context.Background(), "ca1", "cert")
				return err
			},
		},
		{
			name: "chain response",
			body: "",
			call: func(c Client) error {
				_, err := c.GetResourceChain(context.Background(), "ca1", "cert")
				return err
			},
		},
		{
			name: "claim response",
			body: "ignored",
			call: func(c Client) error {
				return c.ClaimCertificate(context.Background(), "ca1", &api.CertClaimInfo{Name: "cert.example.com"})
			},
		},
		{
			name: "delete response",
			body: "ignored",
			call: func(c Client) error {
				return c.DeleteCertByCA(context.Background(), "ca1", "cert")
			},
		},
		{
			name: "delete all response",
			body: "ignored",
			call: func(c Client) error {
				return c.DeleteCertAllCAs(context.Background(), "cert")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := &closeTrackingBody{Reader: strings.NewReader(tt.body)}
			baseURL, err := url.Parse("https://dns3l.example")
			require.NoError(t, err)
			client := NewUnauthenticated(baseURL, &http.Client{
				Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode:    http.StatusOK,
						Header:        make(http.Header),
						Body:          body,
						ContentLength: -1,
						Request:       r,
					}, nil
				}),
			})

			err = tt.call(client)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.True(t, body.closed, "internal response body should be closed")
		})
	}
}

func TestRawResponseBodyIsCallerClosable(t *testing.T) {
	body := &closeTrackingBody{Reader: strings.NewReader(`{"version":{}}`)}
	baseURL, err := url.Parse("https://dns3l.example")
	require.NoError(t, err)
	client := NewUnauthenticated(baseURL, &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode:    http.StatusOK,
				Header:        make(http.Header),
				Body:          body,
				ContentLength: -1,
				Request:       r,
			}, nil
		}),
	})

	reader, err := client.Raw().GetServerInfo(context.Background())
	require.NoError(t, err)
	assert.False(t, body.closed, "raw response should remain open for its caller")
	require.NoError(t, reader.Close())
	assert.True(t, body.closed)
}
