// Package db is the quan's view of the quanar's database. Every call becomes one DbOperation;
// the quanar runs it against the quan's own tables and the row comes back as the quan's type.
//
//	products := db.Model[Product]("products")
//	list, err := products.Find(ctx, db.Query{Where: shopqua.Where{"active": true}})
package db

import (
	"context"
	"encoding/json"

	shopqua "shop-qua"
)

// Query narrows a Find.
type Query struct {
	Where   shopqua.Where
	OrderBy map[string]string // field → "asc" or "desc"
	Limit   int
	Offset  int
}

// Repository is one model of the quan, read into T.
type Repository[T any] struct{ name string }

// Model returns the repository of the quan's model called name.
func Model[T any](name string) Repository[T] { return Repository[T]{name} }

func run[R any](ctx context.Context, op shopqua.DbOperation) (R, error) {
	var out R
	raw, err := shopqua.Core().Db().Run(ctx, op)
	if err != nil {
		return out, err
	}
	return out, json.Unmarshal(raw, &out)
}

func (r Repository[T]) Find(ctx context.Context, q Query) ([]T, error) {
	return run[[]T](ctx, shopqua.DbOperation{Action: "find", Model: r.name, Where: q.Where, OrderBy: q.OrderBy, Limit: q.Limit, Offset: q.Offset})
}

// FindOne returns nil when no row matches.
func (r Repository[T]) FindOne(ctx context.Context, where shopqua.Where) (*T, error) {
	return run[*T](ctx, shopqua.DbOperation{Action: "find_one", Model: r.name, Where: where})
}

func (r Repository[T]) FindByID(ctx context.Context, id string) (*T, error) {
	return r.FindOne(ctx, shopqua.Where{"id": id})
}

func (r Repository[T]) Count(ctx context.Context, where shopqua.Where) (int, error) {
	return run[int](ctx, shopqua.DbOperation{Action: "count", Model: r.name, Where: where})
}

func (r Repository[T]) Create(ctx context.Context, data map[string]any) (T, error) {
	return run[T](ctx, shopqua.DbOperation{Action: "create", Model: r.name, Data: data})
}

// Update returns how many rows changed.
func (r Repository[T]) Update(ctx context.Context, where shopqua.Where, data map[string]any) (int, error) {
	return run[int](ctx, shopqua.DbOperation{Action: "update", Model: r.name, Where: where, Data: data})
}

func (r Repository[T]) UpdateByID(ctx context.Context, id string, data map[string]any) (*T, error) {
	if _, err := r.Update(ctx, shopqua.Where{"id": id}, data); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r Repository[T]) Delete(ctx context.Context, where shopqua.Where) (int, error) {
	return run[int](ctx, shopqua.DbOperation{Action: "delete", Model: r.name, Where: where})
}

// Transaction runs every operation or none.
func Transaction(ctx context.Context, ops []shopqua.DbOperation) ([]json.RawMessage, error) {
	return shopqua.Core().Db().Transaction(ctx, ops)
}

// Increment changes a number in place: {"stock": db.Increment(3)}.
func Increment(by float64) map[string]any { return map[string]any{"increment": by} }

// Decrement changes a number in place: {"stock": db.Decrement(3)}.
func Decrement(by float64) map[string]any { return map[string]any{"decrement": by} }

// Expect marks how many rows an update or delete must change.
func Expect(n int) *int { return &n }
