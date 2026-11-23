package pagination

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type Info struct {
	offset     uint64
	limit      uint64
	totalcount uint64
}

func PaginationInfoFromHeaders(header http.Header) *Info {
	p := &Info{}
	p.offset = strToUint64_0(header.Get("Page-Offset"))
	p.limit = strToUint64_0(header.Get("Page-Limit"))
	p.totalcount = strToUint64_0(header.Get("Total-Count"))
	return p
}

func (p *Info) String() string {
	if p.limit > 0 {
		return fmt.Sprintf("Showing element %d - %d of %d elements", p.offset+1, p.offset+p.limit, p.totalcount)
	}
	return ""
}

func strToUint64_0(s string) uint64 {
	if s == "" {
		return 0
	}
	u, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0
	}
	return u
}

type Paginator struct {
	hidden uint64
}

func NewPaginator(hidden uint64) *Paginator {
	return &Paginator{
		hidden: hidden,
	}
}

func (h *Paginator) Page(olimit, ooffset uint64,
	pg func(limit, offset uint64) (*Info, uint64, error),
) (*Info, uint64, error) {
	if h.hidden == 0 || (olimit > 0 && olimit <= h.hidden) {
		// no hidden paging
		return pg(olimit, ooffset)
	}

	curoffset := ooffset
	curlimit := h.hidden
	total := uint64(0)
	for {
		pinfo, num, err := pg(curlimit, curoffset)
		if err != nil {
			return nil, total, err
		}
		total += num
		if pinfo == nil || pinfo.limit != curlimit || pinfo.offset != curoffset {
			// prevent infinite loops if server did not recognize pagination
			// Behave as if pagination was not supported
			return &Info{}, total, nil
		}
		if num < h.hidden || (olimit > 0 && total >= olimit) {
			// we have all elements, or reached the caller-requested limit
			return &Info{offset: ooffset, limit: olimit, totalcount: pinfo.totalcount}, total, nil
		}
		curoffset += num
		if olimit > 0 {
			curlimit = min(h.hidden, olimit-total)
		}
	}

}

func AddPagination(url *url.URL, limit uint64, offset uint64) {
	if limit <= 0 && offset <= 0 {
		return
	}
	q := url.Query()
	if limit > 0 {
		q.Add("limit", strconv.FormatUint(limit, 10))
	}
	if offset > 0 {
		q.Add("offset", strconv.FormatUint(offset, 10))
	}
	url.RawQuery = q.Encode()
}
