package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/steipete/foodoracli/internal/foodora"
	"github.com/steipete/foodoracli/internal/version"
	"golang.org/x/term"
)

func newLoginCmd(st *state) *cobra.Command {
	var email string
	var password string
	var passwordStdin bool
	var otpMethod string
	var otp string
	var mfaToken string
	var clientSecret string
	var storeClientSecret bool

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login via oauth2/token (email + password; optional MFA)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if st.cfg.BaseURL == "" {
				return errors.New("missing base_url (run `foodoracli config set --country HU` or similar)")
			}
			if email == "" {
				return errors.New("--email required")
			}

			if clientSecret == "" {
				clientSecret = st.clientSecretOrEnv()
			}
			if clientSecret == "" {
				return errors.New("missing client secret (pass --client-secret or set FOODORA_CLIENT_SECRET; optionally persist via --store-client-secret)")
			}

			if passwordStdin {
				b, err := bufio.NewReader(os.Stdin).ReadBytes('\n')
				if err != nil && !errors.Is(err, io.EOF) {
					return err
				}
				password = strings.TrimSpace(string(b))
			}
			if password == "" {
				fmt.Fprint(cmd.ErrOrStderr(), "Password: ")
				b, err := term.ReadPassword(int(os.Stdin.Fd()))
				fmt.Fprintln(cmd.ErrOrStderr())
				if err != nil {
					return err
				}
				password = strings.TrimSpace(string(b))
			}
			if password == "" {
				return errors.New("empty password")
			}

			c, err := foodora.New(foodora.Options{
				BaseURL:          st.cfg.BaseURL,
				DeviceID:         st.cfg.DeviceID,
				GlobalEntityID:   st.cfg.GlobalEntityID,
				TargetCountryISO: st.cfg.TargetCountryISO,
				UserAgent:        "foodoracli/" + version.Version,
			})
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			tok, mfa, err := c.OAuthTokenPassword(ctx, foodora.OAuthPasswordRequest{
				Username:     email,
				Password:     password,
				ClientSecret: clientSecret,
				OTPMethod:    otpMethod,
				OTPCode:      otp,
				MfaToken:     mfaToken,
			})
			if err != nil {
				return err
			}

			if mfa != nil && tok.AccessToken == "" {
				fmt.Fprintf(cmd.ErrOrStderr(), "MFA triggered (%s). Check your %s. Retry with:\n", mfa.Channel, mfa.Channel)
				fmt.Fprintf(cmd.ErrOrStderr(), "  foodoracli login --email %s --otp-method %s --mfa-token %s --otp <CODE>\n", email, mfa.Channel, mfa.MfaToken)
				fmt.Fprintf(cmd.ErrOrStderr(), "rate limit reset: %ds\n", mfa.RateLimitReset)
				return nil
			}

			now := time.Now()
			st.cfg.AccessToken = tok.AccessToken
			st.cfg.RefreshToken = tok.RefreshToken
			st.cfg.ExpiresAt = tok.ExpiresAt(now)
			if storeClientSecret {
				st.cfg.ClientSecret = clientSecret
			}
			st.markDirty()
			fmt.Fprintln(cmd.OutOrStdout(), "ok")
			return nil
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "account email")
	cmd.Flags().StringVar(&password, "password", "", "password (discouraged; prefer --password-stdin or prompt)")
	cmd.Flags().BoolVar(&passwordStdin, "password-stdin", false, "read password from stdin (first line)")
	cmd.Flags().StringVar(&otpMethod, "otp-method", "sms", "OTP/MFA channel (e.g. sms, call)")
	cmd.Flags().StringVar(&mfaToken, "mfa-token", "", "MFA token from a prior mfa_triggered response")
	cmd.Flags().StringVar(&otp, "otp", "", "OTP code for MFA verification")
	cmd.Flags().StringVar(&clientSecret, "client-secret", "", "oauth client secret (see FOODORA_CLIENT_SECRET)")
	cmd.Flags().BoolVar(&storeClientSecret, "store-client-secret", false, "persist client secret into config file")
	return cmd
}

func newLogoutCmd(st *state) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Forget stored tokens",
		Run: func(cmd *cobra.Command, args []string) {
			st.cfg.AccessToken = ""
			st.cfg.RefreshToken = ""
			st.cfg.ExpiresAt = time.Time{}
			st.markDirty()
			fmt.Fprintln(cmd.OutOrStdout(), "ok")
		},
	}
}
