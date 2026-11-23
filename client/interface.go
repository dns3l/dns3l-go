package client

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"io"

	"github.com/dns3l/dns3l-go/api"
	"github.com/dns3l/dns3l-go/client/pagination"
)

type ResourceType string

const (
	ResourceTypeCert      ResourceType = "crt"
	ResourceTypeKey       ResourceType = "key"
	ResourceTypeChain     ResourceType = "chain"
	ResourceTypeRoot      ResourceType = "root"
	ResourceTypeRootChain ResourceType = "rootchain"
	ResourceTypeFullChain ResourceType = "fullchain"
)

// ClientRaw provides access to a DNS3L service. Authentication has been configured before
// the interface is provided, either by NewUnauthed(..), NewByOauth2AccessToken(..),
// NewByApiKey(..)
type Client interface {

	// Raw returns an interface to the DNS3L service with similar scope the DNS3L service,
	// but does not attempt to parse responses, instead, it returns an io.Reader,
	// returning the streamed, raw JSON response from the server.
	Raw() ClientRaw

	// SetHiddenPaging allows setting a hidden page size. Only effective on GetCerts and
	// GetCertsByCA, and not for the raw interface. If not 0, the request to the server sets the page
	// limit to the specified
	// value. While the size of the result list is equal to the page size (i.e. more results to expect),
	// the next rows are retrieved in successive HTTP requests with the same page size. Finally,
	// multiple requests are concatenated and returned.
	SetHiddenPaging(hiddenpaging uint64)

	// GetServerInfo returns public information about the server: versions, contacts, auth and renewal.
	GetServerInfo(ctx context.Context) (*api.ServerInfo, error)

	// GetStatus returns information about the server's health and if dependencies are up.
	GetStatus(ctx context.Context) (*api.ServerStatus, error)

	// GetDNSHandlers returns public information about the DNS plugins that are configured.
	GetDNSHandlers(ctx context.Context) ([]*api.DNSHandlerInfo, error)

	// GetDNSRootzones returns public information the DNS rootzones managed by this instance.
	GetDNSRootzones(ctx context.Context) ([]*api.DNSRootzoneInfo, error)

	// GetCAs lists public information about the certificate authorities (CAs) configured
	// on this instance.
	GetCAs(ctx context.Context) ([]*api.CAInfo, error)

	// GetCA returns public information about a specific certificate authority (CA) given
	// by its ID
	GetCA(ctx context.Context, caID string) (*api.CAInfo, error)

	// GetCertsByCA returns information about all certificates currently managed by this instance and issued
	// by a specific certificate authority (CA), given by its ID. You can optionally supply
	// query options, like paging and search parameters.
	GetCertsByCA(ctx context.Context, caID string, opts *CertQueryOptions) (
		[]*api.CertInfo, *pagination.Info, error)

	// GetCerts returns information about all certificates currently managed by this instance.
	// You can optionally supply query options, like paging and search parameters.
	GetCerts(ctx context.Context, opts *CertQueryOptions) (
		[]*api.CertInfoWithCA, *pagination.Info, error)

	// GetCert returns information about a certificate managed by this instance, given by its name,
	// and issued by a specific certificate authority (CA), given by its ID.
	GetCert(ctx context.Context, caID string, certName string) (*api.CertInfo, error)

	// GetCertAllCAs returns information about all certificates issued by any CA on the instance having
	// the name certName.
	GetCertAllCAs(ctx context.Context, certName string) ([]*api.CertInfoWithCA, error)

	// GetAllResources returns a struct containing all cert and key material associated with
	// a specific certificate. All resources are strings containing the PEM-encoded resource.
	GetAllResources(ctx context.Context, caID, certName string) (
		*api.CertResources, error)

	// GetResourceKey returns the private key in its rsa.PrivateKey Go representation.
	GetResourceKey(ctx context.Context, caID, certName string) (*rsa.PrivateKey, error)

	// GetResourceCert returns the certificate in its X509 Go representation.
	GetResourceCert(ctx context.Context, caID,
		certName string) (*x509.Certificate, error)

	// GetResourceRoot returns the certificate's root certificate in its X509 Go representation.
	GetResourceRoot(ctx context.Context, caID,
		certName string) (*x509.Certificate, error)

	// GetResourceChain returns all intermediate certificates of the certificate (excluding the root).
	GetResourceChain(ctx context.Context, caID,
		certName string) ([]*x509.Certificate, error)

	// GetResourceFullChain returns the entire chain including the certificate itself, all intermediates,
	// and the root certificate.
	GetResourceFullChain(ctx context.Context, caID,
		certName string) ([]*x509.Certificate, error)

	// GetResourceRootChain returns all intermediates and the root certificate.
	GetResourceRootChain(ctx context.Context, caID,
		certName string) ([]*x509.Certificate, error)

	// GetResourceBytes returns a resource in its byte representation (PEM-encoded)
	GetResourceBytes(ctx context.Context, caID,
		certName string, rtype ResourceType) ([]byte, error)
	CommonClient
}

// ClientRaw provides access to the DNS3L service like *Client*,
// but does not attempt to parse responses, instead, it returns an io.Reader,
// returning the streamed, raw JSON response from the server.
type ClientRaw interface {

	// SetUserAgent sets a custom User-Agent to use for requests to the server. If not set,
	// a versioned default will be used.
	SetUserAgent(userAgent string)

	// GetServerInfo returns public information the server: versions, contacts, auth and renewal.
	GetServerInfo(ctx context.Context) (io.Reader, error)

	// GetStatus returns information about the server's health and if dependencies are up.
	GetStatus(ctx context.Context) (io.Reader, error)

	// GetDNSHandlers returns public information about the DNS plugins that are configured.
	GetDNSHandlers(ctx context.Context) (io.Reader, error)

	// GetDNSRootzones returns public information the DNS rootzones managed by this instance.
	GetDNSRootzones(ctx context.Context) (io.Reader, error)

	// GetCAs lists public information about the certificate authorities (CAs) configured
	// on this instance.
	GetCAs(ctx context.Context) (io.Reader, error)

	// GetCA returns public information about a specific certificate authority (CA) given
	// by its ID
	GetCA(ctx context.Context, caID string) (io.Reader, error)

	// GetCertsByCA returns information about all certificates currently managed by this instance and issued
	// by a specific certificate authority (CA), given by its ID. You can optionally supply
	// query options, like paging and search parameters.
	GetCertsByCA(ctx context.Context, caID string, opts *CertQueryOptions) (
		io.Reader, *pagination.Info, error)

	// GetCerts returns information about all certificates currently managed by this instance.
	// You can optionally supply query options, like paging and search parameters.
	GetCerts(ctx context.Context, opts *CertQueryOptions) (
		io.Reader, *pagination.Info, error)

	// GetCert returns information about a certificate managed by this instance, given by its name,
	// and issued by a specific certificate authority (CA), given by its ID.
	GetCert(ctx context.Context, caID string, certName string) (io.Reader, error)

	// GetCertAllCAs returns information about all certificates issued by any CA on the instance having
	// the name certName.
	GetCertAllCAs(ctx context.Context, certName string) (io.Reader, error)

	// GetAllResources returns a JSON map containing all cert and key material associated with
	// a specific certificate. All resources are strings containing the PEM-encoded resource.
	GetAllResources(ctx context.Context, caID, certName string) (
		io.Reader, error)

	// GetResource returns a resource in its byte representation (PEM-encoded)
	GetResource(ctx context.Context, caID, certName string, rtype ResourceType) (
		io.Reader, error)
	CommonClient
}

type CommonClient interface {
	ClaimCertificate(ctx context.Context, caID string, req *api.CertClaimInfo) error
	DeleteCertByCA(ctx context.Context, caID, crtName string) error
	DeleteCertAllCAs(ctx context.Context, crtName string) error
}
