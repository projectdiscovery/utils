package urlutil

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/require"
)

func TestUTF8URLEncoding(t *testing.T) {
	exstring := '上'
	expected := `e4%b8%8a`
	val := getutf8hex(exstring)
	require.Equalf(t, val, expected, "failed to url encode utf char expected %v but got %v", expected, val)
}

func TestParamEncoding(t *testing.T) {
	testcases := []struct {
		Payload  string
		Expected string
	}{
		{"1+AND+(SELECT+*+FROM+(SELECT(SLEEP(12)))nQIP)", "1+AND+(SELECT+*+FROM+(SELECT(SLEEP(12)))nQIP)"},
		{"1 AND SELECT", "1+AND+SELECT"},
	}
	for _, v := range testcases {
		val := ParamEncode(v.Payload)
		require.Equalf(t, val, v.Expected, "failed to url encode payload expected %v got %v", v.Expected, val)
	}
}

func TestRawParam(t *testing.T) {
	p := NewParams()
	p.Add("sqli", "1+AND+(SELECT+*+FROM+(SELECT(SLEEP(12)))nQIP)")
	p.Add("xss", "<script>alert('XSS')</script>")
	p.Add("xssiwthspace", "<svg id=alert(1) onload=eval(id)>")
	p.Add("jsprotocol", "javascript://alert(1)")
	// Note keys are sorted
	expected := "jsprotocol=javascript://alert(1)&sqli=1+AND+(SELECT+*+FROM+(SELECT(SLEEP(12)))nQIP)&xss=<script>alert('XSS')</script>&xssiwthspace=<svg+id=alert(1)+onload=eval(id)>"
	require.Equalf(t, p.Encode(), expected, "failed to encode parameters expected %v but got %v", expected, p.Encode())
}

func TestParamIntegration(t *testing.T) {
	var routerErr error
	expected := "/params?jsprotocol=javascript://alert(1)&sqli=1+AND+(SELECT+*+FROM+(SELECT(SLEEP(12)))nQIP)&xss=<script>alert('XSS')</script>&xssiwthspace=<svg+id=alert(1)+onload=eval(id)>"

	router := httprouter.New()
	router.GET("/params", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		if r.RequestURI != expected {
			routerErr = fmt.Errorf("expected %v but got %v", expected, r.RequestURI)
		}
		w.WriteHeader(http.StatusOK)
	})
	ts := httptest.NewServer(router)
	defer ts.Close()

	p := NewParams()
	p.Add("sqli", "1+AND+(SELECT+*+FROM+(SELECT(SLEEP(12)))nQIP)")
	p.Add("xss", "<script>alert('XSS')</script>")
	p.Add("xssiwthspace", "<svg id=alert(1) onload=eval(id)>")
	p.Add("jsprotocol", "javascript://alert(1)")

	url, _ := url.Parse(ts.URL + "/params")
	url.RawQuery = p.Encode()
	_, err := http.Get(url.String())
	require.Nil(t, err)
	require.Nil(t, routerErr)
}

func TestPercentEncoding(t *testing.T) {
	// From Burpsuite
	expected := "%74%65%73%74%20%26%23%20%28%29%20%70%65%72%63%65%6E%74%20%5B%5D%7C%2A%20%65%6E%63%6F%64%69%6E%67"
	payload := "test &# () percent []|* encoding"
	value := PercentEncoding(payload)
	require.Equalf(t, value, expected, "expected percentencoding to be %v but got %v", expected, value)
	decoded, err := url.QueryUnescape(value)
	require.Nil(t, err)
	require.Equal(t, payload, decoded)
}

func TestGetParams(t *testing.T) {
	values := url.Values{}
	values.Add("sqli", "1+AND+(SELECT+*+FROM+(SELECT(SLEEP(12)))nQIP)")
	values.Add("xss", "<script>alert('XSS')</script>")
	p := GetParams(values)
	require.NotNilf(t, p, "expected params but got nil")
	require.Equalf(t, p.Get("sqli"), values.Get("sqli"), "malformed or missing value for param sqli expected %v but got %v", values.Get("sqli"), p.Get("sqli"))
	require.Equalf(t, p.Get("xss"), values.Get("xss"), "malformed or missing value for param xss expected %v but got %v", values.Get("xss"), p.Get("xss"))
}

func TestURLEncode(t *testing.T) {
	example := "\r\n"
	got := URLEncodeWithEscapes(example)
	require.Equalf(t, "%0D%0A", got, "failed to url encode characters")

	// verify with stdlib
	for r := 0; r < 20; r++ {
		expected := url.QueryEscape(string(rune(r)))
		got := URLEncodeWithEscapes(string(rune(r)))
		require.Equalf(t, expected, got, "url encoding mismatch for non-printable char with ascii val:%v", r)
	}
}

func TestURLDecode(t *testing.T) {
	testcases := []struct {
		url      string
		Expected string
	}{
		{
			"/ctc/servlet/ConfigServlet?param=com.sap.ctc.util.FileSystemConfig;EXECUTE_CMD;CMDLINE=tasklist",
			"param=com.sap.ctc.util.FileSystemConfig;EXECUTE_CMD;CMDLINE=tasklist",
		},
	}
	for _, v := range testcases {
		parsed, err := Parse(v.url)
		require.Nilf(t, err, "failed to parse url %v", v.url)
		require.Equalf(t, v.Expected, parsed.Query().Encode(), "failed to decode params in url %v expected %v got %v", v.url, v.Expected, parsed.Query())
	}
}

func TestPathEncode(t *testing.T) {
	testcases := []struct {
		Input    string
		Expected string
		Desc     string
	}{
		// Space encoding - always %20 in paths
		{"hello world", "hello%20world", "spaces encoded as %20"},
		{"test+value", "test+value", "+ preserved as literal"},
		
		// Special characters that need escaping in paths
		{"path?query", "path%3Fquery", "? must be escaped in paths"},
		{"path#fragment", "path%23fragment", "# must be escaped in paths"},
		{"user@domain", "user%40domain", "@ must be escaped"},
		
		// Characters that don't need escaping in paths (unlike query params)
		{"key=value", "key=value", "= is literal in paths"},
		{"param&other", "param&other", "& is literal in paths"},
		
		// Control characters
		{"test\nline", "test%0Aline", "newline encoded"},
		{"test\tline", "test%09line", "tab encoded"},
		
		// Non-ASCII characters
		{"café", "caf%c3%a9", "unicode encoded"},
		
		// Edge cases
		{"", "", "empty string"},
		{"/", "/", "forward slash preserved"},
		{"../../../etc/passwd", "../../../etc/passwd", "path traversal sequences preserved"},
	}
	
	for _, v := range testcases {
		result := PathEncode(v.Input)
		require.Equalf(t, v.Expected, result, "%s: expected %q but got %q", v.Desc, v.Expected, result)
	}
}

func TestPathDecode(t *testing.T) {
	testcases := []struct {
		Input    string
		Expected string
		Desc     string
	}{
		// Space decoding - only %20 becomes space
		{"hello%20world", "hello world", "%20 decoded to space"},
		{"test+value", "test+value", "+ preserved as literal (not decoded to space)"},
		
		// Hex decoding
		{"path%3Fquery", "path?query", "? decoded"},
		{"path%23fragment", "path#fragment", "# decoded"},
		{"user%40domain", "user@domain", "@ decoded"},
		
		// Characters that don't need decoding
		{"key=value", "key=value", "= preserved"},
		{"param&other", "param&other", "& preserved"},
		
		// Control characters
		{"test%0Aline", "test\nline", "newline decoded"},
		{"test%09line", "test\tline", "tab decoded"},
		
		// Non-ASCII
		{"caf%C3%A9", "café", "unicode decoded"},
		
		// Invalid sequences should be preserved
		{"test%GG", "test%GG", "invalid hex preserved"},
		{"test%2", "test%2", "incomplete hex preserved"},
		
		// Edge cases
		{"", "", "empty string"},
		{"/", "/", "forward slash preserved"},
		{"../../../etc/passwd", "../../../etc/passwd", "path traversal preserved"},
	}
	
	for _, v := range testcases {
		result, err := PathDecode(v.Input)
		require.Nilf(t, err, "%s: unexpected error: %v", v.Desc, err)
		require.Equalf(t, v.Expected, result, "%s: expected %q but got %q", v.Desc, v.Expected, result)
	}
}

func TestPathEncodeDecodeRoundtrip(t *testing.T) {
	testcases := []string{
		"hello world",
		"path?query#fragment",
		"user@domain.com",
		"key=value&param=other",
		"test\nwith\tcontrol\rchars",
		"café with unicode",
		"../../../etc/passwd",
		"test+literal+plus",
	}
	
	for _, input := range testcases {
		encoded := PathEncode(input)
		decoded, err := PathDecode(encoded)
		require.Nilf(t, err, "decode error for input %q", input)
		require.Equalf(t, input, decoded, "roundtrip failed for %q: encoded=%q decoded=%q", input, encoded, decoded)
	}
}

func TestPathVsParamEncodingDifferences(t *testing.T) {
	testcases := []struct {
		Input            string
		ExpectedPath     string
		ExpectedParam    string
		Desc             string
	}{
		// Key difference: space encoding
		{"hello world", "hello%20world", "hello+world", "space encoding difference"},
		
		// + character handling
		{"test+plus", "test+plus", "test+plus", "+ preserved in both"},
		
		// & and = handling
		{"key=val&other=test", "key=val&other=test", "key=val&other=test", "& and = preserved in both by default"},
		
		// ? and # handling  
		{"query?test#frag", "query%3Ftest%23frag", "query?test#frag", "? and # encoded only in paths"},
	}
	
	for _, v := range testcases {
		pathResult := PathEncode(v.Input)
		paramResult := ParamEncode(v.Input)
		
		require.Equalf(t, v.ExpectedPath, pathResult, "%s: path encoding mismatch", v.Desc)
		require.Equalf(t, v.ExpectedParam, paramResult, "%s: param encoding mismatch", v.Desc)
	}
}

func TestSQLInjectionPathEncoding(t *testing.T) {
	testcases := []struct {
		Name             string
		Input            string
		ExpectedEncoded  string
		ExpectedDecoded  string
		Description      string
	}{
		{
			Name:            "SQL injection in path with mixed encoding",
			Input:           "/admin/1' OR 1=1 ?key=y'+1=1&key2=value2",
			ExpectedEncoded: "/admin/1'%20OR%201=1%20%3Fkey=y'+1=1&key2=value2",
			ExpectedDecoded: "/admin/1' OR 1=1 ?key=y'+1=1&key2=value2",
			Description:     "SQL injection path with spaces, quotes, and query-like syntax",
		},
		{
			Name:            "Path with SQL payload and question mark",
			Input:           "/user/1' OR 1=1?admin=true",
			ExpectedEncoded: "/user/1'%20OR%201=1%3Fadmin=true",
			ExpectedDecoded: "/user/1' OR 1=1?admin=true",
			Description:     "SQL injection with question mark that needs encoding in paths",
		},
		{
			Name:            "Complex SQL injection with multiple special chars",
			Input:           "/api/user/1' UNION SELECT * FROM users WHERE admin=1#comment",
			ExpectedEncoded: "/api/user/1'%20UNION%20SELECT%20*%20FROM%20users%20WHERE%20admin=1%23comment",
			ExpectedDecoded: "/api/user/1' UNION SELECT * FROM users WHERE admin=1#comment",
			Description:     "Complex SQL injection with spaces and hash that need encoding",
		},
		{
			Name:            "Path traversal with SQL injection",
			Input:           "/../../../etc/passwd' OR '1'='1",
			ExpectedEncoded: "/../../../etc/passwd'%20OR%20'1'='1",
			ExpectedDecoded: "/../../../etc/passwd' OR '1'='1",
			Description:     "Path traversal combined with SQL injection",
		},
		{
			Name:            "Already encoded SQL injection",
			Input:           "/admin/1' OR 1=1 --",
			ExpectedEncoded: "/admin/1'%20OR%201=1%20--",
			ExpectedDecoded: "/admin/1' OR 1=1 --",
			Description:     "SQL injection should be properly encoded",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			// Test encoding
			encoded := PathEncode(tc.Input)
			require.Equalf(t, tc.ExpectedEncoded, encoded, 
				"%s - Encoding mismatch:\nInput: %q\nExpected: %q\nGot: %q", 
				tc.Description, tc.Input, tc.ExpectedEncoded, encoded)

			// Test decoding
			decoded, err := PathDecode(tc.Input)
			require.Nilf(t, err, "%s - Decode error: %v", tc.Description, err)
			require.Equalf(t, tc.ExpectedDecoded, decoded,
				"%s - Decoding mismatch:\nInput: %q\nExpected: %q\nGot: %q", 
				tc.Description, tc.Input, tc.ExpectedDecoded, decoded)

			// Test roundtrip: encode then decode
			roundtrip, err := PathDecode(encoded)
			require.Nilf(t, err, "%s - Roundtrip decode error: %v", tc.Description, err)
			require.Equalf(t, tc.Input, roundtrip,
				"%s - Roundtrip failed:\nOriginal: %q\nEncoded: %q\nDecoded: %q", 
				tc.Description, tc.Input, encoded, roundtrip)
		})
	}
}

func TestPathEncodingSecurityImplications(t *testing.T) {
	// Test the key security difference: + vs %20 in SQL injection contexts
	sqlPayload := "1 OR 1=1"
	
	// Path encoding (always %20)
	pathEncoded := PathEncode(sqlPayload)
	require.Equal(t, "1%20OR%201=1", pathEncoded, "Path should encode spaces as %20")
	
	// Param encoding (always +)
	paramEncoded := ParamEncode(sqlPayload)
	require.Equal(t, "1+OR+1=1", paramEncoded, "Params should encode spaces as +")
	
	// Decoding behavior difference
	pathDecoded, err := PathDecode("test+plus")
	require.Nil(t, err)
	require.Equal(t, "test+plus", pathDecoded, "Path decode should preserve + as literal")
	
	pathDecodedSpace, err := PathDecode("test%20space")
	require.Nil(t, err)
	require.Equal(t, "test space", pathDecodedSpace, "Path decode should convert %20 to space")

	t.Log("✓ Path encoding uses %20 for spaces (correct for path context)")
	t.Log("✓ Param encoding uses + for spaces (correct for query context)")
	t.Log("✓ Path decode treats + as literal (preventing confusion)")
	t.Log("✓ Path decode converts %20 to space (standard percent decoding)")
}

func TestSpecificSQLInjectionPath(t *testing.T) {
	// Test the specific path you mentioned
	originalPath := "/admin/1'%20OR%201=1%20?key=y'+1=1&key2=value2"
	
	// Test decoding - this should convert %20 to spaces
	decoded, err := PathDecode(originalPath)
	require.Nil(t, err, "Failed to decode path")
	expectedDecoded := "/admin/1' OR 1=1 ?key=y'+1=1&key2=value2"
	require.Equal(t, expectedDecoded, decoded, 
		"Decoded path mismatch:\nInput:    %q\nExpected: %q\nGot:      %q", 
		originalPath, expectedDecoded, decoded)
	
	// Test encoding the decoded version - should re-encode spaces and ?
	encoded := PathEncode(decoded)
	expectedEncoded := "/admin/1'%20OR%201=1%20%3Fkey=y'+1=1&key2=value2"
	require.Equal(t, expectedEncoded, encoded,
		"Encoded path mismatch:\nInput:    %q\nExpected: %q\nGot:      %q", 
		decoded, expectedEncoded, encoded)
	
	// Verify that the + signs are preserved as literals in both operations
	require.Contains(t, decoded, "+1=1", "Plus signs should be preserved as literals during decode")
	require.Contains(t, encoded, "+1=1", "Plus signs should be preserved as literals during encode")
	
	// Verify that spaces are properly encoded as %20 (not +)
	require.Contains(t, encoded, "%20OR%20", "Spaces should be encoded as %20 in paths")
	require.NotContains(t, encoded, "+OR+", "Spaces should NOT be encoded as + in paths")
	
	// Verify that ? is encoded in paths (it has special meaning)
	require.Contains(t, encoded, "%3F", "Question mark should be encoded in paths")
	
	// Log the transformation for clarity
	t.Logf("Original (mixed encoding): %s", originalPath)
	t.Logf("Decoded (human readable):  %s", decoded)  
	t.Logf("Re-encoded (consistent):   %s", encoded)
	t.Log("✓ Percent-20 properly decoded to spaces")
	t.Log("✓ + preserved as literal characters") 
	t.Log("✓ Spaces re-encoded as percent-20 (not +)")
	t.Log("✓ ? encoded as percent-3F (has special meaning in paths)")
}
