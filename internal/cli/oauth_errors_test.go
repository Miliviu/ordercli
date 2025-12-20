package cli

import (
	"testing"

	"github.com/steipete/ordercli/internal/foodora"
)

func TestIsInvalidClientErr(t *testing.T) {
	err := &foodora.HTTPError{
		Method:     "POST",
		URL:        "x",
		StatusCode: 401,
		Body:       []byte(`{"error":"invalid_client"}`),
	}
	if !isInvalidClientErr(err) {
		t.Fatalf("expected true")
	}
	if isInvalidClientErr(&foodora.HTTPError{StatusCode: 500, Body: []byte(`{"error":"invalid_client"}`)}) {
		t.Fatalf("expected false")
	}
}
