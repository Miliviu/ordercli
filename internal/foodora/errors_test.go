package foodora

import (
	"net/http"
	"testing"
)

func TestParseMfaTriggered(t *testing.T) {
	body := []byte(`{
  "code": "mfa_triggered",
  "metadata": {
    "more_information": {
      "channel": "sms",
      "email": "peter@example.com",
      "mfa_token": "tok123"
    }
  }
}`)
	h := http.Header{}
	h.Set("ratelimit-reset", "12")

	ch, ok := parseMfaTriggered(body, h)
	if !ok {
		t.Fatalf("expected ok")
	}
	if ch.Channel != "sms" || ch.Email != "peter@example.com" || ch.MfaToken != "tok123" || ch.RateLimitReset != 12 {
		t.Fatalf("unexpected challenge: %#v", ch)
	}
}
