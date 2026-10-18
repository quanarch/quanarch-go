package orders

const QuanName = "orders"

const (
	routeList   = "/orders"
	routeOne    = "/orders/{id}"
	routeCancel = "/orders/{id}/cancel"
)

const (
	permissionCreate  = "orders:create"
	permissionRead    = "orders:read"
	permissionReadAll = "orders:read-all"
)

const modelOrders = "orders"

// Status is where an order is.
type Status string

const (
	StatusPending   Status = "pending"
	StatusConfirmed Status = "confirmed"
	StatusRejected  Status = "rejected"
	StatusCancelled Status = "cancelled"
)

var statuses = []string{string(StatusPending), string(StatusConfirmed), string(StatusRejected), string(StatusCancelled)}

// cancellable: an order can be cancelled until it is closed.
var cancellable = []any{string(StatusPending), string(StatusConfirmed)}

// Events this quan emits.
const (
	eventPlaced    = "orders.placed"
	eventCancelled = "orders.cancelled"
)

// Events of other quans this quan listens to.
const (
	listensReserved = "products.reserved"
	listensRejected = "products.rejected"
)
