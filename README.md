# dwnld [![GoDoc](https://godoc.org/github.com/coopgo/dwnld?status.svg)](https://godoc.org/github.com/coopgo/dwnld)

dwnld is a go library for downloading files over http/https or ftp. You can download multiple files at the same time and have a terminal output in realtime.

## Example

```go
urls := []string{"https://speed.hetzner.de/100MB.bin", "https://speed.hetzner.de/1GB.bin", "ftp://speedtest.tele2.net/10MB.zip"}

// Create a new downloader manager with configuration
// Here we configure it to download all urls concurently
dwnlder := dwnld.New(dwnld.NewWithMaxConcurrency(-1))

// We start the download
infos := dwnlder.Download(urls...)

// infos contains information about all the downloaded files or the errors
for _, info := range infos {
    fmt.Println(info)
}
```

## Installation

```sh
$ go get -v github.com/coopgo/dwnld
```