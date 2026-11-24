package client

import (
	"context"
	"io"
	"net/http"

	"github.com/dns3l/dns3l-go/api"
)

func (c *clientImpl) GetCert(ctx context.Context, caID string, certName string) (*api.CertInfo, error) {
	r, err := c.raw.GetCert(ctx, caID, certName)
	if err != nil {
		return nil, err
	}
	var result *api.CertInfo
	err = c.raw.decodeJsonResponse(r, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *clientRawImpl) GetCert(ctx context.Context, caID string, certName string) (io.ReadCloser, error) {
	url, err := c.fromBase("/ca/%s/crt/%s", caID, certName)
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

func (c *clientImpl) GetCertAllCAs(ctx context.Context, certName string) ([]*api.CertInfoWithCA, error) {
	r, err := c.raw.GetCertAllCAs(ctx, certName)
	if err != nil {
		return nil, err
	}
	var result []*api.CertInfoWithCA
	err = c.raw.decodeJsonResponse(r, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *clientRawImpl) GetCertAllCAs(ctx context.Context, certName string) (io.ReadCloser, error) {
	url, err := c.fromBase("/crt/%s", certName)
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
