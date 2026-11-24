package client

import (
	"fmt"
	"net/http"

	"github.com/dns3l/dns3l-go/api"
)

func (c *clientRawImpl) doSuccessful(rq *http.Request) (*http.Response, error) {
	rp, err := c.httpclient.Do(rq)
	if err != nil {
		return nil, err
	}
	err = c.checkSuccess(rp)
	if err != nil {
		return nil, err
	}
	return rp, err
}

func (c *clientRawImpl) doSuccessfulNoRespBody(rq *http.Request) error {
	rp, err := c.doSuccessful(rq)
	if err != nil {
		return err
	}
	return rp.Body.Close()
}

func (c *clientRawImpl) checkSuccess(rp *http.Response) error {
	if rp.StatusCode >= 200 && rp.StatusCode < 300 {
		return nil
	}

	var errMsg *api.ErrorMsg

	err := c.decodeJsonResponse(rp.Body, &errMsg)
	if err != nil {
		err = fmt.Errorf("could not decode server error msg: %w, code=%d", err, rp.StatusCode)
	} else {
		// sentinel errors only for the common server return codes
		switch rp.StatusCode {
		case http.StatusBadRequest:
			err = fmt.Errorf("server sent: %s, code=%d: %w", errMsg.Message, rp.StatusCode, ErrBadRequest)
		case http.StatusNotFound:
			err = fmt.Errorf("server sent: %s, code=%d: %w", errMsg.Message, rp.StatusCode, ErrNotFound)
		case http.StatusConflict:
			err = fmt.Errorf("server sent: %s, code=%d: %w", errMsg.Message, rp.StatusCode, ErrConflict)
		default:
			err = fmt.Errorf("server sent: %s, code=%d", errMsg.Message, rp.StatusCode)
		}
	}

	return err

}
