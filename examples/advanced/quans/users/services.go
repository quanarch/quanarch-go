package users

import (
	"context"
	"strings"

	shopqua "shop-qua"
	"shop-qua/auth"
	"shop-qua/db"
	"shop-qua/events"
	"shop-qua/log"
)

var users = db.Model[User](modelUsers)

func public(u User) PublicUser {
	return PublicUser{u.ID, u.Email, u.Name, u.Role, u.CreatedAt}
}

// wrongCredentials is the same answer for a wrong email and a wrong password.
func wrongCredentials() error {
	return &appError{401, "wrong_credentials", "Email or password is wrong"}
}

func register(ctx context.Context, in registerInput) (PublicUser, error) {
	email := strings.ToLower(in.Email)
	if existing, err := users.FindOne(ctx, shopqua.Where{"email": email}); err != nil || existing != nil {
		if err != nil {
			return PublicUser{}, err
		}
		return PublicUser{}, &appError{409, "email_taken", "This email is already registered"}
	}
	hash, err := auth.HashPassword(ctx, in.Password)
	if err != nil {
		return PublicUser{}, err
	}
	u, err := users.Create(ctx, map[string]any{"email": email, "name": strings.TrimSpace(in.Name), "password_hash": hash, "role": string(RoleCustomer)})
	if err != nil {
		return PublicUser{}, err
	}
	if err := events.Emit(ctx, eventRegistered, map[string]any{"user_id": u.ID, "email": u.Email}); err != nil {
		return PublicUser{}, err
	}
	return public(u), nil
}

func login(ctx context.Context, in loginInput) (map[string]any, error) {
	u, err := users.FindOne(ctx, shopqua.Where{"email": strings.ToLower(in.Email)})
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, wrongCredentials()
	}
	ok, err := auth.VerifyPassword(ctx, in.Password, u.PasswordHash)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, wrongCredentials()
	}
	permissions := rolePermissions[u.Role]
	issued, err := auth.IssueToken(ctx, shopqua.TokenClaims{Sub: u.ID, Name: u.Name, Permissions: permissions})
	if err != nil {
		return nil, err
	}
	return map[string]any{"token": issued.Token, "expires_at": issued.ExpiresAt, "user": public(*u), "permissions": permissions}, nil
}

func me(ctx context.Context, id string) (PublicUser, error) {
	u, err := users.FindByID(ctx, id)
	if err != nil {
		return PublicUser{}, err
	}
	if u == nil {
		return PublicUser{}, &appError{404, "user_not_found", "This user no longer exists"}
	}
	return public(*u), nil
}

func listUsers(ctx context.Context) ([]PublicUser, error) {
	all, err := users.Find(ctx, db.Query{OrderBy: map[string]string{"created_at": "asc"}})
	if err != nil {
		return nil, err
	}
	out := make([]PublicUser, len(all))
	for i, u := range all {
		out[i] = public(u)
	}
	return out, nil
}

func setRole(ctx context.Context, id string, role Role, by string) (PublicUser, error) {
	if id == by {
		return PublicUser{}, &appError{409, "own_role", "You cannot change your own role"}
	}
	u, err := users.UpdateByID(ctx, id, map[string]any{"role": string(role)})
	if err != nil {
		return PublicUser{}, err
	}
	if u == nil {
		return PublicUser{}, &appError{404, "user_not_found", "No such user"}
	}
	return public(*u), nil
}

// ensureAdmin creates an admin on the very first start, when there is nobody to manage the shop.
func ensureAdmin(ctx context.Context) error {
	n, err := users.Count(ctx, nil)
	if err != nil || n > 0 {
		return err
	}
	hash, err := auth.HashPassword(ctx, defaultAdmin.Password)
	if err != nil {
		return err
	}
	if _, err := users.Create(ctx, map[string]any{"email": defaultAdmin.Email, "name": defaultAdmin.Name, "password_hash": hash, "role": string(RoleAdmin)}); err != nil {
		return err
	}
	log.Warn("created "+defaultAdmin.Email+" with the default password; change it", nil)
	return nil
}
