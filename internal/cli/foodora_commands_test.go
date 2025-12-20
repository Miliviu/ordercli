package cli

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/steipete/ordercli/internal/config"
)

func TestFoodoraCLI_Flow_Login_History_Reorder(t *testing.T) {
	// touches os.Stdin/env
	cfgPath := filepath.Join(t.TempDir(), "config.json")

	srv := newFoodoraTestServer(t)
	defer srv.Close()

	setEnv(t, "FOODORA_CLIENT_SECRET", "secret")

	// config set
	{
		out, errOut, err := runCLI(cfgPath, []string{"foodora", "config", "set", "--base-url", srv.URL + "/", "--global-entity-id", "MJM_AT", "--target-iso", "AT"}, "")
		if err != nil {
			t.Fatalf("config set: %v out=%s err=%s", err, out, errOut)
		}
	}

	// login triggers MFA (no OTP)
	{
		out, errOut, err := runCLI(cfgPath, []string{"foodora", "login", "--email", "a@example.com", "--password-stdin", "--wait-for-otp=false"}, "pw\n")
		if err != nil {
			t.Fatalf("login mfa: %v out=%s err=%s", err, out, errOut)
		}
		if !strings.Contains(errOut, "MFA triggered") {
			t.Fatalf("expected MFA message, err=%s", errOut)
		}
	}

	// login with OTP succeeds (reuses pending MFA token from config)
	{
		out, errOut, err := runCLI(cfgPath, []string{"foodora", "login", "--email", "a@example.com", "--password-stdin", "--otp-method", "sms", "--otp", "1234", "--wait-for-otp=false"}, "pw\n")
		if err != nil {
			t.Fatalf("login ok: %v out=%s err=%s", err, out, errOut)
		}
		if strings.TrimSpace(out) != "ok" {
			t.Fatalf("unexpected out=%q err=%s", out, errOut)
		}
	}

	// orders
	{
		out, _, err := runCLI(cfgPath, []string{"foodora", "orders"}, "")
		if err != nil {
			t.Fatalf("orders: %v", err)
		}
		if !strings.Contains(out, "OC-1") {
			t.Fatalf("unexpected out=%s", out)
		}
	}

	// single order
	{
		out, _, err := runCLI(cfgPath, []string{"foodora", "order", "OC-1"}, "")
		if err != nil {
			t.Fatalf("order: %v", err)
		}
		if !strings.Contains(out, "status=200") {
			t.Fatalf("unexpected out=%s", out)
		}
	}

	// history list
	{
		out, _, err := runCLI(cfgPath, []string{"foodora", "history", "--limit", "1"}, "")
		if err != nil {
			t.Fatalf("history: %v", err)
		}
		if !strings.Contains(out, "HIST-1") {
			t.Fatalf("unexpected out=%s", out)
		}
	}

	// history show
	{
		out, _, err := runCLI(cfgPath, []string{"foodora", "history", "show", "HIST-1"}, "")
		if err != nil {
			t.Fatalf("history show: %v", err)
		}
		if !strings.Contains(out, "vendor=Test Vendor") || !strings.Contains(out, "items:") {
			t.Fatalf("unexpected out=%s", out)
		}
	}

	// reorder preview (no endpoint call)
	{
		out, errOut, err := runCLI(cfgPath, []string{"foodora", "reorder", "HIST-1"}, "")
		if err != nil {
			t.Fatalf("reorder preview: %v", err)
		}
		if !strings.Contains(out, "vendor=Test Vendor") || !strings.Contains(errOut, "--confirm") {
			t.Fatalf("unexpected out=%s err=%s", out, errOut)
		}
	}

	// reorder confirm (calls endpoint)
	{
		out, errOut, err := runCLI(cfgPath, []string{"foodora", "reorder", "HIST-1", "--confirm"}, "")
		if err != nil {
			t.Fatalf("reorder confirm: %v", err)
		}
		if !strings.Contains(out, "vendor=Test Vendor") || !strings.Contains(out, "items:") {
			t.Fatalf("unexpected out=%s err=%s", out, errOut)
		}
	}

	// session refresh
	{
		out, _, err := runCLI(cfgPath, []string{"foodora", "session", "refresh", "--client-id", "android"}, "")
		if err != nil {
			t.Fatalf("session refresh: %v", err)
		}
		if strings.TrimSpace(out) != "ok" {
			t.Fatalf("unexpected out=%q", out)
		}
	}

	// countries
	{
		out, _, err := runCLI(cfgPath, []string{"foodora", "countries"}, "")
		if err != nil {
			t.Fatalf("countries: %v", err)
		}
		if !strings.Contains(out, "AT") || !strings.Contains(out, "MJM_AT") {
			t.Fatalf("unexpected out=%s", out)
		}
	}

	// logout clears tokens
	{
		out, _, err := runCLI(cfgPath, []string{"foodora", "logout"}, "")
		if err != nil {
			t.Fatalf("logout: %v", err)
		}
		if strings.TrimSpace(out) != "ok" {
			t.Fatalf("unexpected out=%q", out)
		}
		out, _, err = runCLI(cfgPath, []string{"foodora", "config", "show"}, "")
		if err != nil {
			t.Fatalf("config show: %v", err)
		}
		if strings.Contains(out, "access_token=***") || strings.Contains(out, "refresh_token=***") {
			t.Fatalf("expected tokens cleared, got: %s", out)
		}
	}
}

func runCLI(cfgPath string, args []string, stdin string) (stdout string, stderr string, err error) {
	root := newRoot()
	var out bytes.Buffer
	var errOut bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errOut)
	root.SetArgs(append([]string{"--config", cfgPath}, args...))

	if stdin != "" {
		restore := replaceStdin(stdin)
		defer restore()
	}

	err = root.Execute()
	return out.String(), errOut.String(), err
}

func replaceStdin(input string) func() {
	r, w, _ := os.Pipe()
	_, _ = w.WriteString(input)
	_ = w.Close()
	old := os.Stdin
	os.Stdin = r
	return func() {
		_ = r.Close()
		os.Stdin = old
	}
}

func setEnv(t *testing.T, k, v string) {
	t.Helper()
	old, had := os.LookupEnv(k)
	if err := os.Setenv(k, v); err != nil {
		t.Fatalf("setenv: %v", err)
	}
	t.Cleanup(func() {
		if had {
			_ = os.Setenv(k, old)
		} else {
			_ = os.Unsetenv(k)
		}
	})
}

func newFoodoraTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()

	mux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		grant := r.FormValue("grant_type")

		switch grant {
		case "password":
			if r.Header.Get("X-OTP") == "" {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("ratelimit-reset", "12")
				w.WriteHeader(401)
				_, _ = w.Write([]byte(`{"code":"mfa_triggered","metadata":{"more_information":{"channel":"sms","email":"a@example.com","mfa_token":"mfa"}}}`))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"access","refresh_token":"refresh","expires_in":3600}`))
			return

		case "refresh_token":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"access2","refresh_token":"refresh2","expires_in":3600}`))
			return
		default:
			t.Fatalf("unexpected grant_type %q", grant)
		}
	})

	mux.HandleFunc("/tracking/active-orders", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":200,"data":{"count":1,"active_orders":[{"code":"OC-1","is_delivered":false,"vendor":{"code":"V","name":"Vendor"},"status_messages":{"subtitle":"Cooking","titles":[]}}]}}`))
	})

	mux.HandleFunc("/tracking/orders/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":200,"data":{"status_messages":{"subtitle":"Cooking"}}}`))
	})

	mux.HandleFunc("/orders/order_history", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		q := r.URL.Query()
		if q.Get("order_code") != "" {
			_, _ = w.Write([]byte(`{"status":200,"data":{"total_count":1,"items":[{"order_code":"HIST-1","current_status":{"message":"delivered"},"confirmed_delivery_time":{"date":"2025-12-20T00:00:00Z","timezone":"Europe/Vienna"},"vendor":{"code":"V","name":"Test Vendor"},"total_value":12.3,"order_products":[{"name":"Burger","quantity":1,"total_price":12.3}] }]}}`))
			return
		}
		_, _ = w.Write([]byte(`{"status":200,"data":{"total_count":1,"items":[{"order_code":"HIST-1","current_status":{"message":"delivered"},"confirmed_delivery_time":{"date":"2025-12-20T00:00:00Z","timezone":"Europe/Vienna"},"vendor":{"code":"V","name":"Test Vendor"},"total_value":12.3}]}}`))
	})

	mux.HandleFunc("/customers/addresses", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":200,"data":{"items":[{"id":"addr1","is_selected":true,"street":"Main"}]}}`))
	})

	mux.HandleFunc("/orders/", func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/reorder") {
			http.NotFound(w, r)
			return
		}
		// basic validation: ensure JSON body contains reorder_time + address
		b, _ := ioReadAllBody(r)
		if !bytes.Contains(b, []byte(`"reorder_time"`)) || !bytes.Contains(b, []byte(`"address"`)) {
			t.Fatalf("unexpected reorder body: %s", string(b))
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":200,"data":{"vendor_id":1,"vendor_code":"V","vendor_info":{"name":"Test Vendor","vertical":"restaurants","time_zone":"Europe/Vienna"},"cart":{"total_value":12.3,"vendor_cart":[{"products":[{"name":"Burger","variation_name":"Cheese","quantity":1,"total_price":12.3,"price":12.3,"is_available":true,"toppings":[{"id":1,"name":"Bacon","description":"","is_available":true,"position":"","price":1.0}]}]}]}}}`))
	})

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// mimic auth requirement for most endpoints
		switch r.URL.Path {
		case "/oauth2/token":
		default:
			if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") && r.URL.Path != "/oauth2/token" {
				http.Error(w, "missing auth", http.StatusUnauthorized)
				return
			}
		}
		mux.ServeHTTP(w, r)
	}))
}

func ioReadAllBody(r *http.Request) ([]byte, error) {
	defer r.Body.Close()
	return io.ReadAll(r.Body)
}

func TestDefaultWebURLForConfig(t *testing.T) {
	st := &state{}
	st.cfg = config.New()
	st.foodora().TargetCountryISO = "AT"
	u, ok := defaultWebURLForConfig(st)
	if !ok || u != "https://www.foodora.at/" {
		t.Fatalf("got %q ok=%v", u, ok)
	}
	st.foodora().TargetCountryISO = "HU"
	if _, ok := defaultWebURLForConfig(st); ok {
		t.Fatalf("expected false")
	}
}
