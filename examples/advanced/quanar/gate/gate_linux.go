//go:build linux

package gate

// On Linux the gate closes the walls with Landlock and a memory ceiling. Both outlive exec, so
// not a single line of the quan, not even its init functions, runs outside them.

import (
	"flag"
	"fmt"
	"syscall"

	"github.com/landlock-lsm/go-landlock/landlock"
)

// Walled reports whether quans get walls on this operating system.
const Walled = true

// run closes the walls and execs the quan. It only returns on failure.
func run(args []string) error {
	flags := flag.NewFlagSet(Command, flag.ContinueOnError)
	memoryMB := flags.Int("memory-mb", 256, "memory ceiling")
	callTimeout := flags.String("call-timeout", "5s", "how long the quan waits for the quanar")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 1 {
		return fmt.Errorf("usage: %s [flags] <quan program>", Command)
	}
	program := flags.Arg(0)

	size := uint64(*memoryMB) << 20
	// RLIMIT_DATA, not RLIMIT_AS: Go reserves a large address space up front and only counts
	// what it writes; the address space ceiling would stop it before it starts.
	if err := syscall.Setrlimit(syscall.RLIMIT_DATA, &syscall.Rlimit{Cur: size, Max: size}); err != nil {
		return fmt.Errorf("memory ceiling: %w", err)
	}
	// Files: only reading and running the quan's own program. Network: no TCP at all.
	// Signals and abstract sockets: only within the quan's own walls.
	if err := landlock.V9.BestEffort().Restrict(landlock.ROFiles(program)); err != nil {
		return fmt.Errorf("walls: %w", err)
	}
	return syscall.Exec(program, []string{program, "--call-timeout", *callTimeout}, []string{})
}
