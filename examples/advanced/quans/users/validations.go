package users

import shopqua "shop-qua"

var registerValidation = &shopqua.ValidationRules{Body: map[string]shopqua.Rule{
	"email":    {Type: shopqua.RuleEmail, Required: true, Max: shopqua.Num(120)},
	"name":     {Type: shopqua.RuleString, Required: true, Min: shopqua.Num(2), Max: shopqua.Num(60)},
	"password": {Type: shopqua.RuleString, Required: true, Min: shopqua.Num(8), Max: shopqua.Num(100)},
}}

var loginValidation = &shopqua.ValidationRules{Body: map[string]shopqua.Rule{
	"email":    {Type: shopqua.RuleEmail, Required: true},
	"password": {Type: shopqua.RuleString, Required: true, Max: shopqua.Num(100)},
}}

var roleValidation = &shopqua.ValidationRules{Body: map[string]shopqua.Rule{
	"role": {Type: shopqua.RuleEnum, Values: []string{string(RoleAdmin), string(RoleCustomer)}, Required: true},
}}
