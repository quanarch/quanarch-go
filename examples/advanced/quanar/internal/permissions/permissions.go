// Package permissions keeps the permissions quans declare. The quanar knows what each one means,
// not who holds it: that comes in the token a quan asked for.
package permissions

import (
	"fmt"
	"slices"
	"strings"
	"sync"

	"shop-quanar/internal/wire"
)

// Manager keeps every quan's permissions.
type Manager struct {
	mu       sync.RWMutex
	declared map[string][]wire.Permission
}

// New makes a permission manager.
func New() *Manager { return &Manager{declared: map[string][]wire.Permission{}} }

// Register keeps a quan's permissions; each starts with the quan's name.
func (p *Manager) Register(owner string, permissions []wire.Permission) error {
	for _, perm := range permissions {
		if !strings.HasPrefix(perm.Key, owner+":") {
			return fmt.Errorf("%s may not declare %s: permissions start with %s:", owner, perm.Key, owner)
		}
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.declared[owner] = permissions
	return nil
}

// Remove forgets a quan's permissions.
func (p *Manager) Remove(owner string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.declared, owner)
}

// Missing returns the required permissions the user lacks.
func Missing(user *wire.User, required []string) []string {
	missing := []string{}
	for _, r := range required {
		if !slices.Contains(user.Permissions, r) {
			missing = append(missing, r)
		}
	}
	return missing
}

// All lists every declared permission.
func (p *Manager) All() []wire.Permission {
	p.mu.RLock()
	defer p.mu.RUnlock()
	all := []wire.Permission{}
	for _, perms := range p.declared {
		all = append(all, perms...)
	}
	return all
}
