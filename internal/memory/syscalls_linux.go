//go:build linux
// +build linux

package memory

const (
	SYS_PROCESS_VM_READV  = 310
	SYS_PROCESS_VM_WRITEV = 311
)
