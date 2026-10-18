package products

import (
	"context"
	"errors"

	shopqua "shop-qua"
)

// The quan's own middlewares. They run in the quan, after the quanar's checks.

func failure(req *shopqua.ApiRequest, status int, code, message string) shopqua.ApiResponse {
	return shopqua.ApiResponse{Status: status, Body: map[string]any{"error": map[string]string{"code": code, "message": message}, "request_id": req.RequestID}}
}

// catchErrors turns the failures this quan expects into answers; anything else is the quanar's 500.
func catchErrors(_ context.Context, req *shopqua.ApiRequest, next shopqua.Next) (shopqua.ApiResponse, error) {
	res, err := next()
	var e *appError
	if errors.As(err, &e) {
		return failure(req, e.status, e.code, e.message), nil
	}
	return res, err
}

// loadProduct finds the product of the path's {id} for the handler, or answers 404 before it runs.
func loadProduct(ctx context.Context, req *shopqua.ApiRequest, next shopqua.Next) (shopqua.ApiResponse, error) {
	p, err := products.FindByID(ctx, req.Params["id"])
	if err != nil {
		return shopqua.ApiResponse{}, err
	}
	if p == nil {
		return failure(req, 404, "product_not_found", "No such product"), nil
	}
	req.Locals["product"] = *p
	return next()
}
