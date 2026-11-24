package client

import (
	"context"
	"io"
	"net/http"

	"github.com/dns3l/dns3l-go/api"
)

func (c *clientImpl) GetStatus(ctx context.Context) (*api.ServerStatus, error) {
	r, err := c.raw.GetStatus(ctx)
	if err != nil {
		return nil, err
	}
	var result *api.ServerStatus
	err = c.raw.decodeJsonResponse(r, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *clientRawImpl) GetStatus(ctx context.Context) (io.ReadCloser, error) {
	url, err := c.fromBase("/status")
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
