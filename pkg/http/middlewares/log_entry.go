package middlewares

import (
	"errors"
	"io"
	"net"
	"net/http"
	"time"
)

type (
	logEntry struct {
		ReceivedTime      time.Time
		RequestMethod     string
		RequestURL        string
		RequestHeaderSize int64
		RequestBodySize   int64
		UserAgent         string
		Referer           string
		Proto             string

		RemoteIP string
		ServerIP string

		Status             int
		ResponseHeaderSize int64
		ResponseBodySize   int64
		Latency            time.Duration
	}
	writeCounter  int64
	responseStats struct {
		w     http.ResponseWriter
		hsize int64
		wc    writeCounter
		code  int
	}
	readCounterCloser struct {
		r   io.ReadCloser
		n   int64
		err error
	}
)

func ipFromHostPort(hp string) string {
	h, _, err := net.SplitHostPort(hp)
	if err != nil {
		return ""
	}
	if len(h) > 0 && h[0] == '[' {
		return h[1 : len(h)-1]
	}
	return h
}

func (rcc *readCounterCloser) Read(p []byte) (n int, err error) {
	if rcc.err != nil {
		return 0, rcc.err
	}
	n, rcc.err = rcc.r.Read(p)
	rcc.n += int64(n)
	return n, rcc.err
}

func (rcc *readCounterCloser) Close() error {
	rcc.err = errors.New("read from closed reader")
	return rcc.r.Close()
}

func (wc *writeCounter) Write(p []byte) (n int, err error) {
	*wc += writeCounter(len(p))
	return len(p), nil
}

func headerSize(h http.Header) int64 {
	var wc writeCounter
	err := h.Write(&wc)
	if err != nil {
		return 0
	}
	return int64(wc) + 2 // for CRLF
}

func (r *responseStats) Header() http.Header {
	return r.w.Header()
}

func (r *responseStats) WriteHeader(statusCode int) {
	if r.code != 0 {
		return
	}
	r.hsize = headerSize(r.w.Header())
	r.w.WriteHeader(statusCode)
	r.code = statusCode
}

func (r *responseStats) Write(p []byte) (n int, err error) {
	if r.code == 0 {
		r.WriteHeader(http.StatusOK)
	}
	n, err = r.w.Write(p)
	if err != nil {
		return 0, err
	}
	_, err = r.wc.Write(p[:n])
	if err != nil {
		return 0, err
	}
	return
}

func (r *responseStats) size() (hdr, body int64) {
	if r.code == 0 {
		return headerSize(r.w.Header()), 0
	}
	// Use the header size from the time WriteHeader was called.
	// The Header map can be mutated after the call to add HTTP Trailers,
	// which we don't want to count.
	return r.hsize, int64(r.wc)
}
