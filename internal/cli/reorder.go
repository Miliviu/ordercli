package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/steipete/ordercli/internal/foodora"
)

func newReorderCmd(st *state) *cobra.Command {
	var confirm bool
	var addressID string
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "reorder <orderCode>",
		Short: "Reorder a past order (adds to cart when --confirm)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newAuthedClient(st)
			if err != nil {
				return err
			}

			orderCode := strings.TrimSpace(args[0])
			if orderCode == "" {
				return errors.New("missing order code")
			}

			// Safe default: preview only (no reorder endpoint call).
			if !confirm {
				resp, err := c.OrderHistoryByCode(cmd.Context(), foodora.OrderHistoryByCodeRequest{
					OrderCode:       orderCode,
					Include:         "order_products,order_details",
					ItemReplacement: false,
				})
				if err != nil {
					return err
				}
				if len(resp.Data.Items) == 0 {
					return errors.New("no order found")
				}

				printHistoryDetail(cmd.OutOrStdout(), resp.Data.Items[0])
				fmt.Fprintln(cmd.ErrOrStderr(), "hint: run with --confirm to call orders/{orderCode}/reorder (adds items to cart)")
				return nil
			}

			addrs, err := c.CustomerAddresses(cmd.Context())
			if err != nil {
				return err
			}
			addr, err := pickCustomerAddress(addrs.Data.Items, addressID)
			if err != nil {
				return err
			}

			resp, err := c.OrderReorder(cmd.Context(), orderCode, foodora.ReorderRequestBody{
				Address:     addr,
				ReorderTime: foodora.FormatReorderTime(time.Now()),
			})
			if err != nil {
				return err
			}

			if asJSON {
				b, _ := json.MarshalIndent(resp.Data, "", "  ")
				b = append(b, '\n')
				_, _ = cmd.OutOrStdout().Write(b)
				return nil
			}

			printReorderDetail(cmd.OutOrStdout(), resp.Data)
			fmt.Fprintln(cmd.ErrOrStderr(), "note: this only builds a cart; it does not place an order")
			return nil
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "call reorder endpoint (adds to cart)")
	cmd.Flags().StringVar(&addressID, "address-id", "", "override customer address id (safer when multiple addresses)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "print raw JSON (confirm only)")
	return cmd
}

func pickCustomerAddress(items []map[string]any, addressID string) (map[string]any, error) {
	if len(items) == 0 {
		return nil, errors.New("no customer addresses found (add one in the app/site)")
	}

	addressID = strings.TrimSpace(addressID)
	if addressID != "" {
		for _, a := range items {
			if strings.EqualFold(asString(a["id"]), addressID) {
				return a, nil
			}
		}
		return nil, fmt.Errorf("address id %q not found (available: %s)", addressID, strings.Join(addressIDs(items), ","))
	}

	if len(items) == 1 {
		return items[0], nil
	}

	for _, a := range items {
		if isTruthy(a["is_selected"]) || isTruthy(a["selected"]) || isTruthy(a["isSelected"]) {
			return a, nil
		}
	}
	for _, a := range items {
		if isTruthy(a["is_default"]) || isTruthy(a["default"]) || isTruthy(a["isDefault"]) {
			return a, nil
		}
	}

	return nil, fmt.Errorf("multiple addresses found; pass --address-id (available: %s)", strings.Join(addressIDs(items), ","))
}

func addressIDs(items []map[string]any) []string {
	out := make([]string, 0, len(items))
	seen := map[string]bool{}
	for _, a := range items {
		id := asString(a["id"])
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	if len(out) == 0 {
		return []string{"<unknown>"}
	}
	return out
}

func isTruthy(v any) bool {
	switch t := v.(type) {
	case nil:
		return false
	case bool:
		return t
	case string:
		s := strings.TrimSpace(strings.ToLower(t))
		return s == "true" || s == "1" || s == "yes"
	case float64:
		return t != 0
	case int:
		return t != 0
	default:
		return false
	}
}

func printReorderDetail(out io.Writer, d foodora.PastOrderDetails) {
	if d.VendorInfo != nil && d.VendorInfo.Name != "" {
		fmt.Fprintf(out, "vendor=%s\n", d.VendorInfo.Name)
	} else if d.VendorCode != "" {
		fmt.Fprintf(out, "vendor_code=%s\n", d.VendorCode)
	}

	if d.Cart.TotalValue > 0 {
		fmt.Fprintf(out, "total=%.2f\n", d.Cart.TotalValue)
	}

	var products []foodora.ReorderCartProduct
	for _, vc := range d.Cart.VendorCart {
		products = append(products, vc.Products...)
	}
	if len(products) == 0 {
		fmt.Fprintln(out, "items=<none>")
		return
	}

	fmt.Fprintln(out, "items:")
	for _, p := range products {
		name := strings.TrimSpace(p.Name)
		if name == "" {
			name = "<unknown>"
		}

		line := name
		if v := strings.TrimSpace(p.VariationName); v != "" && !strings.EqualFold(v, name) {
			line += " — " + v
		}

		if len(p.Toppings) > 0 {
			var tops []string
			for _, t := range p.Toppings {
				if tn := strings.TrimSpace(t.Name); tn != "" {
					tops = append(tops, tn)
				}
			}
			if len(tops) > 0 {
				line += " (" + strings.Join(tops, ", ") + ")"
			}
		}

		if !p.IsAvailable {
			if s := strings.TrimSpace(p.SoldOutOption); s != "" {
				line += " [unavailable: " + s + "]"
			} else {
				line += " [unavailable]"
			}
		}

		switch {
		case p.Quantity > 0 && p.TotalPrice > 0:
			fmt.Fprintf(out, "- %dx %s (%.2f)\n", p.Quantity, line, p.TotalPrice)
		case p.Quantity > 0:
			fmt.Fprintf(out, "- %dx %s\n", p.Quantity, line)
		case p.TotalPrice > 0:
			fmt.Fprintf(out, "- %s (%.2f)\n", line, p.TotalPrice)
		default:
			fmt.Fprintf(out, "- %s\n", line)
		}
	}
}
