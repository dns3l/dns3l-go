package client

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/dns3l/dns3l-go/api"
)

func (c *clientImpl) GetAllResources(ctx context.Context, caID, certName string) (
	*api.CertResources, error) {
	r, err := c.raw.GetAllResources(ctx, caID, certName)
	if err != nil {
		return nil, err
	}
	var result *api.CertResources
	err = c.raw.decodeJsonResponse(r, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *clientRawImpl) GetAllResources(ctx context.Context, caID, certName string) (
	io.ReadCloser, error) {
	url, err := c.fromBase("/ca/%s/crt/%s/pem", caID, certName)
	if err != nil {
		return nil, err
	}
	rq, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	c.applyDefaultReqHeaders(rq)
	if err != nil {
		return nil, err
	}
	rp, err := c.doSuccessful(rq)
	if err != nil {
		return nil, err
	}

	return rp.Body, nil
}

func (c *clientImpl) GetResourceCert(ctx context.Context, caID,
	certName string) (*x509.Certificate, error) {
	return c.getResourceSingleCert(ctx, caID, certName, ResourceTypeCert)
}

func (c *clientImpl) GetResourceRoot(ctx context.Context, caID,
	certName string) (*x509.Certificate, error) {
	return c.getResourceSingleCert(ctx, caID, certName, ResourceTypeRoot)
}

func (c *clientImpl) getResourceSingleCert(ctx context.Context, caID,
	certName string, rtype ResourceType) (*x509.Certificate, error) {
	r, err := c.raw.GetResource(ctx, caID, certName, rtype)
	if err != nil {
		return nil, err
	}
	//nolint:errcheck // don't check closes
	defer r.Close()
	pemBytes, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	cert, err := pemBytesToSingleCert(pemBytes)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

func pemBytesToSingleCert(pemData []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("pem data is empty")
	}
	if block.Type != "CERTIFICATE" {
		return nil, errors.New("pem block is not 'CERTIFICATE'")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("could not parse cert from pem data: %w", err)
	}
	return cert, nil
}

func (c *clientImpl) GetResourceKey(ctx context.Context, caID, certName string) (
	*rsa.PrivateKey, error) {

	r, err := c.raw.GetResource(ctx, caID, certName, ResourceTypeKey)
	if err != nil {
		return nil, err
	}
	//nolint:errcheck // don't check closes
	defer r.Close()
	pemBytes, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	key, err := pemBytesToKey(pemBytes)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func pemBytesToKey(pemData []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("pem data is empty")
	}
	if block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("pem block is not 'RSA PRIVATE KEY': %s", block.Type)
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("could not parse cert from pem data: %w", err)
	}
	return key, nil
}

func (c *clientImpl) GetResourceChain(ctx context.Context, caID,
	certName string) ([]*x509.Certificate, error) {
	return c.getResourceChain(ctx, caID, certName, ResourceTypeChain)
}

func (c *clientImpl) GetResourceFullChain(ctx context.Context, caID,
	certName string) ([]*x509.Certificate, error) {
	return c.getResourceChain(ctx, caID, certName, ResourceTypeFullChain)
}

func (c *clientImpl) GetResourceRootChain(ctx context.Context, caID,
	certName string) ([]*x509.Certificate, error) {
	return c.getResourceChain(ctx, caID, certName, ResourceTypeRootChain)
}

func (c *clientImpl) getResourceChain(ctx context.Context, caID,
	certName string, rtype ResourceType) ([]*x509.Certificate, error) {
	r, err := c.raw.GetResource(ctx, caID, certName, rtype)
	if err != nil {
		return nil, err
	}
	//nolint:errcheck // don't check closes
	defer r.Close()
	pemBytes, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	chain, err := pemBytesToChain(pemBytes)
	if err != nil {
		return nil, err
	}

	return chain, nil
}

func pemBytesToChain(pemData []byte) ([]*x509.Certificate, error) {
	var certs []*x509.Certificate
	for block, rest := pem.Decode(pemData); block != nil; block, rest = pem.Decode(rest) {
		if block.Type != "CERTIFICATE" {
			return nil, errors.New("pem block is not 'CERTIFICATE'")
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("could not parse cert from pem data: %w", err)
		}
		certs = append(certs, cert)
	}
	return certs, nil
}

func (c *clientImpl) GetResourceBytes(ctx context.Context, caID,
	certName string, rtype ResourceType) (
	[]byte, error) {
	switch rtype {
	case ResourceTypeCert, ResourceTypeChain, ResourceTypeFullChain,
		ResourceTypeKey, ResourceTypeRoot, ResourceTypeRootChain:
	default:
		return nil, fmt.Errorf("illegal resource type %s", rtype)
	}
	r, err := c.raw.GetResource(ctx, caID, certName, rtype)
	if err != nil {
		return nil, err
	}
	//nolint:errcheck // don't check closes
	defer r.Close()
	pemBytes, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return pemBytes, nil
}

func (c *clientRawImpl) GetResource(ctx context.Context, caID, certName string, rtype ResourceType) (
	io.ReadCloser, error) {
	url, err := c.fromBase("/ca/%s/crt/%s/pem/%s", caID, certName, string(rtype))
	if err != nil {
		return nil, err
	}
	rq, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	c.applyDefaultReqHeaders(rq)
	if err != nil {
		return nil, err
	}
	rp, err := c.doSuccessful(rq)
	if err != nil {
		return nil, err
	}

	return rp.Body, nil
}
