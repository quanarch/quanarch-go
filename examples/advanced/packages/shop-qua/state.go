package shopqua

import "errors"

var core QuanarCore

// ErrNotInitialized is what every call meets before the quanar has plugged its core in.
var ErrNotInitialized = errors.New("shop-qua is not initialized: the quanar must call Init first")

// Init plugs the quanar's core in. Only the quanar calls it, once, before any quan runs.
func Init(c QuanarCore) error {
	if core != nil {
		return errors.New("shop-qua is already initialized")
	}
	core = c
	return nil
}

// Reset takes the core out.
func Reset() { core = nil }

// IsInitialized reports whether a core is plugged in.
func IsInitialized() bool { return core != nil }

// Core returns the plugged core, or panics: a quan cannot run before the quanar plugs it in.
func Core() QuanarCore {
	if core == nil {
		panic(ErrNotInitialized)
	}
	return core
}
