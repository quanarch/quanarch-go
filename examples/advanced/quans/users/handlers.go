package users

import (
	"context"
	"encoding/json"

	shopqua "shop-qua"
)

// Handlers take a request and return a response; they only call services.

func registerHandler(ctx context.Context, req *shopqua.ApiRequest) (shopqua.ApiResponse, error) {
	var in registerInput
	if err := json.Unmarshal(req.Body, &in); err != nil {
		return shopqua.ApiResponse{}, err
	}
	u, err := register(ctx, in)
	return shopqua.ApiResponse{Status: 201, Body: u}, err
}

func loginHandler(ctx context.Context, req *shopqua.ApiRequest) (shopqua.ApiResponse, error) {
	var in loginInput
	if err := json.Unmarshal(req.Body, &in); err != nil {
		return shopqua.ApiResponse{}, err
	}
	session, err := login(ctx, in)
	return shopqua.ApiResponse{Body: session}, err
}

func meHandler(ctx context.Context, req *shopqua.ApiRequest) (shopqua.ApiResponse, error) {
	u, err := me(ctx, req.User.ID)
	return shopqua.ApiResponse{Body: u}, err
}

func listUsersHandler(ctx context.Context, _ *shopqua.ApiRequest) (shopqua.ApiResponse, error) {
	all, err := listUsers(ctx)
	return shopqua.ApiResponse{Body: all}, err
}

func setRoleHandler(ctx context.Context, req *shopqua.ApiRequest) (shopqua.ApiResponse, error) {
	var in struct {
		Role Role `json:"role"`
	}
	if err := json.Unmarshal(req.Body, &in); err != nil {
		return shopqua.ApiResponse{}, err
	}
	u, err := setRole(ctx, req.Params["id"], in.Role, req.User.ID)
	return shopqua.ApiResponse{Body: u}, err
}
