// Package config holds the quanar's settings and reads the list of quans it loads.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// QuanSource says where a quan lives and which quas are plugged into it.
type QuanSource struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	URL     string `json:"url"`
	// Events is a stream that announces a new build with a change event.
	Events string `json:"events"`
	// Hash is the SHA-256 of the program. When set, a program that does not match never runs.
	Hash string   `json:"hash"`
	Quas []string `json:"quas"`
}

// Settings of the quanar.
type Settings struct {
	Port int
	// QuansFile lists the quans: name, program, hash, event stream, quas. Written by the build.
	QuansFile    string
	DatabaseFile string
	// JWTSecret signs the tokens. Without it a random one is made and tokens do not survive a restart.
	JWTSecret       string
	TokenTTL        time.Duration
	RateLimitWindow time.Duration
	RateLimitMax    int
	BodyLimit       int64
	Reconnect       time.Duration
	FetchTimeout    time.Duration
	LoadTimeout     time.Duration
	RequestTimeout  time.Duration
	// CallTimeout is how long a quan waits for the quanar to answer one of its calls.
	CallTimeout  time.Duration
	QuanMemoryMB int
}

// Load reads the settings from the environment.
func Load() Settings {
	data := env("DATA_DIR", "data")
	return Settings{
		Port:            envInt("PORT", 3000),
		QuansFile:       env("QUANS_FILE", filepath.Join(data, "quans.json")),
		DatabaseFile:    env("DATABASE_FILE", filepath.Join(data, "shop.db")),
		JWTSecret:       os.Getenv("JWT_SECRET"),
		TokenTTL:        time.Duration(envInt("TOKEN_TTL_SECONDS", 8*60*60)) * time.Second,
		RateLimitWindow: time.Minute,
		RateLimitMax:    envInt("RATE_LIMIT_MAX", 30),
		BodyLimit:       100 << 10,
		Reconnect:       2 * time.Second,
		FetchTimeout:    3 * time.Second,
		LoadTimeout:     5 * time.Second,
		RequestTimeout:  10 * time.Second,
		CallTimeout:     5 * time.Second,
		QuanMemoryMB:    256,
	}
}

// QuanSources reads the quans this quanar loads. The list lives outside the code.
func (s Settings) QuanSources() []QuanSource {
	var sources []QuanSource
	if b, err := os.ReadFile(s.QuansFile); err == nil {
		_ = json.Unmarshal(b, &sources)
	}
	return sources
}

func env(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return fallback
}

func envInt(name string, fallback int) int {
	if v, err := strconv.Atoi(os.Getenv(name)); err == nil {
		return v
	}
	return fallback
}
