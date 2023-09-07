package netutil

import "net"

// TryJoinHostPort joins host and port. If port is empty, it returns host and an error.
func TryJoinHostPort(host, port string) (string, error) {
	if host == "" {
		return "", &net.AddrError{Err: "missing host", Addr: host}
	}

	if port == "" {
		return host, &net.AddrError{Err: "missing port", Addr: host}
	}

	return net.JoinHostPort(host, port), nil
}
