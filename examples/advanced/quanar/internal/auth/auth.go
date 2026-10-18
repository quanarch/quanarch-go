// Package auth is identity, without knowing any user. A quan proves who someone is however it
// likes and asks for a token; the quanar signs it and, on every request, reads it back.
package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"golang.org/x/crypto/scrypt"

	"shop-quanar/internal/wire"
)

// Claims are what a quan asks to put in a token.
type Claims struct {
	Sub         string   `json:"sub"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

// Issued is a signed token and when it expires.
type Issued struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
}

type payload struct {
	Claims
	Iat int64 `json:"iat"`
	Exp int64 `json:"exp"`
}

// Manager signs and reads tokens.
type Manager struct {
	secret []byte
	ttl    time.Duration
}

var header = b64([]byte(`{"alg":"HS256","typ":"JWT"}`))

func b64(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }

// New makes a manager; without a secret it makes one, and tokens do not survive a restart.
func New(secret string, ttl time.Duration) *Manager {
	key := []byte(secret)
	if secret == "" {
		key = make([]byte, 32)
		_, _ = rand.Read(key)
		slog.Warn("JWT_SECRET is not set: a random secret is used and tokens will not survive a restart", "scope", "auth")
	}
	return &Manager{key, ttl}
}

func (a *Manager) sign(data string) string {
	mac := hmac.New(sha256.New, a.secret)
	mac.Write([]byte(data))
	return b64(mac.Sum(nil))
}

// Issue signs a token for claims.
func (a *Manager) Issue(c Claims) Issued {
	now := time.Now()
	exp := now.Add(a.ttl)
	body, _ := json.Marshal(payload{c, now.Unix(), exp.Unix()})
	data := header + "." + b64(body)
	return Issued{Token: data + "." + a.sign(data), ExpiresAt: exp.UTC().Format("2006-01-02T15:04:05Z")}
}

// Verify returns the user behind a token, or nil when it is missing, forged or expired.
func (a *Manager) Verify(token string) *wire.User {
	parts := strings.Split(token, ".")
	if len(parts) != 3 || !hmac.Equal([]byte(a.sign(parts[0]+"."+parts[1])), []byte(parts[2])) {
		return nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil
	}
	var p payload
	if json.Unmarshal(raw, &p) != nil || p.Exp < time.Now().Unix() {
		return nil
	}
	if p.Permissions == nil {
		p.Permissions = []string{}
	}
	return &wire.User{ID: p.Sub, Name: p.Name, Permissions: p.Permissions}
}

// scrypt's cost: about 50 ms and 16 MB per password.
func derive(password string, salt []byte) ([]byte, error) {
	return scrypt.Key([]byte(password), salt, 1<<14, 8, 1, 64)
}

// HashPassword hashes a password with a fresh salt.
func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	_, _ = rand.Read(salt)
	key, err := derive(password, salt)
	if err != nil {
		return "", err
	}
	return "scrypt$" + hex.EncodeToString(salt) + "$" + hex.EncodeToString(key), nil
}

// VerifyPassword checks a password against a stored hash.
func VerifyPassword(password, stored string) bool {
	parts := strings.Split(stored, "$")
	if len(parts) != 3 || parts[0] != "scrypt" {
		return false
	}
	salt, err1 := hex.DecodeString(parts[1])
	want, err2 := hex.DecodeString(parts[2])
	if err1 != nil || err2 != nil {
		return false
	}
	got, err := derive(password, salt)
	return err == nil && subtle.ConstantTimeCompare(got, want) == 1
}
