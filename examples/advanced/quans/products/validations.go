package products

import shopqua "shop-qua"

var listProductsValidation = &shopqua.ValidationRules{Query: map[string]shopqua.Rule{
	"search": {Type: shopqua.RuleString, Max: shopqua.Num(60)},
	"limit":  {Type: shopqua.RuleInteger, Min: shopqua.Num(1), Max: shopqua.Num(100)},
	"offset": {Type: shopqua.RuleInteger, Min: shopqua.Num(0)},
}}

var createProductValidation = &shopqua.ValidationRules{Body: map[string]shopqua.Rule{
	"name":        {Type: shopqua.RuleString, Required: true, Min: shopqua.Num(2), Max: shopqua.Num(80)},
	"description": {Type: shopqua.RuleString, Max: shopqua.Num(1000)},
	"price":       {Type: shopqua.RuleNumber, Required: true, Min: shopqua.Num(0.01), Max: shopqua.Num(100000)},
	"stock":       {Type: shopqua.RuleInteger, Min: shopqua.Num(0), Max: shopqua.Num(100000)},
}}

var updateProductValidation = &shopqua.ValidationRules{Body: map[string]shopqua.Rule{
	"name":        {Type: shopqua.RuleString, Min: shopqua.Num(2), Max: shopqua.Num(80)},
	"description": {Type: shopqua.RuleString, Max: shopqua.Num(1000)},
	"price":       {Type: shopqua.RuleNumber, Min: shopqua.Num(0.01), Max: shopqua.Num(100000)},
	"active":      {Type: shopqua.RuleBoolean},
}}

var restockValidation = &shopqua.ValidationRules{Body: map[string]shopqua.Rule{
	"quantity": {Type: shopqua.RuleInteger, Required: true, Min: shopqua.Num(1), Max: shopqua.Num(10000)},
}}
