package users

// defaultAdmin is created on first start when there is no user at all, so the shop can be managed.
var defaultAdmin = struct{ Email, Name, Password string }{"admin@shop.test", "Admin", "admin1234"}

// rolePermissions says what each role may do, across the shop. Other quans declare their
// permissions; deciding who holds them is this quan's business, so the list lives here.
var rolePermissions = map[Role][]string{
	RoleAdmin:    {permissionRead, permissionManage, "products:manage", "orders:create", "orders:read", "orders:read-all"},
	RoleCustomer: {"orders:create", "orders:read"},
}
