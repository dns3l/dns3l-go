package client

import (
	"context"
	"net/http"
)

func (c *clientImpl) DeleteCertByCA(ctx context.Context, caID, crtName string) error {
	return c.raw.DeleteCertByCA(ctx, caID, crtName)
}

func (c *clientRawImpl) DeleteCertByCA(ctx context.Context, caID, crtName string) error {
	url, err := c.fromBase("/ca/%s/crt/%s", caID, crtName)
	if err != nil {
		return err
	}
	rq, err := http.NewRequestWithContext(ctx, "DELETE", url.String(), nil)
	c.applyDefaultReqHeaders(rq)
	if err != nil {
		return err
	}

	_, err = c.doSuccessful(rq)
	if err != nil {
		return err
	}

	return nil
}

func (c *clientImpl) DeleteCertAllCAs(ctx context.Context, crtName string) error {
	return c.raw.DeleteCertAllCAs(ctx, crtName)
}

func (c *clientRawImpl) DeleteCertAllCAs(ctx context.Context, crtName string) error {
	url, err := c.fromBase("/crt/%s", crtName)
	if err != nil {
		return err
	}
	rq, err := http.NewRequestWithContext(ctx, "DELETE", url.String(), nil)
	c.applyDefaultReqHeaders(rq)
	if err != nil {
		return err
	}

	_, err = c.doSuccessful(rq)
	if err != nil {
		return err
	}

	return nil
}
