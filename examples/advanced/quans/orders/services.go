package orders

import (
	"context"
	"fmt"
	"slices"

	shopqua "shop-qua"
	"shop-qua/db"
	"shop-qua/events"
	"shop-qua/log"
)

// Orders never read the products' tables. An order is placed as pending and announced;
// products answer with an event, and the order follows.

var orders = db.Model[Order](modelOrders)

func place(ctx context.Context, userID string, items []OrderItem) (Order, error) {
	o, err := orders.Create(ctx, map[string]any{"user_id": userID, "status": string(StatusPending), "items": items})
	if err != nil {
		return Order{}, err
	}
	return o, events.Emit(ctx, eventPlaced, map[string]any{"order_id": o.ID, "items": items})
}

func listOrders(ctx context.Context, user *shopqua.AuthUser, status string) ([]Order, error) {
	where := shopqua.Where{}
	if !slices.Contains(user.Permissions, permissionReadAll) {
		where["user_id"] = user.ID
	}
	if status != "" {
		where["status"] = status
	}
	return orders.Find(ctx, db.Query{Where: where, OrderBy: map[string]string{"created_at": "desc"}})
}

func cancel(ctx context.Context, o Order) (Order, error) {
	if !slices.Contains(cancellable, any(string(o.Status))) {
		return Order{}, &appError{409, "not_cancellable", fmt.Sprintf("A %s order cannot be cancelled", o.Status)}
	}
	changed, err := orders.Update(ctx, shopqua.Where{"id": o.ID, "status": map[string]any{"in": cancellable}}, map[string]any{"status": string(StatusCancelled)})
	if err != nil {
		return Order{}, err
	}
	if changed == 0 {
		return Order{}, &appError{409, "not_cancellable", "The order changed meanwhile, look again"}
	}
	if err := events.Emit(ctx, eventCancelled, map[string]any{"order_id": o.ID, "items": o.Items}); err != nil {
		return Order{}, err
	}
	cancelled, err := orders.FindByID(ctx, o.ID)
	if err != nil || cancelled == nil {
		return Order{}, err
	}
	return *cancelled, nil
}

func confirm(ctx context.Context, e ReservedEvent) error {
	changed, err := orders.Update(ctx, shopqua.Where{"id": e.OrderID, "status": string(StatusPending)},
		map[string]any{"status": string(StatusConfirmed), "total": e.Total, "lines": e.Lines})
	if err != nil {
		return err
	}
	if changed == 0 {
		// Cancelled before the stock was taken: give it back.
		return events.Emit(ctx, eventCancelled, map[string]any{"order_id": e.OrderID, "items": e.Lines})
	}
	log.Info(fmt.Sprintf("order %s confirmed, %g", e.OrderID, e.Total), nil)
	return nil
}

func reject(ctx context.Context, e RejectedEvent) error {
	_, err := orders.Update(ctx, shopqua.Where{"id": e.OrderID, "status": string(StatusPending)},
		map[string]any{"status": string(StatusRejected), "reason": e.Reason})
	return err
}
