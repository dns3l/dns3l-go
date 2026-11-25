package client

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testCertificatePEM(t *testing.T, commonName string) []byte {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: commonName},
		NotBefore:             time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		NotAfter:              time.Date(2034, 1, 1, 0, 0, 0, 0, time.UTC),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}

func testRSAPrivateKeyPEM(t *testing.T) (*rsa.PrivateKey, []byte) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return key, pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
}

func TestPEMResourceMethods(t *testing.T) {
	certPEM := testCertificatePEM(t, "leaf.example.com")
	rootPEM := testCertificatePEM(t, "root.example.com")
	chainPEM := append(bytes.Clone(certPEM), rootPEM...)
	key, keyPEM := testRSAPrivateKeyPEM(t)

	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.EscapedPath() {
		case "/ca/ca1/crt/leaf.example.com/pem/crt":
			_, _ = w.Write(certPEM)
		case "/ca/ca1/crt/leaf.example.com/pem/root":
			_, _ = w.Write(rootPEM)
		case "/ca/ca1/crt/leaf.example.com/pem/key":
			_, _ = w.Write(keyPEM)
		case "/ca/ca1/crt/leaf.example.com/pem/chain",
			"/ca/ca1/crt/leaf.example.com/pem/fullchain",
			"/ca/ca1/crt/leaf.example.com/pem/rootchain":
			_, _ = w.Write(chainPEM)
		default:
			assert.Failf(t, "unexpected resource path", "%s", r.URL.EscapedPath())
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	cert, err := client.GetResourceCert(context.Background(), "ca1", "leaf.example.com")
	require.NoError(t, err)
	assert.Equal(t, "leaf.example.com", cert.Subject.CommonName)

	root, err := client.GetResourceRoot(context.Background(), "ca1", "leaf.example.com")
	require.NoError(t, err)
	assert.Equal(t, "root.example.com", root.Subject.CommonName)

	chain, err := client.GetResourceChain(context.Background(), "ca1", "leaf.example.com")
	require.NoError(t, err)
	require.Len(t, chain, 2)
	assert.Equal(t, "leaf.example.com", chain[0].Subject.CommonName)
	assert.Equal(t, "root.example.com", chain[1].Subject.CommonName)

	fullChain, err := client.GetResourceFullChain(context.Background(), "ca1", "leaf.example.com")
	require.NoError(t, err)
	require.Len(t, fullChain, 2)

	rootChain, err := client.GetResourceRootChain(context.Background(), "ca1", "leaf.example.com")
	require.NoError(t, err)
	require.Len(t, rootChain, 2)

	gotKey, err := client.GetResourceKey(context.Background(), "ca1", "leaf.example.com")
	require.NoError(t, err)
	assert.Equal(t, key.N, gotKey.N)
	assert.Equal(t, key.D, gotKey.D)

	for _, resourceType := range []ResourceType{
		ResourceTypeCert, ResourceTypeChain, ResourceTypeFullChain,
		ResourceTypeKey, ResourceTypeRoot, ResourceTypeRootChain,
	} {
		data, err := client.GetResourceBytes(context.Background(), "ca1", "leaf.example.com", resourceType)
		require.NoError(t, err)
		require.NotEmpty(t, data)
	}
}

func TestGetResourceBytesRejectsUnknownType(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(t, "unexpected request for invalid resource type")
	}))
	data, err := client.GetResourceBytes(context.Background(), "ca1", "leaf.example.com", ResourceType("unknown"))
	require.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "illegal resource type unknown")
}

func TestPEMParsingErrors(t *testing.T) {
	_, err := pemBytesToSingleCert(nil)
	require.Error(t, err)
	assert.Equal(t, "pem data is empty", err.Error())

	wrongType := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte("key")})
	_, err = pemBytesToSingleCert(wrongType)
	require.Error(t, err)
	assert.Equal(t, "pem block is not 'CERTIFICATE'", err.Error())

	badCertificate := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte("not DER")})
	_, err = pemBytesToSingleCert(badCertificate)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "could not parse cert from pem data")

	_, err = pemBytesToChain(wrongType)
	require.Error(t, err)
	assert.Equal(t, "pem block is not 'CERTIFICATE'", err.Error())
	_, err = pemBytesToChain(badCertificate)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "could not parse cert from pem data")

	_, err = pemBytesToKey(nil)
	require.Error(t, err)
	assert.Equal(t, "pem data is empty", err.Error())

	_, err = pemBytesToKey(wrongType)
	require.Error(t, err)
	assert.Equal(t, "pem block is not 'RSA PRIVATE KEY': PRIVATE KEY", err.Error())

	badKey := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: []byte("not DER")})
	_, err = pemBytesToKey(badKey)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "could not parse cert from pem data")
}
