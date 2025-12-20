package deliveroo

import (
	"strings"
	"testing"
)

func TestBuildAPIBaseURL_DefaultTemplate(t *testing.T) {
	t.Parallel()

	got, err := BuildAPIBaseURL("", "uk")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !strings.HasPrefix(got, "https://") {
		t.Fatalf("got %q", got)
	}
}

func TestNormalizeBaseURL(t *testing.T) {
	t.Parallel()

	got, err := NormalizeBaseURL("example.com/consumer?x=1#y")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if strings.Contains(got, "?") || strings.Contains(got, "#") {
		t.Fatalf("expected stripped, got %q", got)
	}
	if got != "https://example.com/consumer" {
		t.Fatalf("got %q", got)
	}
}

func TestConsumerBaseURL(t *testing.T) {
	t.Parallel()

	got, err := ConsumerBaseURL("https://example.com")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != "https://example.com/consumer" {
		t.Fatalf("got %q", got)
	}
}

func TestOrderHistoryURL(t *testing.T) {
	t.Parallel()

	_, err := OrderHistoryURL("https://example.com/consumer", OrderHistoryParams{Offset: -1, Limit: 10})
	if err == nil {
		t.Fatalf("expected offset error")
	}
	_, err = OrderHistoryURL("https://example.com/consumer", OrderHistoryParams{Offset: 0, Limit: 0})
	if err == nil {
		t.Fatalf("expected limit error")
	}

	u, err := OrderHistoryURL("https://example.com/consumer", OrderHistoryParams{Offset: 2, Limit: 5, IncludeUgc: true})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !strings.Contains(u, "offset=2") || !strings.Contains(u, "limit=5") || !strings.Contains(u, "include_ugc=true") {
		t.Fatalf("got %q", u)
	}
}
