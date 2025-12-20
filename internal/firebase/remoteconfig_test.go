package firebase

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestRemoteConfigClient_Fetch(t *testing.T) {
	t.Parallel()

	c := NewRemoteConfigClient(MjamAT)

	var sawInstall bool
	var sawFetch bool

	c.http.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.Header.Get("X-Android-Package") != MjamAT.PackageName {
			t.Fatalf("missing/incorrect X-Android-Package")
		}
		if r.Header.Get("X-Android-Cert") != MjamAT.CertSHA1 {
			t.Fatalf("missing/incorrect X-Android-Cert")
		}

		switch {
		case strings.HasPrefix(r.URL.Host, "firebaseinstallations.googleapis.com") && strings.Contains(r.URL.Path, "/installations"):
			sawInstall = true
			body := `{"fid":"fid123","authToken":{"token":"at123"}}`
			return &http.Response{
				StatusCode: 200,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(body)),
			}, nil

		case strings.HasPrefix(r.URL.Host, "firebaseremoteconfig.googleapis.com") && strings.Contains(r.URL.Path, "/namespaces/firebase:fetch"):
			sawFetch = true
			if r.Header.Get("X-Goog-Firebase-Installations-Id") == "" {
				t.Fatalf("missing X-Goog-Firebase-Installations-Id")
			}
			if r.Header.Get("X-Goog-Firebase-Installations-Auth") == "" {
				t.Fatalf("missing X-Goog-Firebase-Installations-Auth")
			}
			body := `{"state":"UPDATE","templateVersion":"1","entries":{"client_secrets":"{\"AT\":\"secret\"}"}}`
			return &http.Response{
				StatusCode: 200,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(body)),
			}, nil
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
			return nil, nil
		}
	})

	resp, err := c.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if !sawInstall || !sawFetch {
		t.Fatalf("expected both calls, install=%v fetch=%v", sawInstall, sawFetch)
	}
	if resp.State != "UPDATE" || resp.TemplateVersion != "1" {
		t.Fatalf("unexpected resp: %#v", resp)
	}
	if resp.Entries["client_secrets"] == "" {
		t.Fatalf("missing entries")
	}
}

func TestNewFID_Length(t *testing.T) {
	t.Parallel()

	fid, err := newFID()
	if err != nil {
		t.Fatalf("newFID: %v", err)
	}
	// 16 bytes base64url => 22 chars, no padding.
	if len(fid) != 22 {
		t.Fatalf("len=%d fid=%q", len(fid), fid)
	}
	if strings.Contains(fid, "=") {
		t.Fatalf("unexpected padding: %q", fid)
	}
}

func TestRemoteConfigClient_Fetch_ErrorStatus(t *testing.T) {
	t.Parallel()

	c := NewRemoteConfigClient(MjamAT)
	c.http.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 500,
			Header:     http.Header{"Content-Type": []string{"text/plain"}},
			Body:       io.NopCloser(bytes.NewReader([]byte("nope"))),
		}, nil
	})

	_, err := c.Fetch(context.Background())
	if err == nil {
		t.Fatalf("expected error")
	}
}
