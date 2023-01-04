package urlutil

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

func TestUTF8URLEncoding(t *testing.T) {
	exstring := '上'
	expected := `e4%b8%8a`
	val := getutf8hex(exstring)
	if val != expected {
		t.Errorf("failed to url encode utf char expected %v but got %v", expected, val)
	}
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
		if val != v.Expected {
			t.Errorf("failed to url encode payload expected %v got %v", v.Expected, val)
		}
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
	if p.Encode() != expected {
		t.Errorf("failed to encode parameters expected %v but got %v", expected, p.Encode())
	}
}

func TestParamIntegration(t *testing.T) {
	var routerErr error
	expected := "/params?jsprotocol=javascript://alert(1)&sqli=1+AND+(SELECT+*+FROM+(SELECT(SLEEP(12)))nQIP)&xss=<script>alert('XSS')</script>&xssiwthspace=<svg+id=alert(1)+onload=eval(id)>"

	http.HandleFunc("/params", func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI != expected {
			routerErr = fmt.Errorf("expected %v but got %v", expected, r.RequestURI)
		}
		w.WriteHeader(http.StatusOK)
	})
	go http.ListenAndServe(":9000", nil)

	p := NewParams()
	p.Add("sqli", "1+AND+(SELECT+*+FROM+(SELECT(SLEEP(12)))nQIP)")
	p.Add("xss", "<script>alert('XSS')</script>")
	p.Add("xssiwthspace", "<svg id=alert(1) onload=eval(id)>")
	p.Add("jsprotocol", "javascript://alert(1)")

	url, _ := url.Parse("http://localhost:9000/params")
	url.RawQuery = p.Encode()
	_, err := http.Get(url.String())
	if err != nil {
		panic(err)
	}
	if routerErr != nil {
		t.Errorf(routerErr.Error())
	}
}

func TestPercentEncoding(t *testing.T) {
	// From Burpsuite
	expected := "%74%65%73%74%20%26%23%20%28%29%20%70%65%72%63%65%6e%74%20%5b%5d%7c%2a%20%65%6e%63%6f%64%69%6e%67"
	payload := "test &# () percent []|* encoding"
	value := PercentEncoding(payload)
	if value != expected {
		t.Errorf("expected percentencoding to be %v but got %v", expected, payload)
	}
}
