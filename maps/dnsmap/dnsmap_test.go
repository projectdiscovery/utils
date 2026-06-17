package dnsmap

import (
	"net"
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/require"
)

func TestDNSToMap(t *testing.T) {
	msg := dns.Msg{}
	msg.Rcode = 1
	msg.Question = []dns.Question{{Name: "test", Qtype: 1, Qclass: 1}}
	msg.Extra = []dns.RR{&dns.A{Hdr: dns.RR_Header{Name: "test", Rrtype: 1, Class: 1, Ttl: 1}, A: net.ParseIP("0.0.0.0")}}
	msg.Answer = []dns.RR{&dns.A{Hdr: dns.RR_Header{Name: "test", Rrtype: 1, Class: 1, Ttl: 1}, A: net.ParseIP("0.0.0.0")}}
	msg.Ns = []dns.RR{&dns.A{Hdr: dns.RR_Header{Name: "test", Rrtype: 1, Class: 1, Ttl: 1}, A: net.ParseIP("0.0.0.0")}}
	m := DNSToMap(&msg, "")
	require.NotNil(t, m)
	require.NotEmpty(t, m)
}
