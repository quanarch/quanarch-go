package users

import shopqua "shop-qua"

var userModel = shopqua.ModelDefinition{
	Name: modelUsers,
	Fields: map[string]shopqua.FieldDefinition{
		"email":         {Type: shopqua.FieldString, Required: true, Unique: true},
		"name":          {Type: shopqua.FieldString, Required: true},
		"password_hash": {Type: shopqua.FieldString, Required: true},
		"role":          {Type: shopqua.FieldEnum, Values: []string{string(RoleAdmin), string(RoleCustomer)}, Required: true, Default: string(RoleCustomer), Index: true},
	},
}
