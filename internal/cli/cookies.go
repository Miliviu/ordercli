package cli

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/steipete/foodcli/internal/chromecookies"
)

func newCookiesCmd(st *state) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cookies",
		Short: "Cookie helpers (Cloudflare / browser sync)",
	}
	cmd.AddCommand(newCookiesChromeCmd(st))
	return cmd
}

func newCookiesChromeCmd(st *state) *cobra.Command {
	var profile string
	var cookiePath string
	var timeout time.Duration
	var filterNames []string
	var sourceURL string

	cmd := &cobra.Command{
		Use:   "chrome",
		Short: "Import cookies from local Chrome into config (for base_url host)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if st.cfg.BaseURL == "" {
				return errors.New("missing base_url (run `foodcli config set --country ...`)")
			}
			host := cookieHost(st.cfg.BaseURL)
			if host == "" {
				return fmt.Errorf("failed to derive host from base_url=%q", st.cfg.BaseURL)
			}

			targetURL := strings.TrimSpace(sourceURL)
			if targetURL == "" {
				targetURL = st.cfg.BaseURL
			}

			if u, err := url.Parse(targetURL); err == nil && u.Hostname() != "" && (u.Scheme == "http" || u.Scheme == "https") {
				targetURL = u.Scheme + "://" + u.Hostname() + "/"
			} else if !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
				targetURL = "https://" + host + "/"
			}

			cacheDir := filepath.Join(filepath.Dir(st.configPath), "chrome-cookies")
			res, err := chromecookies.LoadCookieHeader(cmd.Context(), chromecookies.Options{
				TargetURL:          targetURL,
				ChromeProfile:      profile,
				ExplicitCookiePath: cookiePath,
				FilterNames:        filterNames,
				Timeout:            timeout,
				CacheDir:           cacheDir,
				LogWriter:          cmd.ErrOrStderr(),
			})
			if err != nil {
				return err
			}
			if strings.TrimSpace(res.CookieHeader) == "" {
				return errors.New("no cookies found (are you logged in in Chrome? try --profile \"Default\" / \"Profile 1\" or --cookie-path)")
			}

			if st.cfg.CookiesByHost == nil {
				st.cfg.CookiesByHost = map[string]string{}
			}
			st.cfg.CookiesByHost[strings.ToLower(host)] = res.CookieHeader
			st.markDirty()

			fmt.Fprintf(cmd.OutOrStdout(), "ok host=%s cookies=%d\n", host, res.CookieCount)
			return nil
		},
	}

	cmd.Flags().StringVar(&profile, "profile", "", "Chrome profile name (Default, Profile 1, ...) or path to profile dir")
	cmd.Flags().StringVar(&cookiePath, "cookie-path", "", "explicit Cookies DB path or profile dir (overrides --profile)")
	cmd.Flags().StringVar(&sourceURL, "url", "", "URL to load cookies from (default: base_url origin)")
	cmd.Flags().StringSliceVar(&filterNames, "filter-name", nil, "cookie name to include (repeatable; default: all for target URL)")
	cmd.Flags().DurationVar(&timeout, "timeout", 5*time.Second, "cookie read timeout (keychain prompts may need longer)")
	return cmd
}
