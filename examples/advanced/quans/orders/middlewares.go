package orders

import (
	"context"
	"errors"
	"slices"

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

// ownOrder loads the order of the path's {id} for the handler. Someone else's order answers 404,
// not 403: whether it exists is not told.
func ownOrder(ctx context.Context, req *shopqua.ApiRequest, next shopqua.Next) (shopqua.ApiResponse, error) {
	o, err := orders.FindByID(ctx, req.Params["id"])
	if err != nil {
		return shopqua.ApiResponse{}, err
	}
	if o == nil || (o.UserID != req.User.ID && !slices.Contains(req.User.Permissions, permissionReadAll)) {
		return failure(req, 404, "order_not_found", "No such order"), nil
	}
	req.Locals["order"] = *o
	return next()
}
