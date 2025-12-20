package deliveroo

import "testing"

func TestOrderSummary(t *testing.T) {
	t.Parallel()

	total := 12.5
	o := Order{
		ID:                  "id",
		OrderNumber:         "n",
		Status:              "delivered",
		Restaurant:          &Restaurant{Name: "R"},
		Total:               &total,
		CurrencySymbol:      "€",
		EstimatedDeliveryAt: "2025-12-20T01:00:00Z",
		SubmittedAt:         "2025-12-20T00:00:00Z",
	}
	got := o.Summary()
	if got == "order" {
		t.Fatalf("unexpected: %q", got)
	}
}
