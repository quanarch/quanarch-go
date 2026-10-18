package products

import (
	"context"
	"encoding/json"

	shopqua "shop-qua"
)

// What this quan does when orders announce something.

func onOrder(do func(context.Context, OrderEvent) error) shopqua.EventHandler {
	return func(ctx context.Context, payload json.RawMessage, _ shopqua.EventMeta) error {
		var order OrderEvent
		if err := json.Unmarshal(payload, &order); err != nil {
			return err
		}
		return do(ctx, order)
	}
}

var orderEvents = []shopqua.EventSubscription{
	{Event: listensOrderPlaced, Handler: onOrder(reserve)},
	{Event: listensOrderCancelled, Handler: onOrder(release)},
}
