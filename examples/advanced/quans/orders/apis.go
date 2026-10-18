package orders

import shopqua "shop-qua"

var orderApis = []shopqua.ApiDefinition{
	{
		Endpoint:          routeList,
		Method:            shopqua.POST,
		Handler:           placeOrderHandler,
		Permissions:       []string{permissionCreate},
		QuanarMiddlewares: []shopqua.QuanarMiddleware{shopqua.RateLimit},
		Validation:        placeOrderValidation,
		Description:       "Place an order; it is confirmed or rejected once products answer",
	},
	{
		Endpoint:    routeList,
		Method:      shopqua.GET,
		Handler:     listOrdersHandler,
		Permissions: []string{permissionRead},
		Validation:  listOrdersValidation,
		Description: "Your orders, or everyone's with orders:read-all",
	},
	{
		Endpoint:    routeOne,
		Method:      shopqua.GET,
		Handler:     getOrderHandler,
		Permissions: []string{permissionRead},
		Middlewares: []shopqua.LocalMiddleware{ownOrder},
		Description: "One order",
	},
	{
		Endpoint:    routeCancel,
		Method:      shopqua.POST,
		Handler:     cancelOrderHandler,
		Permissions: []string{permissionRead},
		Middlewares: []shopqua.LocalMiddleware{catchErrors, ownOrder},
		Description: "Cancel an order; products give the stock back",
	},
}
