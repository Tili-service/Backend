package tpe

type TPE struct {
	ID     *string `json:"id,omitempty"`
	Name   *string `json:"name,omitempty"`
	Status *string `json:"status,omitempty"`
}

type TotalAmount struct {
	Value     int64  `json:"value"`
	Currency  string `json:"currency"`
	MinorUnit int    `json:"minor_unit"`
}

type TPEPaymentRequest struct {
	TotalAmount TotalAmount `json:"total_amount"`
	Description string      `json:"description,omitempty"`
	TPEID       string      `json:"tpe_id,omitempty"`
}

type ReaderCheckoutData struct {
	ID              string       `json:"id"`
	Status          string       `json:"status"`
	TransactionCode string       `json:"transaction_code,omitempty"`
	TransactionID   string       `json:"transaction_id,omitempty"`
	TotalAmount     *TotalAmount `json:"total_amount,omitempty"`
}

type ReaderCheckoutResponse struct {
	Data ReaderCheckoutData `json:"data"`
}
