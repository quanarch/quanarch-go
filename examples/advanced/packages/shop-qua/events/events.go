// Package events lets a quan announce what happened; the quanar carries it to whoever listens.
package events

import (
	"context"

	shopqua "shop-qua"
)

// Emit announces an event. Its name starts with the quan's name: orders.placed
func Emit(ctx context.Context, event string, payload any) error {
	return shopqua.Core().Events().Emit(ctx, event, payload)
}
