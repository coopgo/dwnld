package main

import (
	"errors"
	"io"
	"net"
	"net/url"
	"path"
	"time"

	"github.com/jlaffaye/ftp"
)

type Ftp struct {
	User     string
	Password string
	Host     string
	Filepath string
}

func (rs *resource) getFtpSrc() (io.ReadCloser, error) {
	ft, err := parseFtpUrl(rs.url)
	if err != nil {
		return nil, err
	}

	c, err := ftp.Dial(ft.Host, ftp.DialWithTimeout(15*time.Second))
	if err != nil {
		return nil, err
	}

	err = c.Login(ft.User, ft.Password)
	if err != nil {
		return nil, err
	}

	rs.name = path.Base(ft.Filepath)

	return c.Retr(ft.Filepath)
}

func parseFtpUrl(link string) (Ftp, error) {

	ftp := Ftp{
		User:     "anonymous",
		Password: "anonymous",
	}

	u, err := url.Parse(link)
	if err != nil {
		return ftp, err
	}

	if name := u.User.Username(); name != "" {
		ftp.User = name
	}
	if pwd, t := u.User.Password(); t && pwd != "" {
		ftp.Password = pwd
	}

	if u.Host == "" {
		return ftp, errors.New("parseFtpUrl: invalid host")
	}

	host, port, _ := net.SplitHostPort(u.Host)

	if host == "" {
		host = u.Host
	}
	if port == "" {
		port = "21"
	}

	ftp.Host = host + ":" + port

	if u.Path == "" {
		return ftp, errors.New("parseFtpUrl: no path found in url")
	}

	ftp.Filepath = u.Path

	return ftp, nil
}
