package http

// CreateOrderRequest represents the HTTP JSON request payload for creating an order.
type CreateOrderRequest struct {
	CustomerID  string  `json:"customer_id"`
	TotalAmount float64 `json:"total_amount"`
}
