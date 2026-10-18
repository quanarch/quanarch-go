// Package store is the quanar's database. It knows no tables in advance: the quan names them.
// Kept in memory; replace this package to use a real database.
package store

import (
	"context"
	"crypto/rand"
	"fmt"
	"maps"
	"sync"
)

// Row is one stored row, with its id.
type Row = map[string]any

var (
	mu     sync.Mutex
	tables = map[string]map[string]Row{}
	order  = map[string][]string{}
)

func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func Create(_ context.Context, table string, data map[string]any) (Row, error) {
	mu.Lock()
	defer mu.Unlock()
	row := maps.Clone(data)
	if row == nil {
		row = Row{}
	}
	row["id"] = newID()
	if tables[table] == nil {
		tables[table] = map[string]Row{}
	}
	tables[table][row["id"].(string)] = row
	order[table] = append(order[table], row["id"].(string))
	return maps.Clone(row), nil
}

func List(_ context.Context, table string) ([]Row, error) {
	mu.Lock()
	defer mu.Unlock()
	rows := []Row{}
	for _, id := range order[table] {
		if row, ok := tables[table][id]; ok {
			rows = append(rows, maps.Clone(row))
		}
	}
	return rows, nil
}

func Get(_ context.Context, table, id string) (Row, error) {
	mu.Lock()
	defer mu.Unlock()
	return maps.Clone(tables[table][id]), nil
}

func Update(_ context.Context, table, id string, data map[string]any) (Row, error) {
	mu.Lock()
	defer mu.Unlock()
	row, ok := tables[table][id]
	if !ok {
		return nil, nil
	}
	maps.Copy(row, data)
	row["id"] = id
	return maps.Clone(row), nil
}

func Remove(_ context.Context, table, id string) (bool, error) {
	mu.Lock()
	defer mu.Unlock()
	_, ok := tables[table][id]
	delete(tables[table], id)
	return ok, nil
}
