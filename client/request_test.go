package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"github.com/dns3l/dns3l-go/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthenticationAndUserAgentHeaders(t *testing.T) {
	tests := []struct {
		name        string
		authHeader  string
		authValue   string
		apiKey      string
		customAgent string
	}{
		{name: "unauthenticated"},
		{name: "oauth token", authHeader: "Authorization", authValue: "Bearer token-value"},
		{name: "API key", authHeader: "X-DNS3L-API-Key", authValue: "key-value"},
		{name: "custom user agent", customAgent: "test-client/1.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.authHeader == "" {
					assert.Empty(t, r.Header.Get("Authorization"))
					assert.Empty(t, r.Header.Get("X-DNS3L-API-Key"))
				} else {
					assert.Equal(t, tt.authValue, r.Header.Get(tt.authHeader))
				}
				if tt.customAgent != "" {
					assert.Equal(t, tt.customAgent, r.Header.Get("User-Agent"))
				} else {
					assert.Contains(t, r.Header.Get("User-Agent"), "dns3l-go/")
				}
				w.WriteHeader(http.StatusNoContent)
			}))

			baseURL, err := url.Parse(server.URL)
			require.NoError(t, err)
			var authenticatedClient Client
			switch tt.authHeader {
			case "Authorization":
				authenticatedClient = NewByOauth2AccessToken(baseURL, nil, "token-value")
			case "X-DNS3L-API-Key":
				authenticatedClient = NewByApiKey(baseURL, nil, "key-value")
			case "":
				authenticatedClient = NewUnauthenticated(baseURL, nil)
			default:
				authenticatedClient = client
			}
			if tt.customAgent != "" {
				authenticatedClient.Raw().SetUserAgent(tt.customAgent)
			}
			reader, err := authenticatedClient.Raw().GetServerInfo(context.Background())
			require.NoError(t, err)
			require.NoError(t, reader.Close())
		})
	}
}

func TestClaimAndDeleteRequests(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		call   func(Client) error
	}{
		{
			name:   "claim certificate",
			method: http.MethodPost,
			path:   "/ca/ca1/crt",
			call: func(c Client) error {
				return c.ClaimCertificate(context.Background(), "ca1", &api.CertClaimInfo{
					Name:            "www.example.com",
					Wildcard:        true,
					SubjectAltNames: []string{"www.example.com", "api.example.com"},
					Hints:           api.CertClaimHints{TTL: 14},
				})
			},
		},
		{
			name:   "delete certificate by CA",
			method: http.MethodDelete,
			path:   "/ca/ca1/crt/www.example.com",
			call: func(c Client) error {
				return c.DeleteCertByCA(context.Background(), "ca1", "www.example.com")
			},
		},
		{
			name:   "delete certificate across CAs",
			method: http.MethodDelete,
			path:   "/crt/www.example.com",
			call: func(c Client) error {
				return c.DeleteCertAllCAs(context.Background(), "www.example.com")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.method, r.Method)
				assert.Equal(t, tt.path, r.URL.EscapedPath())
				if tt.name == "claim certificate" {
					assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
					var got api.CertClaimInfo
					assert.NoError(t, json.NewDecoder(r.Body).Decode(&got))
					assert.Equal(t, "www.example.com", got.Name)
					assert.True(t, got.Wildcard)
					assert.Equal(t, []string{"www.example.com", "api.example.com"}, got.SubjectAltNames)
					assert.Equal(t, uint16(14), got.Hints.TTL)
				}
				w.WriteHeader(http.StatusNoContent)
			}))

			require.NoError(t, tt.call(client))
		})
	}
}

func TestGetCertsByCAQueryOptions(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/ca/ca1/crt", r.URL.EscapedPath())
		assert.Equal(t, "4", r.URL.Query().Get("limit"))
		assert.Equal(t, "5", r.URL.Query().Get("offset"))
		assert.Equal(t, "api server", r.URL.Query().Get("search"))
		assert.Equal(t, "example.com", r.URL.Query().Get("suffix"))
		w.Header().Set("Page-Limit", "4")
		w.Header().Set("Page-Offset", "5")
		w.Header().Set("Total-Count", "10")
		_, _ = w.Write([]byte(`[]`))
	}))

	reader, page, err := client.Raw().GetCertsByCA(context.Background(), "ca1", &CertQueryOptions{
		Limit: 4, Offset: 5, Search: "api server", Suffix: "example.com",
	})
	require.NoError(t, err)
	require.NotNil(t, page)
	require.NotNil(t, reader)
	require.NoError(t, reader.Close())
}

func TestRawGetCerts(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/crt", r.URL.EscapedPath())
		assert.Equal(t, "2", r.URL.Query().Get("limit"))
		_, _ = w.Write([]byte(`[]`))
	}))

	reader, page, err := client.Raw().GetCerts(context.Background(), &CertQueryOptions{Limit: 2})
	require.NoError(t, err)
	require.NotNil(t, reader)
	require.NotNil(t, page)
	require.NoError(t, reader.Close())
}

func TestHiddenPagingAggregatesCertificatePages(t *testing.T) {
	var calls int
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/crt", r.URL.EscapedPath())
		assert.Equal(t, "example.com", r.URL.Query().Get("suffix"))
		assert.Equal(t, "valid", r.URL.Query().Get("search"))
		switch calls {
		case 0:
			assert.Equal(t, "2", r.URL.Query().Get("limit"))
			assert.Empty(t, r.URL.Query().Get("offset"))
			w.Header().Set("Page-Limit", "2")
			w.Header().Set("Page-Offset", "0")
			w.Header().Set("Total-Count", "3")
			_, _ = w.Write([]byte(`[{"name":"a.example.com","ca":"ca1"},{"name":"b.example.com","ca":"ca1"}]`))
		case 1:
			assert.Equal(t, "1", r.URL.Query().Get("limit"))
			assert.Equal(t, "2", r.URL.Query().Get("offset"))
			w.Header().Set("Page-Limit", "1")
			w.Header().Set("Page-Offset", "2")
			w.Header().Set("Total-Count", "3")
			_, _ = w.Write([]byte(`[{"name":"c.example.com","ca":"ca2"}]`))
		default:
			assert.Fail(t, "unexpected extra page request")
		}
		calls++
	}))
	client.SetHiddenPaging(2)

	certs, page, err := client.GetCerts(context.Background(), &CertQueryOptions{
		Limit: 3, Search: "valid", Suffix: "example.com",
	})
	require.NoError(t, err)
	require.Len(t, certs, 3)
	require.NotNil(t, page)
	assert.Equal(t, []string{"a.example.com", "b.example.com", "c.example.com"},
		[]string{certs[0].Name, certs[1].Name, certs[2].Name})
	assert.Equal(t, 2, calls)
}

func TestHTTPAndDecodeErrors(t *testing.T) {
	tests := []struct {
		name      string
		status    int
		body      string
		wantErr   string
		typedCall bool
	}{
		{
			name:    "API error response",
			status:  http.StatusForbidden,
			body:    `{"code":403,"message":"not authorized"}`,
			wantErr: "server sent: not authorized, code=403",
		},
		{
			name:    "invalid error response",
			status:  http.StatusBadGateway,
			body:    "not JSON",
			wantErr: "could not decode server error msg",
		},
		{
			name:      "invalid successful JSON",
			status:    http.StatusOK,
			body:      "{",
			wantErr:   "unexpected EOF",
			typedCall: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))

			if tt.typedCall {
				_, err := client.GetServerInfo(context.Background())
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			_, err := client.Raw().GetServerInfo(context.Background())
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestHTTPStatusSentinelErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    error
	}{
		{name: "bad request", statusCode: http.StatusBadRequest, wantErr: ErrBadRequest},
		{name: "not found", statusCode: http.StatusNotFound, wantErr: ErrNotFound},
		{name: "conflict", statusCode: http.StatusConflict, wantErr: ErrConflict},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const serverMessage = "exact server response text"
			client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(`{"code":` + strconv.Itoa(tt.statusCode) + `,"message":"` + serverMessage + `"}`))
			}))

			_, err := client.Raw().GetServerInfo(context.Background())
			require.Error(t, err)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Contains(t, err.Error(), serverMessage)
		})
	}
}

func TestRequestContextCancellation(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := client.Raw().GetServerInfo(ctx)
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
}
