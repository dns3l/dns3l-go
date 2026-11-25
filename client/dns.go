package client

import (
	"context"
	"io"
	"net/http"

	"github.com/dns3l/dns3l-go/api"
)

func (c *clientImpl) GetDNSHandlers(ctx context.Context) ([]*api.DNSHandlerInfo, error) {
	r, err := c.raw.GetDNSHandlers(ctx)
	if err != nil {
		return nil, err
	}
	var result []*api.DNSHandlerInfo
	err = c.raw.decodeJsonResponse(r, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *clientRawImpl) GetDNSHandlers(ctx context.Context) (io.ReadCloser, error) {
	url, err := c.fromBase("/dns")
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

func (c *clientImpl) GetDNSRootzones(ctx context.Context) ([]*api.DNSRootzoneInfo, error) {
	r, err := c.raw.GetDNSRootzones(ctx)
	if err != nil {
		return nil, err
	}
	var result []*api.DNSRootzoneInfo
	err = c.raw.decodeJsonResponse(r, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *clientRawImpl) GetDNSRootzones(ctx context.Context) (io.ReadCloser, error) {
	url, err := c.fromBase("/dns/rtzn")
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
