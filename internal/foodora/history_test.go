package foodora

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAuthToken_ExpiresAt(t *testing.T) {
	now := time.Unix(1000, 0).UTC()
	tok := AuthToken{ExpiresIn: 10}
	if got := tok.ExpiresAt(now); got.Unix() != 1010 {
		t.Fatalf("got %v", got)
	}
	if (AuthToken{}).ExpiresAt(now) != (time.Time{}) {
		t.Fatalf("expected zero")
	}
}

func TestClient_OrderHistory_Defaults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/orders/order_history" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("include") == "" || q.Get("limit") == "" || q.Get("offset") == "" {
			t.Fatalf("missing params: %v", q)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":200,"data":{"total_count":1,"items":[{"order_code":"O","total_value":1.0}]}}`))
	}))
	t.Cleanup(srv.Close)

	c, err := New(Options{BaseURL: srv.URL + "/", AccessToken: "tok", UserAgent: "ua"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	resp, err := c.OrderHistory(context.Background(), OrderHistoryRequest{})
	if err != nil {
		t.Fatalf("OrderHistory: %v", err)
	}
	if resp.Status != 200 || len(resp.Data.Items) != 1 || resp.Data.Items[0].OrderCode != "O" {
		t.Fatalf("unexpected: %#v", resp)
	}
}

func TestClient_OrderHistoryByCode_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/orders/order_history" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("order_code") != "X" {
			t.Fatalf("q=%v", q)
		}
		if q.Get("item_replacement") != "true" {
			t.Fatalf("q=%v", q)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":200,"data":{"total_count":1,"items":[{"order_code":"X","order_products":[{"name":"Burger","quantity":1}]}]}}`))
	}))
	t.Cleanup(srv.Close)

	c, err := New(Options{BaseURL: srv.URL + "/", AccessToken: "tok", UserAgent: "ua"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	resp, err := c.OrderHistoryByCode(context.Background(), OrderHistoryByCodeRequest{
		OrderCode:       "X",
		ItemReplacement: true,
	})
	if err != nil {
		t.Fatalf("OrderHistoryByCode: %v", err)
	}
	if resp.Status != 200 || len(resp.Data.Items) != 1 {
		t.Fatalf("unexpected: %#v", resp)
	}
}

func TestClient_SetAccessToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok2" {
			t.Fatalf("auth=%q", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":200,"data":{"count":0,"active_orders":[]}}`))
	}))
	t.Cleanup(srv.Close)

	c, err := New(Options{BaseURL: srv.URL + "/", AccessToken: "tok1", UserAgent: "ua"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	c.SetAccessToken("tok2")
	if _, err := c.ActiveOrders(context.Background()); err != nil {
		t.Fatalf("ActiveOrders: %v", err)
	}
}
