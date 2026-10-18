package orders

import (
	"context"
	"encoding/json"

	shopqua "shop-qua"
)

// Handlers take a request and return a response; they only call services.

// placeOrderHandler answers 202: the order is taken, its fate is decided when products answer.
func placeOrderHandler(ctx context.Context, req *shopqua.ApiRequest) (shopqua.ApiResponse, error) {
	var in struct {
		Items []OrderItem `json:"items"`
	}
	if err := json.Unmarshal(req.Body, &in); err != nil {
		return shopqua.ApiResponse{}, err
	}
	o, err := place(ctx, req.User.ID, in.Items)
	return shopqua.ApiResponse{Status: 202, Body: o}, err
}

func listOrdersHandler(ctx context.Context, req *shopqua.ApiRequest) (shopqua.ApiResponse, error) {
	status, _ := req.Query["status"].(string)
	found, err := listOrders(ctx, req.User, status)
	return shopqua.ApiResponse{Body: found}, err
}

func getOrderHandler(_ context.Context, req *shopqua.ApiRequest) (shopqua.ApiResponse, error) {
	return shopqua.ApiResponse{Body: req.Locals["order"]}, nil
}

func cancelOrderHandler(ctx context.Context, req *shopqua.ApiRequest) (shopqua.ApiResponse, error) {
	o, err := cancel(ctx, req.Locals["order"].(Order))
	return shopqua.ApiResponse{Body: o}, err
}
