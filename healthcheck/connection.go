package healthcheck

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

var DefaultHostsToCheckConnectivity = []string{
	"scanme.sh",
}

type ConnectionInfo struct {
	Host       string
	Successful bool
	Message    string
}

func CheckConnection(host string, port int, protocol string, timeout time.Duration) (*ConnectionInfo, error) {
	address := net.JoinHostPort(host, strconv.Itoa(port))
	conn, err := net.DialTimeout(protocol, address, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	return &ConnectionInfo{
		Host:       host,
		Successful: true,
		Message:    fmt.Sprintf("%s Connect (%s:%v): %s", protocol, host, port, "Successful"),
	}, nil
}

func CheckConnectionsOrDefault(hosts []string, port int, protocol string, timeout time.Duration) ([]ConnectionInfo, error) {
	if len(hosts) == 0 {
		hosts = DefaultHostsToCheckConnectivity
	}

	connectionInfos := []ConnectionInfo{}
	for _, host := range hosts {
		connectivityInfo, err := CheckConnection(host, port, protocol, timeout)
		if err != nil {
			return nil, err
		}
		connectionInfos = append(connectionInfos, *connectivityInfo)
	}
	return connectionInfos, nil
}
