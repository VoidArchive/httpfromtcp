package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	// Step 1: Figure out where we're throwing our paper airplanes
	// This is like getting the address of your friend's house
	addr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatalf("failed to resolve UDP address: %v", err)
	}

	// Step 2: Get ready to throw! This is like opening your window
	// Notice: UDP "connection" isn't really a connection - it's just preparation
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatalf("failed to dial UDP: %v", err)
	}
	defer conn.Close() // Clean up when we're done

	// Step 3: Set up our message reader (like having a notepad ready)
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("UDP paper airplane launcher ready! Type messages and hit Enter.")
	fmt.Println("(The receiver might not be listening, but we'll throw anyway!)")

	// Step 4: The infinite paper airplane throwing loop
	for {
		// Show we're ready for input
		fmt.Print("> ")

		// Read what the user wants to send
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("error reading input: %v", err)
			continue
		}

		// Throw the paper airplane! 
		// Unlike TCP, we don't know if anyone caught it
		_, err = conn.Write([]byte(line))
		if err != nil {
			log.Printf("error sending UDP packet: %v", err)
		}
	}
}