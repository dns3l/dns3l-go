package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/dns3l/dns3l-go/api"
)

func (c *clientImpl) ClaimCertificate(ctx context.Context, caID string, req *api.CertClaimInfo) error {
	return c.raw.ClaimCertificate(ctx, caID, req)
}

func (c *clientRawImpl) ClaimCertificate(ctx context.Context, caID string, req *api.CertClaimInfo) error {
	url, err := c.fromBase("/ca/%s/crt", caID)
	if err != nil {
		return err
	}
	reqjson, err := json.Marshal(req)
	if err != nil {
		return err
	}
	reqr := bytes.NewBuffer(reqjson)
	rq, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), reqr)
	c.applyDefaultReqHeaders(rq)
	rq.Header.Add("Content-Type", "application/json")
	if err != nil {
		return err
	}

	return c.doSuccessfulNoRespBody(rq)
}
