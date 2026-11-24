package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/dns3l/dns3l-go/api"
)

var (
	_ Client    = &clientImpl{}
	_ ClientRaw = &clientRawImpl{}
)

type clientImpl struct {
	raw          *clientRawImpl
	hiddenPaging uint64
}

type clientRawImpl struct {
	baseURL     *url.URL
	httpclient  *http.Client
	bearertoken string
	apikey      string
	userAgent   string
}

// NewUnauthenticated returns a new interface to a dns3ld service
// at the specified URL. No authentication is made.
// If desired, specific HTTP client settings
// can be set, otherwise you can set the httpclient to nil.
// Note that certificate claims can take minutes, so
// do not set an httpclient.Timeout too small.
func NewUnauthenticated(baseURL *url.URL, httpclient *http.Client) Client {
	if httpclient == nil {
		httpclient = http.DefaultClient
	}
	return &clientImpl{
		raw: &clientRawImpl{
			baseURL:    baseURL,
			httpclient: httpclient,
		},
	}
}

// NewByOauth2AccessToken returns a new interface to a dns3ld service
// at the specified URL. The client attempts authentication with the given
// token. If desired, specific HTTP client settings
// can be set, otherwise you can set the httpclient to nil.
// Note that certificate claims can take minutes, so
// do not set an httpclient.Timeout too small.
func NewByOauth2AccessToken(baseURL *url.URL, httpclient *http.Client, token string) Client {
	if httpclient == nil {
		httpclient = http.DefaultClient
	}
	return &clientImpl{
		raw: &clientRawImpl{
			baseURL:     baseURL,
			httpclient:  httpclient,
			bearertoken: token,
		},
	}
}

// NewByOauth2AccessToken returns a new interface to a dns3ld service
// at the specified URL. The client attempts authentication with the given
// API key. If desired, specific HTTP client settings
// can be set, otherwise you can set the httpclient to nil.
// Note that certificate claims can take minutes, so
// do not set an httpclient.Timeout too small.
func NewByApiKey(baseURL *url.URL, httpclient *http.Client, apikey string) Client {
	if httpclient == nil {
		httpclient = http.DefaultClient
	}
	return &clientImpl{
		raw: &clientRawImpl{
			baseURL:    baseURL,
			httpclient: httpclient,
			apikey:     apikey,
		},
	}
}

func (c *clientImpl) SetHiddenPaging(hiddenPaging uint64) {
	c.hiddenPaging = hiddenPaging
}

func (c *clientRawImpl) SetUserAgent(userAgent string) {
	c.userAgent = userAgent
}

func (c *clientImpl) Raw() ClientRaw {
	return c.raw
}

func (c *clientRawImpl) fromBase(path string, params ...string) (*url.URL, error) {
	paramsEscaped := make([]any, len(params))
	for i, param := range params {
		paramsEscaped[i] = url.PathEscape(param)
	}
	return c.baseURL.JoinPath(fmt.Sprintf(path, paramsEscaped...)), nil
}

func (c *clientRawImpl) applyDefaultReqHeaders(r *http.Request) {
	h := r.Header
	h.Add("User-Agent", c.getUserAgent())
	if c.apikey != "" {
		h.Add("X-DNS3L-API-Key", c.apikey)
		return
	}
	if c.bearertoken != "" {
		h.Add("Authorization", fmt.Sprintf("Bearer %s", c.bearertoken))
	}
}

func (c *clientRawImpl) getUserAgent() string {
	if c.userAgent != "" {
		return c.userAgent
	}
	return GetDefaultUserAgent()
}

func (c *clientRawImpl) decodeJsonResponse(r io.ReadCloser, v any) error {
	//nolint:errcheck // don't check closes
	defer r.Close()
	d := json.NewDecoder(r)
	return d.Decode(v)
}

func (c *clientImpl) GetServerInfo(ctx context.Context) (*api.ServerInfo, error) {
	r, err := c.raw.GetServerInfo(ctx)
	if err != nil {
		return nil, err
	}
	result := &api.ServerInfo{}
	err = c.raw.decodeJsonResponse(r, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *clientRawImpl) GetServerInfo(ctx context.Context) (io.ReadCloser, error) {
	url, err := c.fromBase("/info")
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
