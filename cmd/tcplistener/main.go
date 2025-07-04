package main

import (
	"fmt"
	"httpfromtcp/internal/request"
	"net"
	"os"
)

func main() {
	ln, err := net.Listen("tcp", ":42069")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to listen: %v\n", err)
		os.Exit(1)
	}
	defer ln.Close()
	fmt.Println("Listening on :42069")

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to accept connection: %v\n", err)
			continue
		}
		fmt.Println("Connection accepted")

		go func(c net.Conn) {
			defer c.Close()

			req, err := request.RequestFromReader(c)
			if err != nil {
				fmt.Printf("Error parsing request: %v\n", err)
				return
			}
			fmt.Println("Request line:")
			fmt.Printf("- Method: %s\n", req.RequestLine.Method)
			fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
			fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)

			fmt.Println("Connection closed")
		}(conn)
	}
}
