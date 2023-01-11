package urlutil

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"
)

// RawURL is wrapper around url.URL that can handle all
// path and parameters including raw unsafe requests
type RawURL struct {
	Original string // raw or Original string (excluding params and fragment)
	Path     string // Path is Relative Path
	Unsafe   bool   // when unsafe parsing(host,port etc) is not guaranteed
	Fragment string
	// internal
	params  Params
	baseUrl *url.URL
}

// MergePath merges (/blog/admin?user=true + /blog/admin/profile?show=true => /blog/admin/profile?user=true&show=true)
// and all other edgecases
func (r *RawURL) MergePath(relpath string, unsafe bool) {
	// Use RawURL and handle extra data it it is not a relativepath
	rr := RawURL{
		Original: relpath,
		Unsafe:   unsafe,
	}
	rr.fetchParams()
	// automerge parameters
	r.params.Merge(rr.params)
	r.Path = mergePaths(r.Path, rr.Original)
}

// Query returns Parameters
func (r *RawURL) Query() Params {
	return r.params
}

// parseRelativePath from OriginalPath
func (r *RawURL) parseRelativePath() {
	if strings.HasPrefix(r.Original, "/") && !strings.HasPrefix(r.Original, "//") {
		// this is definitely a relative path
		r.Path = r.Original
		return
	}
	if r.Unsafe && (strings.Contains(r.Original, "%") || strings.Contains(r.Original, ".")) {
		// url.parse discards percent encoded data and . in url
		// i.e /%20test%0a =? /
		// we don't want this
		r.Path = r.Original
		return
	}

	fullurl := r.Original
	if !strings.Contains(r.Original, "//") {
		fullurl = "https://" + fullurl
	}
	urx, er := url.Parse(fullurl)
	if er != nil {
		// nothing to say here
		if index := strings.Index(fullurl, "/"); index != -1 {
			r.Path = fullurl[index:]
		} else {
			r.Path = r.Original
		}
	} else {
		r.Path = urx.Path
		// From url.Path:  path (relative paths may omit leading slash)
		// we don't allow this
		if r.Path != "" && !strings.HasPrefix(r.Path, "/") {
			r.Path = "/" + r.Path
		}
	}
}

// fetchParams fetches parameters from
func (r *RawURL) fetchParams() {
	if r.params == nil {
		r.params = make(Params)
	}
	// parse fragments if any
	if i := strings.IndexRune(r.Original, '#'); i != -1 {
		// assuming ?param=value#highlight
		r.Fragment = r.Original[i+1:]
		r.Original = r.Original[:i]
	}
	if index := strings.IndexRune(r.Original, '?'); index == -1 {
		return
	} else {
		encodedParams := r.Original[index+1:]
		r.params.Decode(encodedParams)
		r.Original = r.Original[:index]
	}
}

// String returns complete url if possible
func (r *RawURL) String() string {
	var buff bytes.Buffer
	if r.baseUrl != nil {
		if r.baseUrl.Scheme != "" {
			buff.WriteString(r.baseUrl.Scheme + "://")
		}
		buff.WriteString(r.baseUrl.Host)
		if len(r.Path) > 0 && !strings.HasPrefix(r.Path, "/") {
			r.Path = "/" + r.Path
		}
		buff.WriteString(r.Path)
	} else {
		buff.WriteString(r.Path)
	}
	if len(r.params) > 0 {
		buff.WriteString("?" + r.params.Encode())
	}
	if r.Fragment != "" {
		buff.WriteString("#" + r.Fragment)
	}
	return buff.String()
}

// ParseURL returns parsed URL
func ParseURL(uri string, unsafe bool) (*RawURL, error) {
	if uri == "" {
		return nil, fmt.Errorf("url cannot be empty")
	}
	r := &RawURL{}
	r.Original = uri
	r.fetchParams()
	addedschema := false
	if !strings.HasPrefix(uri, "/") && !strings.Contains(uri, "//") {
		uri = "https://" + uri
		addedschema = true
	}
	if !unsafe {
		u, err := url.Parse(uri)
		if err != nil {
			return nil, err
		} else {
			r.baseUrl = u
		}
	} else {
		r.baseUrl, _ = url.Parse(uri)
		r.Unsafe = true
	}
	if r.baseUrl != nil && addedschema {
		r.baseUrl.Scheme = ""
	}
	r.parseRelativePath()
	return r, nil
}

// AutoMergePaths merges two relative paths including parameters and returns final string
func AutoMergePaths(relpath1 string, relpath2 string) string {
	r1 := RawURL{
		Original: relpath1,
	}
	r1.fetchParams()
	r2 := RawURL{
		Original: relpath2,
	}
	r2.fetchParams()
	r1.params.Merge(r2.params)
	var buff bytes.Buffer
	buff.WriteString(mergePaths(r1.Original, r2.Original))
	if len(r1.params) > 0 {
		buff.WriteString("?" + r1.params.Encode())
	}
	return buff.String()
}

// mergePaths merges two relative paths
func mergePaths(elem1 string, elem2 string) string {
	// if both have slash remove one
	if strings.HasSuffix(elem1, "/") && strings.HasPrefix(elem2, "/") {
		elem2 = strings.TrimLeft(elem2, "/")
	}

	if elem1 == "" {
		return elem2
	} else if elem2 == "" {
		return elem1
	}

	// if both paths donot have a slash add it to beginning of second
	if !strings.HasSuffix(elem1, "/") && !strings.HasPrefix(elem2, "/") {
		elem2 = "/" + elem2
	}

	// Do not normalize but combibe paths same as path.join
	/*
		Merge Examples (Same as path.Join)
		/blog   /admin => /blog/admin
		/blog/wp /wp-content  => /blog/wp/wp-content
		/blog/admin /blog/admin/profile => /blog/admin/profile
		/blog/admin /blog => /blog/admin/blog
		/blog /blog/ => /blog/
	*/

	if elem1 == elem2 {
		return elem1
	} else if len(elem1) > len(elem2) && strings.HasSuffix(elem1, elem2) {
		return elem1
	} else if len(elem1) < len(elem2) && strings.HasPrefix(elem2, elem1) {
		return elem2
	} else {
		return elem1 + elem2
	}
}
