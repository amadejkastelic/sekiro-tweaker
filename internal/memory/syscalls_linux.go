// +build linux

package memory

import "syscall"

const (
	SYS_PROCESS_VM_READV  = 310
	SYS_PROCESS_VM_WRITEV = 311
)

type iovec syscall.Iovec
