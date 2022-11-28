package server

import (
	"bytes"
	"errors"
)

const OptimalBufferSize = 1500

type header struct {
	Name  []byte
	Value []byte
}

type HTTPParser struct {
	Method, Path, Version []byte
	contentLength         int
}

const DefaultHeaderSlice = 10

func NewHTTPParser() *HTTPParser {
	return NewSizedHTTPParser(DefaultHeaderSlice)
}

func NewSizedHTTPParser(size int) *HTTPParser {
	return &HTTPParser{
		contentLength: -1,
	}
}

var (
	ErrBadProto    = errors.New("bad protocol")
	ErrMissingData = errors.New("missing data")
)

const (
	eNextHeader int = iota
	eNextHeaderN
	eHeader
	eHeaderValueSpace
	eHeaderValue
	eHeaderValueN
	eMLHeaderStart
	eMLHeaderValue
)

func (hp *HTTPParser) Reset() {
	hp.contentLength = -1
}

// Parse the buffer as an HTTP Request. The buffer must contain the entire
// request or Parse will return ErrMissingData for the caller to get more
// data. (this thusly favors getting a completed request in a single Read()
// call).
//
// Returns the number of bytes used by the header (thus where the body begins).
// Also can return ErrUnsupported if an HTTP feature is detected but not supported.
func (hp *HTTPParser) Parse(input []byte) error {
	var path int
	var ok bool

	total := len(input)

method:
	for i := 0; i < total; i++ {
		switch input[i] {
		case ' ', '\t':
			hp.Method = input[0:i]
			ok = true
			path = i + 1
			break method
		}
	}

	if !ok {
		return ErrMissingData
	}
	ok = false
path:
	for i := path; i < total; i++ {
		switch input[i] {
		case ' ', '\t':
			ok = true
			hp.Path = input[path:i]
			break path
		}
	}
	return nil
}

func (hp *HTTPParser) ContentLength() int {
	return hp.contentLength
}

var cGet = []byte("GET")

func (hp *HTTPParser) Get() bool {
	return bytes.Equal(hp.Method, cGet)
}

var cPost = []byte("POST")

func (hp *HTTPParser) Post() bool {
	return bytes.Equal(hp.Method, cPost)
}
