package users

const QuanName = "users"

// Routes: every endpoint starts with the quan's name.
const (
	routeRegister = "/users/register"
	routeLogin    = "/users/login"
	routeMe       = "/users/me"
	routeList     = "/users"
	routeRole     = "/users/{id}/role"
)

const (
	permissionRead   = "users:read"
	permissionManage = "users:manage"
)

// Role is what a user may be.
type Role string

const (
	RoleAdmin    Role = "admin"
	RoleCustomer Role = "customer"
)

const modelUsers = "users"

// Events this quan emits.
const eventRegistered = "users.registered"
