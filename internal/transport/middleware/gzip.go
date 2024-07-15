package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
)

type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

// Header is a wrapper of http.ResponseWriter Header()
func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

// Write is a wrapper of gzip.NewWriter Write()
func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

// WriteHeader is a wrapper of http.ResponseWriter WriteHeader()
func (c *compressWriter) WriteHeader(statusCode int) {
	c.w.Header().Set("Content-Encoding", "gzip")
	c.w.WriteHeader(statusCode)
}

// Close is a wrapper of gzip.NewWriter Close()
func (c *compressWriter) Close() error {
	return c.zw.Close()
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

// Read is a wrapper of gzip.NewWriter Read()
func (c *compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

// Close is a wrapper of gzip.NewWriter Read()
func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}
