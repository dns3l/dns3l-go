package client

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientGETMethods(t *testing.T) {
	const cert = `{"name":"www.example.com","claimedBy":{"name":"Alice","email":"alice@example.com"},"valid":true,"renewCount":3}`
	tests := []struct {
		name string
		path string
		body string
		call func(*testing.T, Client)
	}{
		{
			name: "server info",
			path: "/info",
			body: `{"version":{"daemon":"v1","api":"v2"},"contact":{"url":"https://example.com","email":["ops@example.com"]},"auth":{"mode":"oidc"},"renewal":{"successful":4}}`,
			call: func(t *testing.T, c Client) {
				got, err := c.GetServerInfo(context.Background())
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, "v1", got.Version.Daemon)
				assert.Equal(t, []string{"ops@example.com"}, got.Contact.EMail)
				assert.Equal(t, uint(4), got.Renewal.Successful)
			},
		},
		{
			name: "status",
			path: "/status",
			body: `{"dns":[{"id":"provider","healthy":true}],"cached":true,"age":12}`,
			call: func(t *testing.T, c Client) {
				got, err := c.GetStatus(context.Background())
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.True(t, got.Cached)
				assert.Equal(t, int32(12), got.Age)
				assert.Equal(t, "provider", got.DNS[0].ID)
			},
		},
		{
			name: "dns handlers",
			path: "/dns",
			body: `[{"id":"rfc2136","name":"RFC2136","feature":["dns-01"],"zoneNesting":true}]`,
			call: func(t *testing.T, c Client) {
				got, err := c.GetDNSHandlers(context.Background())
				require.NoError(t, err)
				require.Len(t, got, 1)
				assert.Equal(t, "rfc2136", got[0].ID)
				assert.True(t, got[0].ZoneNesting)
			},
		},
		{
			name: "dns rootzones",
			path: "/dns/rtzn",
			body: `[{"root":"example.com","autodns":"rfc2136","acmedns":"rfc2136","ca":["ca1"]}]`,
			call: func(t *testing.T, c Client) {
				got, err := c.GetDNSRootzones(context.Background())
				require.NoError(t, err)
				require.Len(t, got, 1)
				assert.Equal(t, "example.com", got[0].Root)
				assert.Equal(t, []string{"ca1"}, got[0].CA)
			},
		},
		{
			name: "certificate authorities",
			path: "/ca",
			body: `[{"id":"ca1","name":"Example CA","desc":"test CA","acme":true,"enabled":true}]`,
			call: func(t *testing.T, c Client) {
				got, err := c.GetCAs(context.Background())
				require.NoError(t, err)
				require.Len(t, got, 1)
				assert.Equal(t, "ca1", got[0].ID)
				assert.Equal(t, "test CA", got[0].Description)
				assert.True(t, got[0].IsAcme)
			},
		},
		{
			name: "certificate authority",
			path: "/ca/ca1",
			body: `{"id":"ca1","name":"Example CA","type":"acme"}`,
			call: func(t *testing.T, c Client) {
				got, err := c.GetCA(context.Background(), "ca1")
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, "ca1", got.ID)
				assert.Equal(t, "acme", got.Type)
			},
		},
		{
			name: "certificates by CA",
			path: "/ca/ca1/crt",
			body: `[` + cert + `]`,
			call: func(t *testing.T, c Client) {
				got, page, err := c.GetCertsByCA(context.Background(), "ca1", nil)
				require.NoError(t, err)
				require.Len(t, got, 1)
				require.NotNil(t, page)
				assert.Equal(t, "www.example.com", got[0].Name)
				assert.Equal(t, uint(3), got[0].RenewCount)
			},
		},
		{
			name: "certificates across CAs",
			path: "/crt",
			body: `[` + cert[:len(cert)-1] + `,"ca":"ca1"}]`,
			call: func(t *testing.T, c Client) {
				got, page, err := c.GetCerts(context.Background(), nil)
				require.NoError(t, err)
				require.Len(t, got, 1)
				require.NotNil(t, page)
				assert.Equal(t, "ca1", got[0].CA)
				assert.Equal(t, "www.example.com", got[0].Name)
			},
		},
		{
			name: "certificate by CA and name",
			path: "/ca/ca1/crt/www.example.com",
			body: cert,
			call: func(t *testing.T, c Client) {
				got, err := c.GetCert(context.Background(), "ca1", "www.example.com")
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, "www.example.com", got.Name)
				assert.Equal(t, "Alice", got.ClaimedBy.Name)
				assert.True(t, got.Valid)
			},
		},
		{
			name: "certificate by name across CAs",
			path: "/crt/www.example.com",
			body: `[` + cert[:len(cert)-1] + `,"ca":"ca1"}]`,
			call: func(t *testing.T, c Client) {
				got, err := c.GetCertAllCAs(context.Background(), "www.example.com")
				require.NoError(t, err)
				require.Len(t, got, 1)
				assert.Equal(t, "ca1", got[0].CA)
			},
		},
		{
			name: "all certificate resources",
			path: "/ca/ca1/crt/www.example.com/pem",
			body: `{"cert":"certificate-pem","key":"key-pem","chain":"chain-pem","root":"root-pem","rootchain":"root-chain-pem","fullchain":"full-chain-pem"}`,
			call: func(t *testing.T, c Client) {
				got, err := c.GetAllResources(context.Background(), "ca1", "www.example.com")
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, "certificate-pem", got.Certificate)
				assert.Equal(t, "key-pem", got.Key)
				assert.Equal(t, "full-chain-pem", got.FullChain)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, tt.path, r.URL.EscapedPath())
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tt.body))
			}))
			tt.call(t, client)
		})
	}
}

func TestClientRawReturnsUnparsedResponse(t *testing.T) {
	const response = `{"unknownField":{"future":true}}`
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(response))
	}))

	reader, err := client.Raw().GetServerInfo(context.Background())
	require.NoError(t, err)
	got, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, response, string(got))
	require.NoError(t, reader.Close())
}
