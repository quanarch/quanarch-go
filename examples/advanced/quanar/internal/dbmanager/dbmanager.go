// Package dbmanager runs the operations quans send, against their own tables only. Every value
// goes in as a parameter, every name is checked against the model, booleans and json come back
// as they were stored.
package dbmanager

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"shop-quanar/internal/schema"
	"shop-quanar/internal/wire"
)

const maxLimit = 500

var operators = map[string]string{"eq": "=", "ne": "!=", "gt": ">", "gte": ">=", "lt": "<", "lte": "<=", "like": "LIKE"}

// Error is a problem with an operation itself: the quan sent something its model does not allow.
type Error struct{ msg string }

func (e *Error) Error() string { return e.msg }

func fail(format string, args ...any) error { return &Error{fmt.Sprintf(format, args...)} }

// Operation is one database operation, as a quan sends it.
type Operation struct {
	Action  string            `json:"action"`
	Model   string            `json:"model"`
	Where   map[string]any    `json:"where"`
	OrderBy map[string]string `json:"order_by"`
	Limit   int               `json:"limit"`
	Offset  int               `json:"offset"`
	Data    map[string]any    `json:"data"`
	Expect  *int              `json:"expect"`
}

// Row is a row as it goes back to the quan.
type Row = map[string]any

type executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// Manager runs operations.
type Manager struct {
	db     *sql.DB
	schema *schema.Manager
}

// New makes a db manager.
func New(db *sql.DB, s *schema.Manager) *Manager { return &Manager{db, s} }

func now() string { return time.Now().UTC().Format("2006-01-02T15:04:05.000Z") }

func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func field(m wire.Model, name string) (wire.Field, error) {
	f, ok := schema.Field(m, name)
	if !ok {
		return f, fail("%s has no field %s", m.Name, name)
	}
	return f, nil
}

func encode(m wire.Model, name string, value any) (any, error) {
	if value == nil {
		return nil, nil
	}
	f, err := field(m, name)
	if err != nil {
		return nil, err
	}
	where := m.Name + "." + name
	switch f.Type {
	case "boolean":
		b, ok := value.(bool)
		if !ok {
			return nil, fail("%s must be true or false", where)
		}
		if b {
			return 1, nil
		}
		return 0, nil
	case "integer", "number":
		n, ok := value.(float64)
		if !ok {
			return nil, fail("%s must be a number", where)
		}
		if f.Type == "integer" {
			if n != math.Trunc(n) {
				return nil, fail("%s must be a whole number", where)
			}
			return int64(n), nil
		}
		return n, nil
	case "json":
		b, err := json.Marshal(value)
		return string(b), err
	case "enum":
		s, ok := value.(string)
		if !ok || !slices.Contains(f.Values, s) {
			return nil, fail("%s must be one of %s", where, strings.Join(f.Values, ", "))
		}
		return s, nil
	default:
		s, ok := value.(string)
		if !ok {
			return nil, fail("%s must be text", where)
		}
		return s, nil
	}
}

func decode(m wire.Model, columns []string, values []any) Row {
	row := Row{}
	for i, name := range columns {
		v := values[i]
		f, ok := schema.Field(m, name)
		switch {
		case v == nil || !ok:
			row[name] = v
		case f.Type == "boolean":
			n, _ := v.(int64)
			row[name] = n == 1
		case f.Type == "json":
			var parsed any
			s, _ := v.(string)
			_ = json.Unmarshal([]byte(s), &parsed)
			row[name] = parsed
		default:
			row[name] = v
		}
	}
	return row
}

func whereSQL(m wire.Model, where map[string]any, params *[]any) (string, error) {
	var parts []string
	names := make([]string, 0, len(where))
	for name := range where {
		names = append(names, name)
	}
	slices.Sort(names) // the same operation is always the same SQL
	for _, name := range names {
		condition := where[name]
		if _, err := field(m, name); err != nil {
			return "", err
		}
		col := schema.Quote(name)
		ops, isOps := condition.(map[string]any)
		switch {
		case condition == nil:
			parts = append(parts, col+" IS NULL")
		case isOps:
			for op, value := range ops {
				switch {
				case op == "in":
					list, _ := value.([]any)
					if len(list) == 0 {
						parts = append(parts, "0")
						continue
					}
					marks := make([]string, len(list))
					for i, v := range list {
						enc, err := encode(m, name, v)
						if err != nil {
							return "", err
						}
						marks[i] = "?"
						*params = append(*params, enc)
					}
					parts = append(parts, col+" IN ("+strings.Join(marks, ", ")+")")
				case operators[op] != "":
					parts = append(parts, col+" "+operators[op]+" ?")
					if op == "like" {
						*params = append(*params, fmt.Sprint(value))
					} else {
						enc, err := encode(m, name, value)
						if err != nil {
							return "", err
						}
						*params = append(*params, enc)
					}
				default:
					return "", fail("unknown operator %s", op)
				}
			}
		default:
			enc, err := encode(m, name, condition)
			if err != nil {
				return "", err
			}
			parts = append(parts, col+" = ?")
			*params = append(*params, enc)
		}
	}
	if len(parts) == 0 {
		return "", nil
	}
	return " WHERE " + strings.Join(parts, " AND "), nil
}

func query(ctx context.Context, ex executor, m wire.Model, sqlText string, params []any) ([]Row, error) {
	rows, err := ex.QueryContext(ctx, sqlText, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	out := []Row{}
	for rows.Next() {
		values := make([]any, len(columns))
		pointers := make([]any, len(columns))
		for i := range values {
			pointers[i] = &values[i]
		}
		if err := rows.Scan(pointers...); err != nil {
			return nil, err
		}
		out = append(out, decode(m, columns, values))
	}
	return out, rows.Err()
}

func (d *Manager) find(ctx context.Context, ex executor, m wire.Model, table string, op Operation) ([]Row, error) {
	var params []any
	where, err := whereSQL(m, op.Where, &params)
	if err != nil {
		return nil, err
	}
	sqlText := "SELECT * FROM " + schema.Quote(table) + where
	if len(op.OrderBy) > 0 {
		var terms []string
		for name, dir := range op.OrderBy {
			if _, err := field(m, name); err != nil {
				return nil, err
			}
			if dir == "desc" {
				terms = append(terms, schema.Quote(name)+" DESC")
			} else {
				terms = append(terms, schema.Quote(name)+" ASC")
			}
		}
		sqlText += " ORDER BY " + strings.Join(terms, ", ")
	}
	limit := op.Limit
	if limit <= 0 || limit > maxLimit {
		limit = maxLimit
	}
	params = append(params, limit, op.Offset)
	return query(ctx, ex, m, sqlText+" LIMIT ? OFFSET ?", params)
}

func (d *Manager) execute(ctx context.Context, ex executor, owner string, op Operation) (any, error) {
	m, err := d.schema.Model(owner, op.Model)
	if err != nil {
		return nil, fail("%s", err)
	}
	table := schema.Table(owner, op.Model)
	switch op.Action {
	case "find":
		return d.find(ctx, ex, m, table, op)
	case "find_one":
		op.Limit = 1
		rows, err := d.find(ctx, ex, m, table, op)
		if err != nil || len(rows) == 0 {
			return nil, err
		}
		return rows[0], nil
	case "count":
		var params []any
		where, err := whereSQL(m, op.Where, &params)
		if err != nil {
			return nil, err
		}
		rows, err := ex.QueryContext(ctx, "SELECT COUNT(*) FROM "+schema.Quote(table)+where, params...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var n int
		rows.Next()
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		return n, nil
	case "create":
		data := map[string]any{}
		for k, v := range op.Data {
			data[k] = v
		}
		for name, f := range m.Fields {
			if data[name] == nil && f.Default != nil {
				data[name] = f.Default
			}
			if f.Required && data[name] == nil {
				return nil, fail("%s.%s is required", m.Name, name)
			}
		}
		stamp := now()
		data["id"], data["created_at"], data["updated_at"] = newID(), stamp, stamp
		var names, marks []string
		var params []any
		for name, value := range data {
			enc, err := encode(m, name, value)
			if err != nil {
				return nil, err
			}
			names, marks, params = append(names, schema.Quote(name)), append(marks, "?"), append(params, enc)
		}
		if _, err := ex.ExecContext(ctx, fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", schema.Quote(table), strings.Join(names, ", "), strings.Join(marks, ", ")), params...); err != nil {
			return nil, err
		}
		rows, err := query(ctx, ex, m, "SELECT * FROM "+schema.Quote(table)+` WHERE "id" = ?`, []any{data["id"]})
		if err != nil || len(rows) == 0 {
			return nil, err
		}
		return rows[0], nil
	case "update":
		var sets []string
		var params []any
		for name, value := range op.Data {
			if schema.Reserved[name] {
				return nil, fail("%s cannot be changed", name)
			}
			if change, ok := value.(map[string]any); ok && (change["increment"] != nil || change["decrement"] != nil) {
				f, err := field(m, name)
				if err != nil {
					return nil, err
				}
				if f.Type != "integer" && f.Type != "number" {
					return nil, fail("%s.%s is not a number", m.Name, name)
				}
				by, _ := change["increment"].(float64)
				if d, ok := change["decrement"].(float64); ok {
					by = -d
				}
				sets, params = append(sets, schema.Quote(name)+" = "+schema.Quote(name)+" + ?"), append(params, by)
				continue
			}
			enc, err := encode(m, name, value)
			if err != nil {
				return nil, err
			}
			sets, params = append(sets, schema.Quote(name)+" = ?"), append(params, enc)
		}
		sets, params = append(sets, `"updated_at" = ?`), append(params, now())
		where, err := whereSQL(m, op.Where, &params)
		if err != nil {
			return nil, err
		}
		result, err := ex.ExecContext(ctx, "UPDATE "+schema.Quote(table)+" SET "+strings.Join(sets, ", ")+where, params...)
		if err != nil {
			return nil, err
		}
		changed, _ := result.RowsAffected()
		if op.Expect != nil && int(changed) != *op.Expect {
			return nil, fail("expected %d %s to change, %d did", *op.Expect, m.Name, changed)
		}
		return changed, nil
	case "delete":
		var params []any
		where, err := whereSQL(m, op.Where, &params)
		if err != nil {
			return nil, err
		}
		result, err := ex.ExecContext(ctx, "DELETE FROM "+schema.Quote(table)+where, params...)
		if err != nil {
			return nil, err
		}
		changed, _ := result.RowsAffected()
		if op.Expect != nil && int(changed) != *op.Expect {
			return nil, fail("expected %d %s to be deleted, %d were", *op.Expect, m.Name, changed)
		}
		return changed, nil
	}
	return nil, fail("unknown action %s", op.Action)
}

// explain turns SQLite's messages into ones a quan can act on.
func explain(owner string, err error) error {
	var own *Error
	if errors.As(err, &own) {
		return err
	}
	msg := err.Error()
	if i := strings.Index(msg, "UNIQUE constraint failed: "); i >= 0 {
		column := msg[strings.LastIndex(msg, ".")+1:]
		column = strings.TrimRight(column, ")")
		return fail("%s must be unique", column)
	}
	return fail("%s: %s", owner, msg)
}

// Run runs one operation.
func (d *Manager) Run(ctx context.Context, owner string, op Operation) (any, error) {
	v, err := d.execute(ctx, d.db, owner, op)
	if err != nil {
		return nil, explain(owner, err)
	}
	return v, nil
}

// Transaction runs every operation or none.
func (d *Manager) Transaction(ctx context.Context, owner string, ops []Operation) ([]any, error) {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	results := make([]any, 0, len(ops))
	for _, op := range ops {
		v, err := d.execute(ctx, tx, owner, op)
		if err != nil {
			_ = tx.Rollback()
			return nil, explain(owner, err)
		}
		results = append(results, v)
	}
	return results, tx.Commit()
}
