package orders

// OrderItem is one line of an order, as the customer sent it.
type OrderItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

// OrderLine is one line of a confirmed order, priced.
type OrderLine struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
}

// Order is a row of the orders model.
type Order struct {
	ID        string      `json:"id"`
	UserID    string      `json:"user_id"`
	Status    Status      `json:"status"`
	Items     []OrderItem `json:"items"`
	Lines     []OrderLine `json:"lines"`
	Total     *float64    `json:"total"`
	Reason    *string     `json:"reason"`
	CreatedAt string      `json:"created_at"`
	UpdatedAt string      `json:"updated_at"`
}

// ReservedEvent is what products.reserved carries, as this quan reads it.
type ReservedEvent struct {
	OrderID string      `json:"order_id"`
	Total   float64     `json:"total"`
	Lines   []OrderLine `json:"lines"`
}

// RejectedEvent is what products.rejected carries.
type RejectedEvent struct {
	OrderID string `json:"order_id"`
	Reason  string `json:"reason"`
}
