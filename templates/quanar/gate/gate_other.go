//go:build !linux

package gate

import "errors"

// Walled reports whether quans get walls on this operating system.
const Walled = false

func run([]string) error { return errors.New("the gate needs Linux") }
