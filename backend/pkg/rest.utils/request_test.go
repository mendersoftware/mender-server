package rest

import (
	"net/http"
	"testing"
)

func TestRewriteForwardedRequest(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "http://localhost/foo/bar", nil)
	r.Header.Set("X-Forwarded-Method", http.MethodPost)
	r.Header.Set("X-Forwarded-Host", "hosted.mender.io")
	r.Header.Set("X-Forwarded-Uri", "/bar/baz")

	rf := RewriteForwardedRequest(r)
	if rf.Method != http.MethodPost {
		t.Errorf("unexpected method in forwarded request: %s (actual) != %s (expected)",
			rf.Method, http.MethodPost)
	}

	if rf.URL.Host != "hosted.mender.io" {
		t.Errorf("unexpected host in forwarded request: %s (actual) != hosted.mender.io",
			rf.URL.Host)
	}

	if rf.URL.Path != "/bar/baz" {
		t.Errorf("unexpected path in forwraded request: %s (actual) != /bar/baz", rf.URL.Path)
	}

	if r := RewriteForwardedRequest(nil); r != nil {
		t.Errorf("unexpected result rewriting nil request: %v", *r)
	}
}
