// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package networking

import (
	"errors"
	"net"

	stream "github.com/nknorg/encrypted-stream"
)

func strto32bytes(s string) *[32]byte {
	if len(s) != 32 {
		return nil
	}

	b := [32]byte{}
	for i := 0; i < 32; i++ {
		b[i] = s[i]
	}
	return &b
}

func connectToHost(host, joinKey string) (net.Conn, error) {
	// Get the join key as a byte array.
	joinKeyBytes := strto32bytes(joinKey)
	if joinKeyBytes == nil {
		return nil, errors.New("join key is not valid")
	}

	// Make the network connection.
	c, err := net.Dial("tcp", host)
	if err != nil {
		return nil, err
	}

	// Create a encrypted socket.
	encryptedConn, err := stream.NewEncryptedStream(c, &stream.Config{
		Cipher:    stream.NewXSalsa20Poly1305Cipher(joinKeyBytes),
		Initiator: true,
	})
	if err != nil {
		return nil, err
	}

	// Return the encrypted socket.
	return encryptedConn, nil
}

func acceptHostConnections(joinKey string, ln net.Listener, hn func(net.Conn)) {
	// Get the join key as a byte array.
	joinKeyBytes := strto32bytes(joinKey)
	if joinKeyBytes == nil {
		return
	}

	// Accept connections.
	for {
		// Accept a connection.
		c, err := ln.Accept()
		if err != nil {
			return
		}

		go func() {
			// Defer the close.
			defer c.Close()

			// Create a encrypted socket.
			encryptedConn, err := stream.NewEncryptedStream(c, &stream.Config{
				Cipher:    stream.NewXSalsa20Poly1305Cipher(joinKeyBytes),
				Initiator: false,
			})
			if err != nil {
				return
			}

			// Handle the connection.
			hn(encryptedConn)
		}()
	}
}
