package store

import "net/url"

func parseURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, errInvalidURL
	}
	return u, nil
}

var errInvalidURL = errString("无效的 URL")

type errString string

func (e errString) Error() string { return string(e) }
