package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"strconv"
)

// StatusCode represents an HTTP status code
type StatusCode int

// HTTP status codes we support
const (
	StatusOK                  StatusCode = 200
	StatusBadRequest         StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

// WriteStatusLine writes the HTTP status line to the writer
func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	var reasonPhrase string
	
	switch statusCode {
	case StatusOK:
		reasonPhrase = "OK"
	case StatusBadRequest:
		reasonPhrase = "Bad Request"
	case StatusInternalServerError:
		reasonPhrase = "Internal Server Error"
	default:
		reasonPhrase = ""
	}
	
	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", int(statusCode), reasonPhrase)
	_, err := w.Write([]byte(statusLine))
	return err
}

// GetDefaultHeaders returns the default headers for our responses
func GetDefaultHeaders(contentLen int) headers.Headers {
	defaultHeaders := headers.NewHeaders()
	defaultHeaders.Set("Content-Length", strconv.Itoa(contentLen))
	defaultHeaders.Set("Connection", "close")
	defaultHeaders.Set("Content-Type", "text/plain")
	return defaultHeaders
}

// WriteHeaders writes HTTP headers to the writer
func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for key, value := range headers {
		headerLine := fmt.Sprintf("%s: %s\r\n", key, value)
		_, err := w.Write([]byte(headerLine))
		if err != nil {
			return err
		}
	}
	
	// Write empty line to separate headers from body
	_, err := w.Write([]byte("\r\n"))
	return err
}

// writerState tracks the state of the response writer
type writerState int

const (
	stateStart writerState = iota
	stateStatusWritten
	stateHeadersWritten
	stateBodyWritten
	stateChunkedBodyWriting
	stateChunkedBodyDone
	stateTrailersWritten
)

// Writer provides a structured way to write HTTP responses
type Writer struct {
	writer io.Writer
	state  writerState
}

// NewWriter creates a new response writer
func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer: w,
		state:  stateStart,
	}
}

// WriteStatusLine writes the HTTP status line
func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != stateStart {
		return fmt.Errorf("status line must be written first")
	}
	
	err := WriteStatusLine(w.writer, statusCode)
	if err == nil {
		w.state = stateStatusWritten
	}
	return err
}

// WriteHeaders writes the HTTP headers
func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.state != stateStatusWritten {
		return fmt.Errorf("headers must be written after status line and before body")
	}
	
	err := WriteHeaders(w.writer, headers)
	if err == nil {
		w.state = stateHeadersWritten
	}
	return err
}

// WriteBody writes the response body
func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.state != stateHeadersWritten {
		return 0, fmt.Errorf("body must be written after headers")
	}
	
	n, err := w.writer.Write(p)
	if err == nil {
		w.state = stateBodyWritten
	}
	return n, err
}

// WriteChunkedBody writes a chunk of data using HTTP chunked transfer encoding
func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.state != stateHeadersWritten && w.state != stateChunkedBodyWriting {
		return 0, fmt.Errorf("chunked body must be written after headers")
	}
	
	if len(p) == 0 {
		return 0, nil
	}
	
	// Write chunk size in hexadecimal
	chunkSize := fmt.Sprintf("%x\r\n", len(p))
	_, err := w.writer.Write([]byte(chunkSize))
	if err != nil {
		return 0, err
	}
	
	// Write chunk data
	n, err := w.writer.Write(p)
	if err != nil {
		return n, err
	}
	
	// Write trailing CRLF
	_, err = w.writer.Write([]byte("\r\n"))
	if err != nil {
		return n, err
	}
	
	w.state = stateChunkedBodyWriting
	return n, nil
}

// WriteChunkedBodyDone signals the end of chunked transfer encoding
func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.state != stateChunkedBodyWriting {
		return 0, fmt.Errorf("chunked body done can only be called during chunked transfer")
	}
	
	// Write final chunk (size 0) without ending CRLF if trailers will follow
	_, err := w.writer.Write([]byte("0\r\n"))
	if err != nil {
		return 0, err
	}
	
	w.state = stateChunkedBodyDone
	return 0, nil
}

// WriteTrailers writes HTTP trailers after chunked body
func (w *Writer) WriteTrailers(trailers headers.Headers) error {
	if w.state != stateChunkedBodyDone {
		return fmt.Errorf("trailers can only be written after chunked body is done")
	}
	
	// Write trailers (formatted like headers)
	for key, value := range trailers {
		trailerLine := fmt.Sprintf("%s: %s\r\n", key, value)
		_, err := w.writer.Write([]byte(trailerLine))
		if err != nil {
			return err
		}
	}
	
	// Write final CRLF to end the message
	_, err := w.writer.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	
	w.state = stateTrailersWritten
	return nil
}