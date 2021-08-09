package dwnld

import "net/url"

type scheme int

const (
	unknownScheme = iota
	httpScheme
	ftpScheme
)

func findScheme(str string) scheme {
	u, err := url.Parse(str)
	if err != nil {
		return unknownScheme
	}

	if u.Scheme == "http" || u.Scheme == "https" {
		return httpScheme
	}
	if u.Scheme == "ftp" {
		return ftpScheme
	}

	return unknownScheme
}
