package orders

import shopqua "shop-qua"

var orderModel = shopqua.ModelDefinition{
	Name: modelOrders,
	Fields: map[string]shopqua.FieldDefinition{
		"user_id": {Type: shopqua.FieldString, Required: true, Index: true},
		"status":  {Type: shopqua.FieldEnum, Values: statuses, Required: true, Default: string(StatusPending)},
		"items":   {Type: shopqua.FieldJSON, Required: true},
		"lines":   {Type: shopqua.FieldJSON},
		"total":   {Type: shopqua.FieldNumber},
		"reason":  {Type: shopqua.FieldString},
	},
	Indexes: [][]string{{"user_id", "status"}},
}
