//go:build linux

package process

import (
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
)

var (
	sharedNetwork atomic.Bool
	warnOnce      sync.Once
)

func base() *syscall.SysProcAttr {
	// The quan dies with the quanar.
	return &syscall.SysProcAttr{Pdeathsig: syscall.SIGKILL, Setpgid: true}
}

// sysProcAttrs are the ways to start a quan, best first: its own network, an empty loopback where
// nothing it opens can be reached; then, where the system allows no such thing to an unprivileged
// user (Ubuntu 24.04 does not), the quanar's network, with Landlock still closing TCP.
func sysProcAttrs() []*syscall.SysProcAttr {
	if sharedNetwork.Load() {
		return []*syscall.SysProcAttr{base()}
	}
	own := base()
	own.Cloneflags = syscall.CLONE_NEWUSER | syscall.CLONE_NEWNET
	own.UidMappings = []syscall.SysProcIDMap{{ContainerID: os.Getuid(), HostID: os.Getuid(), Size: 1}}
	own.GidMappings = []syscall.SysProcIDMap{{ContainerID: os.Getgid(), HostID: os.Getgid(), Size: 1}}
	return []*syscall.SysProcAttr{own, base()}
}

// fellBack notes that quans start without a network of their own, and says so once.
func fellBack(err error) {
	sharedNetwork.Store(true)
	warnOnce.Do(func() {
		slog.Warn("quans start without a network of their own: the system does not allow it to this user; Landlock still closes TCP, but a port a quan opens can be reached", "reason", err)
	})
}
