// Package users is a quan of the shop: accounts, roles and sign-in. It has no main and serves
// nothing itself; the quanar runs what its Manifest declares.
package users

import shopqua "shop-qua"

// Manifest is everything this quan is. It listens to no events; it only emits users.registered.
var Manifest = shopqua.QuanManifest{
	Name:        QuanName,
	Version:     "1.0.0",
	Models:      []shopqua.ModelDefinition{userModel},
	Permissions: userPermissions,
	Apis:        userApis,
	OnMount:     ensureAdmin,
}
