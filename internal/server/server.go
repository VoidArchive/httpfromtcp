package server

import (
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"net"
	"sync/atomic"
)

// Handler function type that processes HTTP requests
type Handler func(w *response.Writer, req *request.Request)

// Server represents an HTTP server
type Server struct {
	listener net.Listener
	handler  Handler
	closed   atomic.Bool
}

// Serve creates a new server and starts listening on the given port
func Serve(port int, handler Handler) (*Server, error) {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %w", port, err)
	}

	server := &Server{
		listener: listener,
		handler:  handler,
	}

	// Start listening in a background goroutine
	go server.listen()

	return server, nil
}

// Close stops the server and closes the listener
func (s *Server) Close() error {
	s.closed.Store(true)
	return s.listener.Close()
}

// listen accepts incoming connections and handles them
func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			// If server is closed, ignore connection errors
			if s.closed.Load() {
				return
			}
			// TODO: Log error in production
			continue
		}

		// Handle each connection in a separate goroutine
		go s.handle(conn)
	}
}

// handle processes a single connection
func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	// Parse the request from the connection
	req, err := request.RequestFromReader(conn)
	if err != nil {
		// If parsing fails, return 400 Bad Request using response.Writer
		writer := response.NewWriter(conn)
		writer.WriteStatusLine(response.StatusBadRequest)
		headers := response.GetDefaultHeaders(len("Bad Request\n"))
		writer.WriteHeaders(headers)
		writer.WriteBody([]byte("Bad Request\n"))
		return
	}

	// Create a response writer for the handler
	writer := response.NewWriter(conn)

	// Call the handler function
	s.handler(writer, req)
}
