package products

import shopqua "shop-qua"

var productApis = []shopqua.ApiDefinition{
	{
		Endpoint:    routeList,
		Method:      shopqua.GET,
		Handler:     listProductsHandler,
		Validation:  listProductsValidation,
		Description: "The catalogue; managers also see inactive products",
	},
	{
		Endpoint:    routeOne,
		Method:      shopqua.GET,
		Handler:     getProductHandler,
		Middlewares: []shopqua.LocalMiddleware{loadProduct},
		Description: "One product",
	},
	{
		Endpoint:    routeList,
		Method:      shopqua.POST,
		Handler:     createProductHandler,
		Permissions: []string{permissionManage},
		Middlewares: []shopqua.LocalMiddleware{catchErrors},
		Validation:  createProductValidation,
		Description: "Add a product",
	},
	{
		Endpoint:    routeOne,
		Method:      shopqua.PATCH,
		Handler:     updateProductHandler,
		Permissions: []string{permissionManage},
		Middlewares: []shopqua.LocalMiddleware{catchErrors, loadProduct},
		Validation:  updateProductValidation,
		Description: "Change a product",
	},
	{
		Endpoint:    routeRestock,
		Method:      shopqua.POST,
		Handler:     restockProductHandler,
		Permissions: []string{permissionManage},
		Middlewares: []shopqua.LocalMiddleware{catchErrors, loadProduct},
		Validation:  restockValidation,
		Description: "Add stock",
	},
}
