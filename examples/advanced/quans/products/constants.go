package products

const QuanName = "products"

// lowStock is the stock below which the quan says the stock is low.
const lowStock = 5

const (
	routeList    = "/products"
	routeOne     = "/products/{id}"
	routeRestock = "/products/{id}/restock"
)

const permissionManage = "products:manage"

const (
	modelProducts     = "products"
	modelReservations = "reservations"
)

// Events this quan emits.
const (
	eventReserved = "products.reserved"
	eventRejected = "products.rejected"
	eventStockLow = "products.stock-low"
)

// Events of other quans this quan listens to. It knows their names, never their code.
const (
	listensOrderPlaced    = "orders.placed"
	listensOrderCancelled = "orders.cancelled"
)
