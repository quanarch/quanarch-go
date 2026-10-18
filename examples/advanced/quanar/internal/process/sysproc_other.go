//go:build !linux

package process

import "syscall"

func sysProcAttrs() []*syscall.SysProcAttr { return []*syscall.SysProcAttr{nil} }

func fellBack(error) {}
