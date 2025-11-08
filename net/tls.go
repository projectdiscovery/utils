package net

import (
	"crypto/rand"
	"net"
	"os"
	"time"
)

// DetectTLS attempts to detect TLS on a connection by sending a ClientHello
// and checking for early TLS fingerprints without performing a full handshake.
//
// It returns true if TLS is detected, false otherwise.
//
// The function uses the following steps:
// 1. Generate a random 32-byte value for the ClientHello.
// 2. Create a ServerNameIndication (SNI) extension for the hostname.
// 3. Create a ClientHello message with the random value and the SNI extension.
// 4. Send the ClientHello message to the server.
// 5. Read the response from the server.
// 6. Check if the response contains a ServerHello or tls alert message.
// 7. If the response contains a ServerHello or tls alert message, return true.
// 8. Otherwise, return false.
func DetectTLS(conn net.Conn, host string, timeout time.Duration) bool {
	hostname := ""
	if host != "" {
		if ip := net.ParseIP(host); ip == nil {
			hostname = host
		}
	}

	var sniExtension []byte
	var extensions []byte
	var extensionsLength int

	if hostname != "" {
		hostnameBytes := []byte(hostname)
		sniListLength := 1 + 2 + len(hostnameBytes)
		sniLength := 2 + sniListLength
		sniExtension = make([]byte, 4+sniLength)
		sniExtension[0] = 0x00 // extension type: server_name
		sniExtension[1] = 0x00
		sniExtension[2] = byte(sniLength >> 8) // extension length
		sniExtension[3] = byte(sniLength)
		sniExtension[4] = byte(sniListLength >> 8) // server_name_list length
		sniExtension[5] = byte(sniListLength)
		sniExtension[6] = 0x00                          // name_type: host_name
		sniExtension[7] = byte(len(hostnameBytes) >> 8) // hostname length
		sniExtension[8] = byte(len(hostnameBytes))
		copy(sniExtension[9:], hostnameBytes)

		extensions = sniExtension
		extensionsLength = len(extensions)
	}

	clientHelloBodyLength := 2 + 32 + 1 + 2 + 2 + 1 + 1 + 2 + extensionsLength
	handshakeLength := 4 + clientHelloBodyLength
	recordLength := handshakeLength
	clientHello := make([]byte, 5+handshakeLength)
	offset := 0

	// TLS record header
	clientHello[offset] = 0x16                      // content type: Handshake
	clientHello[offset+1] = 0x03                    // version: TLS 1.0 (major)
	clientHello[offset+2] = 0x01                    // version: TLS 1.0 (minor)
	clientHello[offset+3] = byte(recordLength >> 8) // length (high)
	clientHello[offset+4] = byte(recordLength)      // length (low)
	offset += 5

	// Handshake header
	clientHello[offset] = 0x01                                // handshake type: ClientHello
	clientHello[offset+1] = byte(clientHelloBodyLength >> 16) // length (high)
	clientHello[offset+2] = byte(clientHelloBodyLength >> 8)  // length (mid)
	clientHello[offset+3] = byte(clientHelloBodyLength)       // length (low)
	offset += 4

	// ClientHello message
	clientHello[offset] = 0x03   // version: TLS 1.2 (major)
	clientHello[offset+1] = 0x03 // version: TLS 1.2 (minor)
	offset += 2
	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		return false
	}
	copy(clientHello[offset:], random) // random (32 bytes)
	offset += 32
	clientHello[offset] = 0x00 // session_id length
	offset++
	clientHello[offset] = 0x00   // cipher_suites length (high)
	clientHello[offset+1] = 0x02 // cipher_suites length (low)
	offset += 2
	clientHello[offset] = 0x00   // cipher_suite: TLS_RSA_WITH_AES_128_CBC_SHA (high)
	clientHello[offset+1] = 0x2f // cipher_suite (low)
	offset += 2
	clientHello[offset] = 0x01 // compression_methods length
	offset++
	clientHello[offset] = 0x00 // compression_method: null
	offset++
	clientHello[offset] = byte(extensionsLength >> 8) // extensions length (high)
	clientHello[offset+1] = byte(extensionsLength)    // extensions length (low)
	offset += 2
	if extensionsLength > 0 {
		copy(clientHello[offset:], extensions)
		offset += extensionsLength
	}

	actualRecordLength := offset - 5
	clientHello[3] = byte(actualRecordLength >> 8)
	clientHello[4] = byte(actualRecordLength)

	if err := conn.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
		return false
	}

	if _, err := conn.Write(clientHello[:offset]); err != nil {
		return false
	}
	readTimeout := 2 * time.Second
	if timeout < readTimeout {
		readTimeout = timeout
	}
	if err := conn.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
		return false
	}

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil && !os.IsTimeout(err) {
		return false
	}

	if n < 5 {
		return false
	}

	// check content type (0x15=Alert, 0x16=Handshake)
	contentType := buffer[0]
	if contentType != 0x15 && contentType != 0x16 {
		return false
	}

	// check TLS version (major must be 0x03)
	if buffer[1] != 0x03 {
		return false
	}

	// if handshake, check for ServerHello (0x02)
	if contentType == 0x16 && n >= 6 {
		if buffer[5] == 0x02 {
			return true
		}
	}

	return true
}
