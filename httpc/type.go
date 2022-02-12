package httpc

import (
  "net/url"
  "path"
)

type RawURL string

func (r *RawURL) String() (l string, err error) {
  u,err := url.Parse(string(*r))
  if err != nil {
    return "", err
  }
  u.Path = path.Clean(u.Path)
  if u.Scheme == "" {
    u.Scheme = "http"
  }

  return u.String(), nil
}
