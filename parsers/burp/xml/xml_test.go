package burpxml

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const testXML = `<?xml version="1.0"?>
<items>
  <item>
    <time>Mon Jan 01 12:00:00 UTC 2024</time>
    <url>https://example.com/api/users</url>
    <host ip="93.184.216.34">example.com</host>
    <port>443</port>
    <protocol>https</protocol>
    <path>/api/users</path>
    <extension>json</extension>
    <request base64="false">GET /api/users HTTP/1.1
Host: example.com

</request>
    <status>200</status>
    <responselength>1234</responselength>
    <mimetype>application/json</mimetype>
    <response base64="false">HTTP/1.1 200 OK
Content-Type: application/json

{"users":[]}</response>
    <comment>Test request</comment>
  </item>
</items>`

func TestParseXML(t *testing.T) {
	items, err := ParseXML(strings.NewReader(testXML), XMLParseOptions{})
	require.NoError(t, err)
	require.Len(t, items.Items, 1)

	item := items.Items[0]
	require.Equal(t, "Mon Jan 01 12:00:00 UTC 2024", item.Time)
	require.Equal(t, "https://example.com/api/users", item.URL)
	require.Equal(t, "example.com", item.Host.Name)
	require.Equal(t, "93.184.216.34", item.Host.IP)
	require.Equal(t, "443", item.Port)
	require.Equal(t, "https", item.Protocol)
	require.Equal(t, "/api/users", item.Path)
	require.Equal(t, "json", item.Extension)
	require.Equal(t, "200", item.Status)
	require.Equal(t, "1234", item.ResponseLength)
	require.Equal(t, "application/json", item.MimeType)
	require.Equal(t, "Test request", item.Comment)
	require.Contains(t, item.Request.Raw, "GET /api/users HTTP/1.1")
	require.Contains(t, item.Response.Raw, "HTTP/1.1 200 OK")
}

func TestParseXMLWithBase64Decode(t *testing.T) {
	reqBody := "GET /secret HTTP/1.1\nHost: test.com\n"
	respBody := "HTTP/1.1 200 OK\n\nSecret data"

	xml := `<?xml version="1.0"?>
<items>
  <item>
    <time>Mon Jan 01 12:00:00 UTC 2024</time>
    <url>https://test.com/secret</url>
    <host ip="1.2.3.4">test.com</host>
    <port>443</port>
    <protocol>https</protocol>
    <path>/secret</path>
    <extension></extension>
    <request base64="true">` + base64.StdEncoding.EncodeToString([]byte(reqBody)) + `</request>
    <status>200</status>
    <responselength>100</responselength>
    <mimetype>text/plain</mimetype>
    <response base64="true">` + base64.StdEncoding.EncodeToString([]byte(respBody)) + `</response>
    <comment></comment>
  </item>
</items>`

	items, err := ParseXML(strings.NewReader(xml), XMLParseOptions{DecodeBase64: true})
	require.NoError(t, err)
	require.Len(t, items.Items, 1)

	item := items.Items[0]
	require.Equal(t, reqBody, item.Request.Body)
	require.Equal(t, respBody, item.Response.Body)
}

func TestParseXMLInvalidXML(t *testing.T) {
	_, err := ParseXML(strings.NewReader("<invalid"), XMLParseOptions{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "xml decode")
}

func TestParseXMLInvalidBase64(t *testing.T) {
	xml := `<?xml version="1.0"?>
<items>
  <item>
    <time></time>
    <url></url>
    <host ip=""></host>
    <port></port>
    <protocol></protocol>
    <path></path>
    <extension></extension>
    <request base64="true">not-valid-base64!!!</request>
    <status></status>
    <responselength></responselength>
    <mimetype></mimetype>
    <response base64="false"></response>
    <comment></comment>
  </item>
</items>`

	_, err := ParseXML(strings.NewReader(xml), XMLParseOptions{DecodeBase64: true})
	require.Error(t, err)
	require.Contains(t, err.Error(), "base64 decode")
}

func TestParseXMLMultipleItems(t *testing.T) {
	xml := `<?xml version="1.0"?>
<items>
  <item>
    <time>Time1</time>
    <url>https://a.com</url>
    <host ip="1.1.1.1">a.com</host>
    <port>443</port>
    <protocol>https</protocol>
    <path>/a</path>
    <extension></extension>
    <request base64="false">req1</request>
    <status>200</status>
    <responselength>10</responselength>
    <mimetype>text/html</mimetype>
    <response base64="false">resp1</response>
    <comment></comment>
  </item>
  <item>
    <time>Time2</time>
    <url>https://b.com</url>
    <host ip="2.2.2.2">b.com</host>
    <port>80</port>
    <protocol>http</protocol>
    <path>/b</path>
    <extension></extension>
    <request base64="false">req2</request>
    <status>404</status>
    <responselength>20</responselength>
    <mimetype>text/plain</mimetype>
    <response base64="false">resp2</response>
    <comment></comment>
  </item>
</items>`

	items, err := ParseXML(strings.NewReader(xml), XMLParseOptions{})
	require.NoError(t, err)
	require.Len(t, items.Items, 2)

	require.Equal(t, "https://a.com", items.Items[0].URL)
	require.Equal(t, "https://b.com", items.Items[1].URL)
	require.Equal(t, "200", items.Items[0].Status)
	require.Equal(t, "404", items.Items[1].Status)
}

func TestParseXMLEmptyItems(t *testing.T) {
	xml := `<?xml version="1.0"?><items></items>`
	items, err := ParseXML(strings.NewReader(xml), XMLParseOptions{})
	require.NoError(t, err)
	require.Empty(t, items.Items)
}

func TestItemsToJSON(t *testing.T) {
	items := &Items{
		Items: []Item{
			{
				URL:      "https://example.com",
				Host:     Host{Name: "example.com", IP: "1.2.3.4"},
				Port:     "443",
				Protocol: "https",
				Status:   "200",
			},
		},
	}

	var buf bytes.Buffer
	err := items.ToJSON(&buf)
	require.NoError(t, err)

	json := buf.String()
	require.Contains(t, json, `"url": "https://example.com"`)
	require.Contains(t, json, `"status": "200"`)
}

func TestItemsToCSV(t *testing.T) {
	items := &Items{
		Items: []Item{
			{
				Time:           "Time1",
				URL:            "https://example.com",
				Host:           Host{Name: "example.com", IP: "1.2.3.4"},
				Port:           "443",
				Protocol:       "https",
				Path:           "/path",
				Extension:      "html",
				Request:        Request{Raw: "GET / HTTP/1.1"},
				Status:         "200",
				ResponseLength: "100",
				MimeType:       "text/html",
				Response:       Response{Raw: "HTTP/1.1 200 OK"},
				Comment:        "test",
			},
		},
	}

	var buf bytes.Buffer
	err := items.ToCSV(&buf, CSVOptions{})
	require.NoError(t, err)

	csv := buf.String()
	require.Contains(t, csv, "Time1")
	require.Contains(t, csv, "https://example.com")
	require.Contains(t, csv, "GET / HTTP/1.1")
}

func TestItemsToCSVExcludeRequestResponse(t *testing.T) {
	items := &Items{
		Items: []Item{
			{
				URL:      "https://example.com",
				Request:  Request{Raw: "SECRET_REQUEST"},
				Response: Response{Raw: "SECRET_RESPONSE"},
			},
		},
	}

	var buf bytes.Buffer
	err := items.ToCSV(&buf, CSVOptions{ExcludeRequest: true, ExcludeResponse: true})
	require.NoError(t, err)

	csv := buf.String()
	require.NotContains(t, csv, "SECRET_REQUEST")
	require.NotContains(t, csv, "SECRET_RESPONSE")
}

func TestRequestContent(t *testing.T) {
	t.Run("returns body when set", func(t *testing.T) {
		r := &Request{Raw: "raw", Body: "decoded"}
		require.Equal(t, "decoded", r.content())
	})

	t.Run("returns raw when body empty", func(t *testing.T) {
		r := &Request{Raw: "raw"}
		require.Equal(t, "raw", r.content())
	})
}

func TestResponseContent(t *testing.T) {
	t.Run("returns body when set", func(t *testing.T) {
		r := &Response{Raw: "raw", Body: "decoded"}
		require.Equal(t, "decoded", r.content())
	})

	t.Run("returns raw when body empty", func(t *testing.T) {
		r := &Response{Raw: "raw"}
		require.Equal(t, "raw", r.content())
	})
}

func TestItemString(t *testing.T) {
	item := Item{URL: "https://example.com", Host: Host{Name: "example.com"}, Status: "200"}
	s := item.String()
	require.Contains(t, s, "https://example.com")
	require.Contains(t, s, "example.com")
	require.Contains(t, s, "200")
	require.Contains(t, s, "Item{")
}

func TestRequestString(t *testing.T) {
	t.Run("shows body when decoded", func(t *testing.T) {
		r := Request{Raw: "raw", Body: "decoded body"}
		s := r.String()
		require.Contains(t, s, "Request{")
		require.Contains(t, s, "Body = decoded body")
	})

	t.Run("shows base64 and raw when not decoded", func(t *testing.T) {
		r := Request{Base64Encoded: "true", Raw: "raw content"}
		s := r.String()
		require.Contains(t, s, "Request{")
		require.Contains(t, s, "Base64")
		require.Contains(t, s, "raw content")
	})
}

func TestResponseString(t *testing.T) {
	t.Run("shows body when decoded", func(t *testing.T) {
		r := Response{Raw: "raw", Body: "decoded body"}
		s := r.String()
		require.Contains(t, s, "Response{")
		require.Contains(t, s, "Body = decoded body")
	})

	t.Run("shows base64 and raw when not decoded", func(t *testing.T) {
		r := Response{Base64Encoded: "false", Raw: "raw content"}
		s := r.String()
		require.Contains(t, s, "Response{")
		require.Contains(t, s, "Base64")
		require.Contains(t, s, "raw content")
	})
}

func TestItemToStrings(t *testing.T) {
	item := Item{
		Time:           "time",
		URL:            "url",
		Host:           Host{Name: "host", IP: "ip"},
		Port:           "port",
		Protocol:       "proto",
		Path:           "path",
		Extension:      "ext",
		Request:        Request{Body: "req"},
		Status:         "200",
		ResponseLength: "100",
		MimeType:       "text",
		Response:       Response{Body: "resp"},
		Comment:        "comment",
	}

	strs := item.ToStrings(false, false)
	require.Contains(t, strs, "time")
	require.Contains(t, strs, "url")
	require.Contains(t, strs, "host")
	require.Contains(t, strs, "req")
	require.Contains(t, strs, "resp")
}

func TestItemFlatString(t *testing.T) {
	item := Item{URL: "https://example.com", Status: "200"}
	s := item.FlatString()
	require.Contains(t, s, "https://example.com")
	require.Contains(t, s, "200")
}
