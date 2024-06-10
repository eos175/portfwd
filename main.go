package main

import (
	"io"
	"net"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	flag "github.com/spf13/pflag"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Define and parse command-line arguments
	var listenAddr, proxyAddr string
	var useProxyProtocol bool
	flag.StringVarP(&listenAddr, "listen", "l", ":8080", "Address to listen for connections")
	flag.StringVarP(&proxyAddr, "proxy", "p", "127.0.0.1:80", "Proxy server address")
	flag.BoolVarP(&useProxyProtocol, "use-proxy-protocol", "u", false, "Enable Proxy Protocol v2")
	flag.Parse()

	// Start listening on the specified address
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatal().Err(err).Str("address", listenAddr).Msg("Failed to start listening")
	}
	defer ln.Close()

	log.Info().
		Str("listen_address", listenAddr).
		Str("proxy_address", proxyAddr).
		Msg("Listening and proxying")

	for {
		// Accept incoming connections
		conn, err := ln.Accept()
		if err != nil {
			log.Error().Err(err).Msg("Error accepting incoming connection")
			continue
		}

		// Handle the connection in a goroutine
		go handleRequest(conn, proxyAddr, useProxyProtocol)
	}
}

func handleRequest(conn net.Conn, proxyAddr string, useProxyProtocol bool) {
	defer conn.Close()

	_ = useProxyProtocol

	// Connect to the proxy server
	proxy, err := net.Dial("tcp", proxyAddr)
	if err != nil {
		log.Error().Err(err).Str("proxy_address", proxyAddr).Msg("Failed to connect to proxy server")
		return
	}
	defer proxy.Close()

	log.Info().
		Str("client_ip", conn.RemoteAddr().String()).
		Str("proxy_ip", proxy.LocalAddr().String()).
		Str("proxy_address", proxyAddr).Msg("New client connection")

	// Use a channel to synchronize the completion of goroutines
	done := make(chan struct{})

	var clientBytes, proxyBytes int64

	// Copy data from incoming connection to the proxy
	go func() {
		defer func() {
			done <- struct{}{}
			proxy.Close()
		}()
		n, err := io.Copy(proxy, conn)
		if err != nil {
			log.Error().Err(err).Msg("Error copying data from incoming connection to proxy")
		}
		clientBytes = n
	}()

	// Copy data from proxy to the incoming connection
	go func() {
		defer func() {
			done <- struct{}{}
			conn.Close()
		}()
		n, err := io.Copy(conn, proxy)
		if err != nil {
			log.Error().Err(err).Msg("Error copying data from proxy to incoming connection")
		}
		proxyBytes = n
	}()

	// Wait for both goroutines to finish
	<-done
	<-done

	log.Info().
		Int64("bytes_written_to_client", clientBytes).
		Int64("bytes_written_to_proxy", proxyBytes).
		Msg("Client connection closed")

}
