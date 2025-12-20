package cli

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/steipete/ordercli/internal/config"
)

type rtFunc func(*http.Request) (*http.Response, error)

func (f rtFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestResolveClientSecret_FetchesAndCaches(t *testing.T) {
	// touches http.DefaultTransport/env
	withEnvMap(t, map[string]string{"FOODORA_CLIENT_SECRET": ""})

	orig := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = orig })

	http.DefaultTransport = rtFunc(func(r *http.Request) (*http.Response, error) {
		switch {
		case strings.Contains(r.URL.Host, "firebaseinstallations.googleapis.com"):
			body := `{"fid":"fid123","authToken":{"token":"at123"}}`
			return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(body))}, nil
		case strings.Contains(r.URL.Host, "firebaseremoteconfig.googleapis.com"):
			body := `{"state":"UPDATE","templateVersion":"1","entries":{"client_secrets":"{\"MJ\":\"\",\"AT\":\"{\\\"android\\\":\\\"sec\\\"}\"}"}}`
			return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(body))}, nil
		default:
			return &http.Response{StatusCode: 500, Body: io.NopCloser(bytes.NewReader([]byte("unexpected")))}, nil
		}
	})

	st := &state{cfg: config.New(), configPath: "x"}
	fc := st.foodora()
	fc.BaseURL = "https://mj.fd-api.com/api/v5/"
	fc.TargetCountryISO = "AT"
	fc.ClientSecret = ""
	fc.OAuthClientID = ""

	sec, err := st.resolveClientSecret(context.Background(), "android")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if sec.Secret != "sec" || !sec.FromFetch {
		t.Fatalf("unexpected: %#v", sec)
	}
	if fc.ClientSecret != "sec" || fc.OAuthClientID != "android" {
		t.Fatalf("unexpected cfg: %#v", fc)
	}
	if !st.dirty {
		t.Fatalf("expected dirty")
	}

	sec2, err := st.forceFetchClientSecret(context.Background(), "android")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if sec2.Secret != "sec" || !sec2.FromFetch {
		t.Fatalf("unexpected: %#v", sec2)
	}
}
