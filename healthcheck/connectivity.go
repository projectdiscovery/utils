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

type ConnectivityInfo struct {
	Host       string
	Successful bool
	Message    string
}

func CheckConnection(host string, port int, protocol string, timeout time.Duration) (*ConnectivityInfo, error) {
	address := net.JoinHostPort(host, strconv.Itoa(port))
	conn, err := net.DialTimeout(protocol, address, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	return &ConnectivityInfo{
		Host:       host,
		Successful: true,
		Message:    fmt.Sprintf("%s IPv4 Connect (%s:%v): %s", protocol, host, port, "Successful"),
	}, nil
}

func CheckConnectionsOrDefault(hosts []string, port int, protocol string, timeout time.Duration) ([]ConnectivityInfo, error) {
	if len(hosts) == 0 {
		hosts = DefaultHostsToCheckConnectivity
	}

	connectivityInfos := []ConnectivityInfo{}
	for _, host := range hosts {
		connectivityInfo, err := CheckConnection(host, port, protocol, timeout)
		if err != nil {
			return nil, err
		}
		connectivityInfos = append(connectivityInfos, *connectivityInfo)
	}
	return connectivityInfos, nil
}
