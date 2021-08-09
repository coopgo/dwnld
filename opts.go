package main

import "time"

type Option func(*Downloader)

func NewWithDefaultPath(path string) Option {
	return func(d *Downloader) {
		d.Path = path
	}
}

func NewWithSilentOutput() Option {
	return func(d *Downloader) {
		d.Silent = true
	}
}

func NewWithMaxConcurrency(i int) Option {
	if i < 0 {
		i = -1
	}
	return func(d *Downloader) {
		d.MaxConcurrency = i
	}
}

func NewWithRefreshRate(dur time.Duration) Option {
	return func(d *Downloader) {
		d.RefreshRate = dur
	}
}
