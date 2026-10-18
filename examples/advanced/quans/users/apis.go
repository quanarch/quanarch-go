package users

import shopqua "shop-qua"

var userApis = []shopqua.ApiDefinition{
	{
		Endpoint:          routeRegister,
		Method:            shopqua.POST,
		Handler:           registerHandler,
		QuanarMiddlewares: []shopqua.QuanarMiddleware{shopqua.RateLimit},
		Middlewares:       []shopqua.LocalMiddleware{catchErrors},
		Validation:        registerValidation,
		Description:       "Create an account",
	},
	{
		Endpoint:          routeLogin,
		Method:            shopqua.POST,
		Handler:           loginHandler,
		QuanarMiddlewares: []shopqua.QuanarMiddleware{shopqua.RateLimit},
		Middlewares:       []shopqua.LocalMiddleware{catchErrors},
		Validation:        loginValidation,
		Description:       "Sign in and get a token",
	},
	{
		Endpoint:          routeMe,
		Method:            shopqua.GET,
		Handler:           meHandler,
		QuanarMiddlewares: []shopqua.QuanarMiddleware{shopqua.Authenticate},
		Middlewares:       []shopqua.LocalMiddleware{catchErrors},
		Description:       "The signed-in user",
	},
	{
		Endpoint:    routeList,
		Method:      shopqua.GET,
		Handler:     listUsersHandler,
		Permissions: []string{permissionRead},
		Description: "Every user",
	},
	{
		Endpoint:    routeRole,
		Method:      shopqua.PUT,
		Handler:     setRoleHandler,
		Permissions: []string{permissionManage},
		Middlewares: []shopqua.LocalMiddleware{catchErrors, audit},
		Validation:  roleValidation,
		Description: "Change a user's role",
	},
}
