package main

import (
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

const port = 42069

// myHandler handles HTTP requests with HTML responses
func myHandler(w *response.Writer, req *request.Request) {
	var statusCode response.StatusCode
	var htmlContent string

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
