package users

import shopqua "shop-qua"

var userPermissions = []shopqua.PermissionDefinition{
	{Key: permissionRead, Description: "See every user"},
	{Key: permissionManage, Description: "Change what a user may do"},
}
