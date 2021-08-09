package dwnld

import (
	"io"
	"mime"
	"net/http"
	"path"
)

func (rs *resource) getHttpSrc() (io.ReadCloser, error) {
	resp, err := http.Get(rs.url)
	if err != nil {
		return nil, err
	}

	rs.name = httpFilename(resp)
	rs.size = httpFileSize(resp)

	return resp.Body, nil

}

func httpFilename(resp *http.Response) string {

	var name string

	content_disp := resp.Header.Get("Content-Disposition")
	_, params, err := mime.ParseMediaType(content_disp)
	if err == nil {
		name = params["filename"]
	}

	if name == "" {
		name = path.Base(resp.Request.URL.Path)
	}

	if name == "" {
		name = randStringBytes(8)
	}

	return name
}

func httpFileSize(resp *http.Response) int64 {
	var size int64 = -1
	if resp.ContentLength > 0 {
		size = resp.ContentLength
	}
	return size
}
