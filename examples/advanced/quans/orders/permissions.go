package orders

import shopqua "shop-qua"

var orderPermissions = []shopqua.PermissionDefinition{
	{Key: permissionCreate, Description: "Place orders"},
	{Key: permissionRead, Description: "See and cancel your own orders"},
	{Key: permissionReadAll, Description: "See everyone's orders"},
}
