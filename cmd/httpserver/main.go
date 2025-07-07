package main

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

const port = 42069

// myHandler handles HTTP requests with HTML responses
func myHandler(w *response.Writer, req *request.Request) {
	var statusCode response.StatusCode
	var htmlContent string

	// Check if this is a proxy request to httpbin
	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
		handleHttpbinProxy(w, req)
		return
	}
	
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		statusCode = response.StatusBadRequest
		htmlContent = `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`
	case "/myproblem":
		statusCode = response.StatusInternalServerError
		htmlContent = `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`
	default:
		statusCode = response.StatusOK
		htmlContent = `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`
	}

	// Write status line
	w.WriteStatusLine(statusCode)

	// Create headers with HTML content type
	responseHeaders := headers.NewHeaders()
	responseHeaders.Override("Content-Length", strconv.Itoa(len(htmlContent)))
	responseHeaders.Override("Connection", "close")
	responseHeaders.Override("Content-Type", "text/html")

	// Write headers
	w.WriteHeaders(responseHeaders)

	// Write body
	w.WriteBody([]byte(htmlContent))
}

// handleHttpbinProxy proxies requests to httpbin.org with chunked responses
func handleHttpbinProxy(w *response.Writer, req *request.Request) {
	// Extract the path after /httpbin/
	httpbinPath := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")
	if httpbinPath == "" {
		httpbinPath = "/"
	}
	
	// Make request to httpbin.org
	proxyURL := "https://httpbin.org" + httpbinPath
	fmt.Printf("Proxying to: %s\n", proxyURL)
	
	resp, err := http.Get(proxyURL)
	if err != nil {
		// Error making request
		w.WriteStatusLine(response.StatusInternalServerError)
		responseHeaders := headers.NewHeaders()
		errorMsg := "Failed to proxy request"
		responseHeaders.Override("Content-Length", strconv.Itoa(len(errorMsg)))
		responseHeaders.Override("Connection", "close")
		responseHeaders.Override("Content-Type", "text/plain")
		w.WriteHeaders(responseHeaders)
		w.WriteBody([]byte(errorMsg))
		return
	}
	defer resp.Body.Close()
	
	// Write status line (convert from http.Response status code)
	statusCode := response.StatusCode(resp.StatusCode)
	w.WriteStatusLine(statusCode)
	
	// Create headers for chunked response
	responseHeaders := headers.NewHeaders()
	responseHeaders.Override("Transfer-Encoding", "chunked")
	responseHeaders.Override("Connection", "close")
	
	// Copy content type from original response
	if contentType := resp.Header.Get("Content-Type"); contentType != "" {
		responseHeaders.Override("Content-Type", contentType)
	}
	
	// Write headers
	w.WriteHeaders(responseHeaders)
	
	// Stream the response in chunks
	buffer := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			fmt.Printf("Read %d bytes from httpbin.org\n", n)
			w.WriteChunkedBody(buffer[:n])
		}
		
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Error reading from httpbin.org: %v\n", err)
			break
		}
	}
	
	// Signal end of chunked response
	w.WriteChunkedBodyDone()
}

func main() {
	server, err := server.Serve(port, myHandler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
