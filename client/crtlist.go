package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/dns3l/dns3l-go/api"
	"github.com/dns3l/dns3l-go/client/pagination"
)

type CertQueryOptions struct {
	Limit  uint64
	Offset uint64
	Search string
	Suffix string
}

func (c *clientImpl) GetCertsByCA(ctx context.Context, caID string, opts *CertQueryOptions) (
	[]*api.CertInfo, *pagination.Info, error) {
	return getPaginatedListResource[*api.CertInfo](ctx, c, opts, "/ca/%s/crt", caID)
}

func (c *clientRawImpl) GetCertsByCA(ctx context.Context, caID string, opts *CertQueryOptions) (
	io.ReadCloser, *pagination.Info, error) {
	return c.getPaginatedListResource(ctx, opts, "/ca/%s/crt", caID)
}

func (c *clientImpl) GetCerts(ctx context.Context, opts *CertQueryOptions) (
	[]*api.CertInfoWithCA, *pagination.Info, error) {
	return getPaginatedListResource[*api.CertInfoWithCA](ctx, c, opts, "/crt")
}

func (c *clientRawImpl) GetCerts(ctx context.Context, opts *CertQueryOptions) (
	io.ReadCloser, *pagination.Info, error) {
	return c.getPaginatedListResource(ctx, opts, "/crt")
}

func getPaginatedListResource[T any](ctx context.Context, c *clientImpl,
	opts *CertQueryOptions, path string, params ...string) (
	[]T, *pagination.Info, error) {

	if opts == nil {
		opts = &CertQueryOptions{}
	}

	pg := pagination.NewPaginator(c.hiddenPaging)

	var result []T

	pinfo, _, err := pg.Page(opts.Limit, opts.Offset, func(limit, offset uint64) (
		*pagination.Info, uint64, error) {
		pgOpts := &CertQueryOptions{
			Limit:  limit,
			Offset: offset,
			Search: opts.Search,
			Suffix: opts.Suffix,
		}
		r, pginfo, err := c.raw.getPaginatedListResource(ctx, pgOpts, path, params...)
		if err != nil {
			return nil, 0, err
		}
		var presult []T
		err = c.raw.decodeJsonResponse(r, &presult)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, presult...)
		return pginfo, uint64(len(result)), nil
	})
	if err != nil {
		return nil, nil, err
	}

	return result, pinfo, nil
}

func (c clientRawImpl) getPaginatedListResource(ctx context.Context, opts *CertQueryOptions,
	path string, params ...string) (
	io.ReadCloser, *pagination.Info, error) {

	if opts == nil {
		opts = &CertQueryOptions{}
	}

	url, err := c.fromBase(path, params...)
	if err != nil {
		return nil, nil, err
	}
	pagination.AddPagination(url, opts.Limit, opts.Offset)
	addSearchOpts(url, opts.Search, opts.Suffix)
	rq, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	c.applyDefaultReqHeaders(rq)
	fmt.Println(rq)

	if err != nil {
		return nil, nil, err
	}
	rp, err := c.doSuccessful(rq)
	if err != nil {
		return nil, nil, err
	}

	return rp.Body, pagination.PaginationInfoFromHeaders(rp.Header), nil
}

func addSearchOpts(url *url.URL, search, suffix string) {
	if search == "" && suffix == "" {
		return
	}
	q := url.Query()
	if search != "" {
		q.Add("search", search)
	}
	if suffix != "" {
		q.Add("suffix", suffix)
	}
	url.RawQuery = q.Encode()
}
