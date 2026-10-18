package products

import (
	"context"
	"encoding/json"
	"slices"

	shopqua "shop-qua"
)

// Handlers take a request and return a response; they only call services.

func listProductsHandler(ctx context.Context, req *shopqua.ApiRequest) (shopqua.ApiResponse, error) {
	search, _ := req.Query["search"].(string)
	limit, _ := req.Query["limit"].(float64)
	offset, _ := req.Query["offset"].(float64)
	everything := req.User != nil && slices.Contains(req.User.Permissions, permissionManage)
	found, err := listProducts(ctx, search, int(limit), int(offset), everything)
	return shopqua.ApiResponse{Body: found}, err
}

func getProductHandler(_ context.Context, req *shopqua.ApiRequest) (shopqua.ApiResponse, error) {
	return shopqua.ApiResponse{Body: req.Locals["product"]}, nil
}

func createProductHandler(ctx context.Context, req *shopqua.ApiRequest) (shopqua.ApiResponse, error) {
	var in newProduct
	if err := json.Unmarshal(req.Body, &in); err != nil {
		return shopqua.ApiResponse{}, err
	}
	p, err := createProduct(ctx, in)
	return shopqua.ApiResponse{Status: 201, Body: p}, err
}

func updateProductHandler(ctx context.Context, req *shopqua.ApiRequest) (shopqua.ApiResponse, error) {
	var changes map[string]any
	if err := json.Unmarshal(req.Body, &changes); err != nil {
		return shopqua.ApiResponse{}, err
	}
	p, err := updateProduct(ctx, req.Params["id"], changes)
	return shopqua.ApiResponse{Body: p}, err
}

func restockProductHandler(ctx context.Context, req *shopqua.ApiRequest) (shopqua.ApiResponse, error) {
	var in struct {
		Quantity int `json:"quantity"`
	}
	if err := json.Unmarshal(req.Body, &in); err != nil {
		return shopqua.ApiResponse{}, err
	}
	p, err := restock(ctx, req.Params["id"], in.Quantity)
	return shopqua.ApiResponse{Body: p}, err
}
