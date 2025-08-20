package proxy

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"errors"
	"io"
	"net/http"

	"github.com/andybalholm/brotli"
)

func decodeBody(encoding string, body bytes.Buffer) ([]byte, error) {
	reader := io.NopCloser(&body)
	defer reader.Close()

	var err error

	switch encoding {
	case "gzip":
		reader, err = gzip.NewReader(reader)
	case "deflate":
		reader, err = zlib.NewReader(reader)
	case "br":
		reader = io.NopCloser(brotli.NewReader(reader))
	case "identity", "":
	default:
		return nil, errors.New("unsupported encoding")
	}

	if err != nil {
		return nil, err
	}

	return io.ReadAll(reader)
}

func buildURL(req *http.Request) string {
	scheme := "http"
	if req.TLS != nil {
		scheme = "https"
	}

	return scheme + "://" + req.Host + req.URL.RequestURI()
}
