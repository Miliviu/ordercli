package foodora

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestClient_ActiveOrders_DecodeFallback(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tracking/active-orders" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer tok" {
			t.Fatalf("auth=%q", r.Header.Get("Authorization"))
		}
		// includes unknown field to trigger DisallowUnknownFields decode failure and fallback
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
  "status": 200,
  "unknown": 1,
  "data": {"count":1,"active_orders":[{"code":"c","is_delivered":false,"vendor":{"code":"v","name":"N"},"status_messages":{"subtitle":"s","titles":[]}}]}
}`))
	}))
	t.Cleanup(srv.Close)

	c, err := New(Options{BaseURL: srv.URL + "/", AccessToken: "tok", UserAgent: "ua"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	resp, err := c.ActiveOrders(context.Background())
	if err != nil {
		t.Fatalf("ActiveOrders: %v", err)
	}
	if resp.Status != 200 || resp.Data.Count != 1 || len(resp.Data.ActiveOrders) != 1 {
		t.Fatalf("unexpected resp: %#v", resp)
	}
}

func TestClient_OrderHistoryByCode_Validation(t *testing.T) {
	t.Parallel()

	c, err := New(Options{BaseURL: "https://example.invalid/api/v5/", UserAgent: "ua"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.OrderHistoryByCode(context.Background(), OrderHistoryByCodeRequest{})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestClient_OrderStatus_QueryParams(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tracking/orders/abc" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("q") != "0" || q.Get("show_map_early_variation") == "" || q.Get("vendor_details_variation") == "" {
			t.Fatalf("unexpected query: %v", q)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":200,"data":{"status_messages":{"subtitle":"x"}}}`))
	}))
	t.Cleanup(srv.Close)

	c, err := New(Options{BaseURL: srv.URL + "/", AccessToken: "tok", UserAgent: "ua"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.OrderStatus(context.Background(), "abc")
	if err != nil {
		t.Fatalf("OrderStatus: %v", err)
	}
}

func TestClient_OrderReorder_PostJSONFallback(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method=%s", r.Method)
		}
		if r.URL.Path != "/orders/abc/reorder" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer tok" {
			t.Fatalf("auth=%q", r.Header.Get("Authorization"))
		}

		w.Header().Set("Content-Type", "application/json")
		// unknown field triggers strict decode failure
		_, _ = w.Write([]byte(`{
  "status": 200,
  "unknown": 1,
  "data": {
    "vendor_id": 1,
    "vendor_code": "v",
    "vendor_info": {"name":"Vendor","vertical":"restaurants","time_zone":"Europe/Vienna"},
    "cart": {"total_value": 1.0, "vendor_cart":[{"products":[]}]}
  }
}`))
	}))
	t.Cleanup(srv.Close)

	c, err := New(Options{BaseURL: srv.URL + "/", AccessToken: "tok", UserAgent: "ua"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	resp, err := c.OrderReorder(context.Background(), "abc", ReorderRequestBody{
		Address:     map[string]any{"id": "1"},
		ReorderTime: "2025-12-20T01:02:03+0100",
	})
	if err != nil {
		t.Fatalf("OrderReorder: %v", err)
	}
	if resp.Status != 200 || resp.Data.VendorCode != "v" {
		t.Fatalf("unexpected resp: %#v", resp)
	}
}

func TestNew_AddsTrailingSlash(t *testing.T) {
	t.Parallel()
	c, err := New(Options{BaseURL: "https://example.invalid/api/v5", UserAgent: "ua"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	u := c.baseURL.ResolveReference(&url.URL{Path: "x"}).String()
	if u != "https://example.invalid/api/v5/x" {
		t.Fatalf("got %q", u)
	}
}
