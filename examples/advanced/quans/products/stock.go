package products

import (
	"context"
	"math"
	"strings"

	shopqua "shop-qua"
	"shop-qua/db"
	"shop-qua/events"
	"shop-qua/log"
)

// Stock is taken for orders and given back. Orders never touch this quan's tables: they say an
// order was placed, and hear back whether it could be served.

var reservations = db.Model[Reservation](modelReservations)

func reserve(ctx context.Context, order OrderEvent) error {
	if existing, err := reservations.FindOne(ctx, shopqua.Where{"order_id": order.OrderID}); err != nil || existing != nil {
		return err
	}
	reject := func(reason string) error {
		return events.Emit(ctx, eventRejected, map[string]any{"order_id": order.OrderID, "reason": reason})
	}

	// The same product asked twice in one order is one line.
	var ids []any
	wanted := map[string]int{}
	for _, item := range order.Items {
		if _, seen := wanted[item.ProductID]; !seen {
			ids = append(ids, item.ProductID)
		}
		wanted[item.ProductID] += item.Quantity
	}
	found, err := products.Find(ctx, db.Query{Where: shopqua.Where{"id": map[string]any{"in": ids}, "active": true}})
	if err != nil {
		return err
	}
	byID := map[string]Product{}
	for _, p := range found {
		byID[p.ID] = p
	}
	var lines []ReservedLine
	total := 0.0
	for _, id := range ids {
		p, ok := byID[id.(string)]
		if !ok {
			return reject("product " + id.(string) + " is not sold")
		}
		line := ReservedLine{ProductID: p.ID, Name: p.Name, Price: p.Price, Quantity: wanted[p.ID]}
		lines = append(lines, line)
		total += line.Price * float64(line.Quantity)
	}
	total = math.Round(total*100) / 100

	// Every line or none: an update that finds less stock than asked changes no row, and the batch rolls back.
	var ops []shopqua.DbOperation
	for _, l := range lines {
		ops = append(ops, shopqua.DbOperation{
			Action: "update", Model: modelProducts,
			Where:  shopqua.Where{"id": l.ProductID, "stock": map[string]any{"gte": l.Quantity}},
			Data:   map[string]any{"stock": db.Decrement(float64(l.Quantity))},
			Expect: db.Expect(1),
		})
	}
	ops = append(ops, shopqua.DbOperation{Action: "create", Model: modelReservations, Data: map[string]any{"order_id": order.OrderID, "items": lines, "total": total}})
	if _, err := db.Transaction(ctx, ops); err != nil {
		var short []string
		for _, l := range lines {
			if byID[l.ProductID].Stock < l.Quantity {
				short = append(short, l.Name)
			}
		}
		if len(short) > 0 {
			return reject("not enough " + strings.Join(short, ", "))
		}
		return reject("stock changed, try again")
	}

	if err := events.Emit(ctx, eventReserved, map[string]any{"order_id": order.OrderID, "total": total, "lines": lines}); err != nil {
		return err
	}
	low, err := products.Find(ctx, db.Query{Where: shopqua.Where{"id": map[string]any{"in": ids}, "stock": map[string]any{"lt": lowStock}}})
	if err != nil {
		return err
	}
	for _, p := range low {
		if err := events.Emit(ctx, eventStockLow, map[string]any{"product_id": p.ID, "name": p.Name, "stock": p.Stock}); err != nil {
			return err
		}
	}
	return nil
}

func release(ctx context.Context, order OrderEvent) error {
	r, err := reservations.FindOne(ctx, shopqua.Where{"order_id": order.OrderID})
	if err != nil || r == nil {
		return err
	}
	var ops []shopqua.DbOperation
	for _, l := range r.Items {
		ops = append(ops, shopqua.DbOperation{Action: "update", Model: modelProducts, Where: shopqua.Where{"id": l.ProductID}, Data: map[string]any{"stock": db.Increment(float64(l.Quantity))}})
	}
	ops = append(ops, shopqua.DbOperation{Action: "delete", Model: modelReservations, Where: shopqua.Where{"order_id": order.OrderID}, Expect: db.Expect(1)})
	if _, err := db.Transaction(ctx, ops); err != nil {
		return err
	}
	log.Info("gave back the stock of order "+order.OrderID, nil)
	return nil
}
