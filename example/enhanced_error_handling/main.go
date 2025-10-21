package main

import (
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/yutopp/go-rtmp"
)

func main() {
	// Create enhanced handler with error handling
	handler := NewEnhancedHandler()
	
	// Create server with enhanced configuration
	server := rtmp.NewServer(&rtmp.ServerConfig{
		OnConnect: func(conn net.Conn) (io.ReadWriteCloser, *rtmp.ConnConfig) {
			log.Printf("New connection from: %s", conn.RemoteAddr())
			
			// Return enhanced connection config
			return conn, &rtmp.ConnConfig{
				Handler: handler,
				Logger:  logrus.New(),
			}
		},
	})
	
	// Start server
	addr := ":1935"
	log.Printf("Starting enhanced RTMP server on %s", addr)
	
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	
	go func() {
		if err := server.Serve(listener); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()
	
	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("Shutting down server...")
	
	if err := server.Close(); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
	
	log.Println("Server stopped")
}
