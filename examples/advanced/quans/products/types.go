package products

// Product is a row of the products model.
type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	Active      bool    `json:"active"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// ReservedLine is one line of an order, priced.
type ReservedLine struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
}

// Reservation is stock taken for an order.
type Reservation struct {
	ID      string         `json:"id"`
	OrderID string         `json:"order_id"`
	Items   []ReservedLine `json:"items"`
	Total   float64        `json:"total"`
}

// OrderItem is one line of an order as it arrives.
type OrderItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

// OrderEvent is what orders.placed and orders.cancelled carry, as this quan reads it.
type OrderEvent struct {
	OrderID string      `json:"order_id"`
	Items   []OrderItem `json:"items"`
}

type newProduct struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Price       float64 `json:"price"`
	Stock       *int    `json:"stock"`
}
