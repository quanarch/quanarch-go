package orders

import (
	"context"
	"encoding/json"

	shopqua "shop-qua"
)

// What this quan does when products answer.

func on[T any](do func(context.Context, T) error) shopqua.EventHandler {
	return func(ctx context.Context, payload json.RawMessage, _ shopqua.EventMeta) error {
		var e T
		if err := json.Unmarshal(payload, &e); err != nil {
			return err
		}
		return do(ctx, e)
	}
}

var productEvents = []shopqua.EventSubscription{
	{Event: listensReserved, Handler: on(confirm)},
	{Event: listensRejected, Handler: on(reject)},
}
