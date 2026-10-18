package users

import (
	"context"
	"errors"

	shopqua "shop-qua"
	"shop-qua/log"
)

// The quan's own middlewares. They run in the quan, after the quanar's checks.

// catchErrors turns the failures this quan expects into answers; anything else is the quanar's 500.
func catchErrors(_ context.Context, req *shopqua.ApiRequest, next shopqua.Next) (shopqua.ApiResponse, error) {
	res, err := next()
	var e *appError
	if errors.As(err, &e) {
		return shopqua.ApiResponse{Status: e.status, Body: map[string]any{"error": map[string]string{"code": e.code, "message": e.message}, "request_id": req.RequestID}}, nil
	}
	return res, err
}

// audit writes down who changed what, after the change went through.
func audit(_ context.Context, req *shopqua.ApiRequest, next shopqua.Next) (shopqua.ApiResponse, error) {
	res, err := next()
	if err == nil && (res.Status == 0 || res.Status < 400) {
		who := "someone"
		if req.User != nil {
			who = req.User.Name
		}
		log.Info(who+" "+req.Method+" "+req.Path, req.Body)
	}
	return res, err
}
