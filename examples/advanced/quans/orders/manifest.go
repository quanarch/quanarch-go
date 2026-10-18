// Package orders is a quan of the shop. It never reads the products' tables: it announces an
// order and follows what products answer. It has no main and serves nothing itself; the quanar
// runs what its Manifest declares.
package orders

import shopqua "shop-qua"

// Manifest is everything this quan is.
var Manifest = shopqua.QuanManifest{
	Name:        QuanName,
	Version:     "1.0.0",
	Models:      []shopqua.ModelDefinition{orderModel},
	Permissions: orderPermissions,
	Apis:        orderApis,
	Events:      productEvents,
}
