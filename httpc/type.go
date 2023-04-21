package httpc

import (
	"net/url"
	"path"
)

type RawURL string

func (r RawURL) String() string {
	u, err := url.Parse(string(r))
	if err != nil {
		return string(r)
	}
	u.Path = path.Clean(u.Path)
	if u.Scheme == "" {
		u.Scheme = "http"
	}

	return u.String()
}
