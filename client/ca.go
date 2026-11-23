package client

import (
	"context"
	"io"
	"net/http"

	"github.com/dns3l/dns3l-go/api"
)

func (c *clientImpl) GetCAs(ctx context.Context) ([]*api.CAInfo, error) {
	r, err := c.raw.GetCAs(ctx)
	if err != nil {
		return nil, err
	}
	var result []*api.CAInfo
	err = c.raw.decodeJsonResponse(r, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *clientRawImpl) GetCAs(ctx context.Context) (io.Reader, error) {
	url, err := c.fromBase("/ca")
	if err != nil {
		return nil, err
	}
	rq, err := http.NewRequestWithContext(ctx, "GET", url.String(), nil)
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

func (c *clientImpl) GetCA(ctx context.Context, caID string) (*api.CAInfo, error) {
	r, err := c.raw.GetCA(ctx, caID)
	if err != nil {
		return nil, err
	}
	var result *api.CAInfo
	err = c.raw.decodeJsonResponse(r, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *clientRawImpl) GetCA(ctx context.Context, caID string) (io.Reader, error) {
	url, err := c.fromBase("/ca/%s", caID)
	if err != nil {
		return nil, err
	}
	rq, err := http.NewRequestWithContext(ctx, "GET", url.String(), nil)
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
