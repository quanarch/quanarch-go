package orders

import shopqua "shop-qua"

var placeOrderValidation = &shopqua.ValidationRules{Body: map[string]shopqua.Rule{
	"items": {
		Type: shopqua.RuleArray, Required: true, Min: shopqua.Num(1), Max: shopqua.Num(20),
		Items: &shopqua.Rule{Type: shopqua.RuleObject, Fields: map[string]shopqua.Rule{
			"product_id": {Type: shopqua.RuleString, Required: true},
			"quantity":   {Type: shopqua.RuleInteger, Required: true, Min: shopqua.Num(1), Max: shopqua.Num(100)},
		}},
	},
}}

var listOrdersValidation = &shopqua.ValidationRules{Query: map[string]shopqua.Rule{
	"status": {Type: shopqua.RuleEnum, Values: statuses},
}}
