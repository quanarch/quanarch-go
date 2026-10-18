// Package gate is the door a quan's program goes through. The quanar starts itself as the gate;
// the gate closes the walls and only then becomes the quan's program. Where the operating system
// gives no such walls, the quanar starts the quan's program directly.
package gate

import (
	"fmt"
	"os"
)

// Command is the hidden argument that turns the quanar's program into the gate.
const Command = "__quan-gate"

// Main is for the quanar's main: it runs the gate when asked and reports whether it did.
func Main() bool {
	if len(os.Args) < 2 || os.Args[1] != Command {
		return false
	}
	if err := run(os.Args[2:]); err != nil {
		fmt.Fprintln(os.Stderr, "gate:", err)
		os.Exit(1)
	}
	return true
}
