// Package auth asks the quanar for identity work: hashing, checking and signing never leave it.
package auth

import (
	"context"

	shopqua "shop-qua"
)

func IssueToken(ctx context.Context, claims shopqua.TokenClaims) (shopqua.IssuedToken, error) {
	return shopqua.Core().Auth().IssueToken(ctx, claims)
}

func HashPassword(ctx context.Context, password string) (string, error) {
	return shopqua.Core().Auth().HashPassword(ctx, password)
}

func VerifyPassword(ctx context.Context, password, hashed string) (bool, error) {
	return shopqua.Core().Auth().VerifyPassword(ctx, password, hashed)
}
