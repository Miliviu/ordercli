package foodora

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type HTTPError struct {
	Method     string
	URL        string
	StatusCode int
	Body       []byte
}

func (e *HTTPError) Error() string {
	body := string(e.Body)
	if len(body) > 300 {
		body = body[:300] + "…"
	}
	return fmt.Sprintf("%s %s: HTTP %d: %s", e.Method, e.URL, e.StatusCode, body)
}

type MfaChallenge struct {
	Channel        string
	Email          string
	MfaToken       string
	RateLimitReset int
}

func parseMfaTriggered(body []byte, header http.Header) (MfaChallenge, bool) {
	var raw struct {
		Code     string `json:"code"`
		Metadata struct {
			MoreInformation struct {
				Channel  string `json:"channel"`
				Email    string `json:"email"`
				MfaToken string `json:"mfa_token"`
			} `json:"more_information"`
		} `json:"metadata"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return MfaChallenge{}, false
	}
	if raw.Code != "mfa_triggered" {
		return MfaChallenge{}, false
	}

	reset := 30
	if v := header.Get("ratelimit-reset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			reset = n
		}
	}

	ch := MfaChallenge{
		Channel:        raw.Metadata.MoreInformation.Channel,
		Email:          raw.Metadata.MoreInformation.Email,
		MfaToken:       raw.Metadata.MoreInformation.MfaToken,
		RateLimitReset: reset,
	}
	if ch.MfaToken == "" {
		return MfaChallenge{}, false
	}
	return ch, true
}
