package foodora

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestClient_OAuthTokenPassword_Success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method=%s", r.Method)
		}
		if r.URL.Path != "/oauth2/token" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		if r.Header.Get("X-Device") != "dev" || r.Header.Get("Device-Id") != "dev" {
			t.Fatalf("missing device headers")
		}
		if r.Header.Get("X-FP-API-KEY") != "android" {
			t.Fatalf("missing fp api key")
		}
		if r.Header.Get("App-Name") != "at.mjam" {
			t.Fatalf("missing app name")
		}

		b, _ := ioReadAll(r)
		form, err := url.ParseQuery(string(b))
		if err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if form.Get("grant_type") != "password" || form.Get("username") != "u" || form.Get("password") != "p" {
			t.Fatalf("bad form: %v", form)
		}
		if form.Get("client_secret") != "s" || form.Get("client_id") != "android" || form.Get("scope") != "API_CUSTOMER" {
			t.Fatalf("bad oauth fields: %v", form)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"a","refresh_token":"r","expires_in":3600}`))
	}))
	t.Cleanup(srv.Close)

	c, err := New(Options{
		BaseURL:     srv.URL + "/",
		DeviceID:    "dev",
		UserAgent:   "ua",
		FPAPIKey:    "android",
		AppName:     "at.mjam",
		AccessToken: "ignored",
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	tok, mfa, err := c.OAuthTokenPassword(context.Background(), OAuthPasswordRequest{
		Username:     "u",
		Password:     "p",
		ClientSecret: "s",
		ClientID:     "android",
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if mfa != nil {
		t.Fatalf("unexpected mfa: %#v", mfa)
	}
	if tok.AccessToken != "a" || tok.RefreshToken != "r" || tok.ExpiresIn != 3600 {
		t.Fatalf("unexpected token: %#v", tok)
	}
}

func TestClient_OAuthTokenPassword_MfaTriggered(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("ratelimit-reset", "12")
		w.WriteHeader(401)
		_, _ = w.Write([]byte(`{
  "code":"mfa_triggered",
  "metadata":{"more_information":{"channel":"sms","email":"e","mfa_token":"tok"}}
}`))
	}))
	t.Cleanup(srv.Close)

	c, err := New(Options{BaseURL: srv.URL + "/", UserAgent: "ua"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	tok, mfa, err := c.OAuthTokenPassword(context.Background(), OAuthPasswordRequest{
		Username:     "u",
		Password:     "p",
		ClientSecret: "s",
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if tok.AccessToken != "" || tok.RefreshToken != "" {
		t.Fatalf("unexpected token: %#v", tok)
	}
	if mfa == nil || mfa.MfaToken != "tok" || mfa.Channel != "sms" || mfa.Email != "e" || mfa.RateLimitReset != 12 {
		t.Fatalf("unexpected mfa: %#v", mfa)
	}
}

func TestClient_OAuthTokenRefresh_SetsClientIDDefault(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := ioReadAll(r)
		form, err := url.ParseQuery(string(b))
		if err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if form.Get("grant_type") != "refresh_token" || form.Get("refresh_token") != "rr" {
			t.Fatalf("bad form: %v", form)
		}
		if form.Get("client_id") != "android" {
			t.Fatalf("expected default client_id android, got %q", form.Get("client_id"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"a","refresh_token":"r","expires_in":1}`))
	}))
	t.Cleanup(srv.Close)

	c, err := New(Options{BaseURL: srv.URL + "/", UserAgent: "ua"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	_, err = c.OAuthTokenRefresh(context.Background(), OAuthRefreshRequest{RefreshToken: "rr", ClientSecret: "s"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
}

func ioReadAll(r *http.Request) ([]byte, error) {
	defer r.Body.Close()
	return io.ReadAll(r.Body)
}
