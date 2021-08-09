package dwnld

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path"
	"time"

	"github.com/coopgo/dwnld/writer"
)

type Downloader struct {
	Path   string
	Silent bool

	// 0 is no concurrency
	// positive is max number of concurrent download
	// negative is no limit to concurrent download
	MaxConcurrency int

	RefreshRate time.Duration
}

func New(opts ...Option) Downloader {
	dwnlder := Downloader{
		Path:        ".",
		RefreshRate: 25 * time.Millisecond,
	}

	for _, opt := range opts {
		opt(&dwnlder)
	}

	return dwnlder
}

type Info struct {
	Name  string
	Path  string
	Size  int64
	Start time.Time
	End   time.Time
	Error error
}

func (dwnlder Downloader) Download(urls ...string) []Info {

	// Setup the resources
	rss := make([]resource, len(urls))
	for i := range rss {
		rss[i].path = dwnlder.Path
		rss[i].url = urls[i]
		rss[i].ch = make(chan status, 1)
	}

	// Determine the number of worker
	n := dwnlder.MaxConcurrency
	if n < 0 {
		n = len(rss)
	} else if n == 0 {
		n = 1
	}

	// Make the channel that will give the resources to the workers
	rssch := make(chan resource, len(rss))

	// Launch the workers
	for i := 0; i < n; i++ {
		go worker(rssch)
	}

	// Launch the routine that will send the resources to the channel
	go func() {
		for _, rs := range rss {
			rssch <- rs
		}

		close(rssch)
	}()

	writer := writer.New()

	pad := 0

	// Receiver of status
	for {

		// update
		for i := range rss {
			if rss[i].ch != nil {
				select {
				case st, ok := <-rss[i].ch:
					if !ok {
						rss[i].ch = nil
						continue
					}
					rss[i].update(st)
				default:
				}
			}

			if len(rss[i].name) > pad {
				pad = len(rss[i].name)
			}
		}

		// print
		if !dwnlder.Silent {
			for _, rs := range rss {
				print(writer, rs, pad)
			}
			writer.Flush()
		}

		// check if all workers have finished
		finished := true
		for _, rs := range rss {
			if rs.ch != nil {
				finished = false
			}
		}

		if finished {
			break
		}

		time.Sleep(dwnlder.RefreshRate)
	}

	// We create the infos
	infos := make([]Info, len(rss))
	for i := range infos {
		infos[i] = Info{
			Name:  rss[i].name,
			Size:  rss[i].size,
			Error: rss[i].err,
			Start: rss[i].start,
			End:   rss[i].end,
		}

		if infos[i].Name != "" && infos[i].Error == nil {
			infos[i].Path = path.Join(dwnlder.Path, infos[i].Name)
		}
	}

	return infos
}

func (rs *resource) update(st status) {
	rs.written = st.written
	rs.status = st.code
	rs.err = st.err
	rs.size = st.size
	if st.name != "" {
		rs.name = st.name
	}
	if !st.start.Equal(time.Time{}) {
		rs.start = st.start
	}

	if rs.status == completeCode || rs.status == errorCode {
		rs.end = time.Now()
	}
}

type resource struct {
	path    string
	name    string
	url     string
	scheme  scheme
	size    int64
	written int64
	status  code

	start time.Time
	end   time.Time

	err error
	ch  chan status
}

func (rs *resource) download() {
	defer close(rs.ch)

	rs.scheme = findScheme(rs.url)

	if rs.scheme == unknownScheme {
		rs.ch <- status{err: errors.New("unknown scheme"), code: -1}
		return
	}

	src, err := rs.getSrc()
	if err != nil {
		rs.ch <- status{err: err, code: -1}
		return
	}
	defer src.Close()

	p := path.Join(rs.path, rs.name)
	out, err := os.Create(p)
	if err != nil {
		// we try to change the name
		rs.name = randStringBytes(8)
		p = path.Join(rs.path, rs.name)
		out, err = os.Create(p)
		if err != nil {
			rs.ch <- status{err: err, code: -1}
			return
		}
	}
	defer out.Close()

	size, err := rs.copy(src, out)
	if err != nil {
		rs.ch <- status{err: err, code: -1}
		return
	}

	if rs.size <= 0 {
		rs.size = size
	}

	if rs.size != size {
		rs.ch <- status{err: errors.New("invalid size after download"), code: -1}
		return
	}

	rs.ch <- status{name: rs.name, code: 2, size: rs.size}
}

func (rs *resource) getSrc() (io.ReadCloser, error) {
	if rs.scheme == httpScheme {
		src, err := rs.getHttpSrc()
		if err != nil {
			err = fmt.Errorf("could not get http source: %w", err)
		}
		return src, err
	}

	if rs.scheme == ftpScheme {
		src, err := rs.getFtpSrc()
		if err != nil {
			err = fmt.Errorf("could not get ftp source: %w", err)
		}

		return src, err
	}

	return nil, errors.New("could not get unknown scheme source")
}

// Modified from copybuffer io.go: https://golang.org/src/io/io.go
func (rs *resource) copy(src io.Reader, des io.Writer) (int64, error) {

	buf := make([]byte, 32*1024)

	var err error
	var written int64

	start := time.Now()

	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := des.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = errors.New("invalid write result")
				}
			}
			written += int64(nw)
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}

		if rs.ch != nil {
			st := status{
				written: written,
				code:    1,
				name:    rs.name,
				size:    rs.size,
				start:   start,
			}

			select {
			case rs.ch <- st:
			default:
			}
		}

	}

	return int64(written), err
}

func worker(rss <-chan resource) {
	for rs := range rss {
		rs.download()
	}
}

type status struct {
	written int64
	size    int64
	name    string
	start   time.Time
	code    code
	err     error
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

type code int

const (
	errorCode    = -1
	startingCode = 0
	dwnldingCode = 1
	completeCode = 2
)
