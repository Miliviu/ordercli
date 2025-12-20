package foodora

import "time"

type AuthToken struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

func (t AuthToken) ExpiresAt(now time.Time) time.Time {
	if t.ExpiresIn <= 0 {
		return time.Time{}
	}
	return now.Add(time.Duration(t.ExpiresIn) * time.Second)
}

type OAuthPasswordRequest struct {
	Username     string
	Password     string
	ClientSecret string
	ClientID     string
	OTPMethod    string
	OTPCode      string
	MfaToken     string
}

type OAuthRefreshRequest struct {
	RefreshToken string
	ClientSecret string
	ClientID     string
}

type ActiveOrdersResponse struct {
	Status int              `json:"status"`
	Data   ActiveOrdersData `json:"data"`
}

type ActiveOrdersData struct {
	Count         int           `json:"count"`
	ActiveOrders  []ActiveOrder `json:"active_orders"`
	PollInSeconds *int          `json:"poll_in_sec"`
}

type ActiveOrder struct {
	Code        string            `json:"code"`
	IsDelivered bool              `json:"is_delivered"`
	Vendor      ActiveOrderVendor `json:"vendor"`
	Status      StatusMessages    `json:"status_messages"`
}

type ActiveOrderVendor struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type StatusMessages struct {
	Subtitle string        `json:"subtitle"`
	Titles   []StatusTitle `json:"titles"`
}

type StatusTitle struct {
	Name   string `json:"name"`
	Active bool   `json:"active"`
	Filled bool   `json:"is_filled"`
}

type OrderStatusResponse struct {
	Status int            `json:"status"`
	Data   map[string]any `json:"data"`
}

type OrderHistoryRequest struct {
	Include        string
	Offset         int
	Limit          int
	PandaGoEnabled bool
}

type OrderHistoryResponse struct {
	Status int              `json:"status"`
	Data   OrderHistoryData `json:"data"`
}

type OrderHistoryData struct {
	TotalCount FlexibleInt        `json:"total_count"`
	Items      []OrderHistoryItem `json:"items"`
}

type OrderHistoryItem struct {
	OrderCode             string              `json:"order_code"`
	CurrentStatus         *OrderHistoryStatus `json:"current_status"`
	ConfirmedDeliveryTime *OrderHistoryTime   `json:"confirmed_delivery_time"`
	Vendor                *OrderHistoryVendor `json:"vendor"`
	TotalValue            float64             `json:"total_value"`
}

type OrderHistoryVendor struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type OrderHistoryStatus struct {
	Code               FlexibleString `json:"code"`
	Message            string         `json:"message"`
	InternalStatusCode FlexibleString `json:"internal_status_code"`
}

type OrderHistoryTime struct {
	Date     FlexibleTime `json:"date"`
	Timezone string       `json:"timezone"`
}

type OrderHistoryByCodeRequest struct {
	OrderCode       string
	Include         string
	ItemReplacement bool
}

type OrderHistoryRawResponse struct {
	Status int                 `json:"status"`
	Data   OrderHistoryRawData `json:"data"`
}

type OrderHistoryRawData struct {
	TotalCount FlexibleInt      `json:"total_count"`
	Items      []map[string]any `json:"items"`
}

type CustomerAddressesResponse struct {
	Status int                   `json:"status"`
	Data   CustomerAddressesData `json:"data"`
}

type CustomerAddressesData struct {
	Items []map[string]any `json:"items"`
}

type ReorderRequestBody struct {
	Address     map[string]any `json:"address"`
	ReorderTime string         `json:"reorder_time"`
}

type OrderReorderResponse struct {
	Status int              `json:"status"`
	Data   PastOrderDetails `json:"data"`
}

type PastOrderDetails struct {
	Cart       ReorderCart        `json:"cart"`
	VendorID   int                `json:"vendor_id"`
	VendorCode string             `json:"vendor_code"`
	VendorInfo *ReorderVendorInfo `json:"vendor_info"`
}

type ReorderVendorInfo struct {
	Name     string `json:"name"`
	Vertical string `json:"vertical"`
	TimeZone string `json:"time_zone"`
}

type ReorderCart struct {
	TotalValue float64             `json:"total_value"`
	VendorCart []ReorderVendorCart `json:"vendor_cart"`
}

type ReorderVendorCart struct {
	Products []ReorderCartProduct `json:"products"`
}

type ReorderCartProduct struct {
	Name                string           `json:"name"`
	VariationName       string           `json:"variation_name"`
	Quantity            int              `json:"quantity"`
	TotalPrice          float64          `json:"total_price"`
	Price               float64          `json:"price"`
	IsAvailable         bool             `json:"is_available"`
	SoldOutOption       string           `json:"sold_out_option"`
	SpecialInstructions string           `json:"special_instructions"`
	Toppings            []ReorderTopping `json:"toppings"`
}

type ReorderTopping struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	IsAvailable bool    `json:"is_available"`
	Position    string  `json:"position"`
	Price       float64 `json:"price"`
}
