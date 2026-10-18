// Package schema turns the models a quan declares into tables and keeps them in step: a new
// model becomes a table, a new field a column, a unique or indexed field an index. Nothing is
// ever dropped, so a quan going away or coming back never loses data.
package schema

import (
	"database/sql"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"sync"

	"shop-quanar/internal/wire"
)

var name = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// Reserved fields are added by the quanar to every model.
var Reserved = map[string]bool{"id": true, "created_at": true, "updated_at": true}

var column = map[string]string{
	"string": "TEXT", "text": "TEXT", "enum": "TEXT", "datetime": "TEXT", "json": "TEXT",
	"integer": "INTEGER", "boolean": "INTEGER", "number": "REAL",
}

// Manager keeps every quan's models.
type Manager struct {
	db       *sql.DB
	mu       sync.RWMutex
	registry map[string]map[string]wire.Model // owner → model name → model
}

// New makes a schema manager over db.
func New(db *sql.DB) *Manager {
	return &Manager{db: db, registry: map[string]map[string]wire.Model{}}
}

// Quote quotes a checked name.
func Quote(n string) string { return `"` + n + `"` }

// Table is where a quan's model lives: every quan's tables share the database, each under its prefix.
func Table(owner, model string) string {
	return strings.ReplaceAll(owner, "-", "_") + "__" + model
}

func check(owner string, m wire.Model) error {
	if !name.MatchString(m.Name) {
		return fmt.Errorf("%s: model name %q must be lowercase letters, digits and _", owner, m.Name)
	}
	for field, spec := range m.Fields {
		where := owner + "." + m.Name + "." + field
		switch {
		case Reserved[field]:
			return fmt.Errorf("%s is added by the quanar", where)
		case !name.MatchString(field):
			return fmt.Errorf("%s: field names are lowercase letters, digits and _", where)
		case column[spec.Type] == "":
			return fmt.Errorf("%s: unknown type %s", where, spec.Type)
		case spec.Type == "enum" && len(spec.Values) == 0:
			return fmt.Errorf("%s: an enum needs values", where)
		}
	}
	for _, index := range m.Indexes {
		for _, field := range index {
			if _, ok := m.Fields[field]; !ok && !Reserved[field] {
				return fmt.Errorf("%s.%s: index over unknown field %s", owner, m.Name, field)
			}
		}
	}
	return nil
}

// Sync creates and migrates a quan's tables.
func (s *Manager) Sync(owner string, models []wire.Model) error {
	for _, m := range models {
		if err := check(owner, m); err != nil {
			return err
		}
	}
	var changes []string
	for _, m := range models {
		table := Table(owner, m.Name)
		existing, err := s.columns(table)
		if err != nil {
			return err
		}
		if len(existing) == 0 {
			var cols strings.Builder
			for field, spec := range m.Fields {
				fmt.Fprintf(&cols, "%s %s, ", Quote(field), column[spec.Type])
			}
			if _, err := s.db.Exec(fmt.Sprintf(`CREATE TABLE %s ("id" TEXT PRIMARY KEY, %s"created_at" TEXT NOT NULL, "updated_at" TEXT NOT NULL)`, Quote(table), cols.String())); err != nil {
				return err
			}
			changes = append(changes, "created "+table)
		} else {
			for field, spec := range m.Fields {
				if existing[field] {
					continue
				}
				if _, err := s.db.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", Quote(table), Quote(field), column[spec.Type])); err != nil {
					return err
				}
				changes = append(changes, "added "+table+"."+field)
			}
		}
		type index struct {
			fields []string
			unique bool
		}
		var indexes []index
		for field, spec := range m.Fields {
			if spec.Unique || spec.Index {
				indexes = append(indexes, index{[]string{field}, spec.Unique})
			}
		}
		for _, fields := range m.Indexes {
			indexes = append(indexes, index{fields, false})
		}
		for _, ix := range indexes {
			n := table + "__" + strings.Join(ix.fields, "_")
			unique := ""
			if ix.unique {
				n, unique = n+"__unique", "UNIQUE "
			}
			quoted := make([]string, len(ix.fields))
			for i, f := range ix.fields {
				quoted[i] = Quote(f)
			}
			if _, err := s.db.Exec(fmt.Sprintf("CREATE %sINDEX IF NOT EXISTS %s ON %s (%s)", unique, Quote(n), Quote(table), strings.Join(quoted, ", "))); err != nil {
				return err
			}
		}
	}
	s.mu.Lock()
	s.registry[owner] = map[string]wire.Model{}
	for _, m := range models {
		s.registry[owner][m.Name] = m
	}
	s.mu.Unlock()
	if len(changes) > 0 {
		slog.Info(owner+": "+strings.Join(changes, ", "), "scope", "schema")
	}
	return nil
}

func (s *Manager) columns(table string) (map[string]bool, error) {
	rows, err := s.db.Query(fmt.Sprintf("PRAGMA table_info(%s)", Quote(table)))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	have := map[string]bool{}
	for rows.Next() {
		var (
			cid     int
			n, kind string
			notnull int
			dflt    any
			pk      int
		)
		if err := rows.Scan(&cid, &n, &kind, &notnull, &dflt, &pk); err != nil {
			return nil, err
		}
		have[n] = true
	}
	return have, rows.Err()
}

// Remove forgets a quan's models. Its tables and rows stay.
func (s *Manager) Remove(owner string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.registry, owner)
}

// Model is a model of the quan, or an error: a quan can only reach its own tables.
func (s *Manager) Model(owner, model string) (wire.Model, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.registry[owner][model]
	if !ok {
		return wire.Model{}, fmt.Errorf("%s has no model %s", owner, model)
	}
	return m, nil
}

// Field is a field of the model, with the reserved ones.
func Field(m wire.Model, field string) (wire.Field, bool) {
	if Reserved[field] {
		return wire.Field{Type: "string"}, true
	}
	f, ok := m.Fields[field]
	return f, ok
}

// Describe lists every quan's models.
func (s *Manager) Describe() map[string][]wire.Model {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := map[string][]wire.Model{}
	for owner, models := range s.registry {
		for _, m := range models {
			out[owner] = append(out[owner], m)
		}
	}
	return out
}
