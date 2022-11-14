package iputil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsIP(t *testing.T) {
	require.True(t, IsIP("127.0.0.1"), "valid ip not recognized")
	require.False(t, IsIP("127.0.0.0/24"), "cidr recognized as ip")
	require.False(t, IsIP("test"), "string recognized as ip")
}

func TestIsInternalIPv4(t *testing.T) {
	// Test this ipv4
	require.False(t, IsInternal("153.12.14.1"), "internal ipv4 address recognized as not valid")
	require.True(t, IsInternal("172.16.0.0"), "internal ipv4 address recognized as valid")
	// Test with ipv6
	require.False(t, IsInternal("684D:1111:222:3333:4444:5555:6:77"), "internal ipv4 address recognized as not valid")
	require.True(t, IsInternal("fc00:7e5b:cfa9::"), "internal ipv4 address recognized as valid")
}

func TestIsPort(t *testing.T) {
	require.False(t, IsPort("0"), "invalid port 0")
	require.False(t, IsPort("-1"), "negative port")
	require.True(t, IsPort("1"), "valid port not recognized")
	require.True(t, IsPort("65535"), "valid port not recognized")
	require.False(t, IsPort("65536"), "valid port not recognized")
	require.False(t, IsPort("0xff"), "hex port considered valid")
	require.False(t, IsPort("12.12"), "float recognized as valid")
}

func TestIsIPv4(t *testing.T) {
	require.True(t, IsIPv4("127.0.0.1"), "valid ipv4 address not recognized")
	require.False(t, IsIPv4("2001:0db8:85a3:0000:0000:8a2e:0370:7334"), "ipv6 address recognized as valid")
}

func TestIsIPv6(t *testing.T) {
	require.False(t, IsIPv6("127.0.0.1"), "ipv4 address recognized as valid")
	require.True(t, IsIPv6("2001:0db8:85a3:0000:0000:8a2e:0370:7334"), "valid ipv6 address not recognized")
}

func TestIsCIDR(t *testing.T) {
	require.False(t, IsCIDR("127.0.0.1"), "ipv4 address recognized as cidr")
	require.True(t, IsCIDR("127.0.0.0/24"), "valid cidr not recognized")
	require.True(t, IsCIDR("127.0.0.0/1"), "valid cidr not recognized")
	require.True(t, IsCIDR("127.0.0.0/32"), "valid cidr not recognized")
	require.False(t, IsCIDR("2001:0db8:85a3:0000:0000:8a2e:0370:7334"), "ipv6 address recognized as cidr")
}

func TestIsCidrWithExpansion(t *testing.T) {
	require.True(t, IsCidrWithExpansion("127.0.0.1-32"), "valid cidr /32 not recognized")
	require.False(t, IsCidrWithExpansion("127.0.0.0-55"), "invalid cidr /55")
}

func TestToCidr(t *testing.T) {
	tests := map[string]bool{
		"127.0.0.0/24": true,
		"127.0.0.1":    true,
		"aaa":          false,
	}
	for item, ok := range tests {
		tocidr := ToCidr(item)
		if ok {
			require.NotNil(t, tocidr, "valid cidr not recognized")
		} else {
			require.Nil(t, tocidr, "invalid cidr")
		}
	}
}

func TestAsIPV4IpNet(t *testing.T) {
	tests := map[string]bool{
		"127.0.0.0/24": true,
		"127.0.0.1":    true,
		"aaa":          false,
		"2001:0db8:85a3:0000:0000:8a2e:0370:7334": false,
	}
	for item, ok := range tests {
		tocidr := AsIPV4IpNet(item)
		if ok {
			require.NotNil(t, tocidr, "valid cidr not recognized")
		} else {
			require.Nil(t, tocidr, "invalid cidr")
		}
	}
}

func TestAsIPV6IpNet(t *testing.T) {
	tests := map[string]bool{
		"2001:0db8:85a3:0000:0000:8a2e:0370:7334": true,
		"2002::1234:abcd:ffff:c0a8:101/64":        true,
	}
	for item, ok := range tests {
		tocidr := AsIPV6CIDR(item)
		if ok {
			require.NotNil(t, tocidr, "valid cidr not recognized")
		} else {
			require.Nil(t, tocidr, "invalid cidr")
		}
	}
}

func TestWhatsMyIP(t *testing.T) {
	// we can't compare the ip with local interfaces as it might be the external gateway one
	// so we just verify we can contact the api endpoint
	_, err := WhatsMyIP()
	require.Nil(t, err, "couldn't retrieve ip")
}
