package products

import shopqua "shop-qua"

var productModel = shopqua.ModelDefinition{
	Name: modelProducts,
	Fields: map[string]shopqua.FieldDefinition{
		"name":        {Type: shopqua.FieldString, Required: true, Unique: true},
		"description": {Type: shopqua.FieldText},
		"price":       {Type: shopqua.FieldNumber, Required: true},
		"stock":       {Type: shopqua.FieldInteger, Required: true, Default: 0},
		"active":      {Type: shopqua.FieldBoolean, Required: true, Default: true, Index: true},
	},
}

// reservationModel is stock taken for an order. One per order, so an event delivered twice takes it once.
var reservationModel = shopqua.ModelDefinition{
	Name: modelReservations,
	Fields: map[string]shopqua.FieldDefinition{
		"order_id": {Type: shopqua.FieldString, Required: true, Unique: true},
		"items":    {Type: shopqua.FieldJSON, Required: true},
		"total":    {Type: shopqua.FieldNumber, Required: true},
	},
}
