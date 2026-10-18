// Package products is a quan of the shop: the catalogue and its stock. It has no main and serves
// nothing itself; the quanar runs what its Manifest declares.
package products

import shopqua "shop-qua"

// Manifest is everything this quan is.
var Manifest = shopqua.QuanManifest{
	Name:        QuanName,
	Version:     "1.0.0",
	Models:      []shopqua.ModelDefinition{productModel, reservationModel},
	Permissions: productPermissions,
	Apis:        productApis,
	Events:      orderEvents,
}
