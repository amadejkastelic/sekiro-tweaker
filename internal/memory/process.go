package memory

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"github.com/amadejkastelic/sekiro-tweaker/internal/logger"
	"go.uber.org/zap"
)

type ProcessMemory struct {
	PID               int
	allocationOffsets map[int64]int64
}

func NewProcessMemory(pid int) *ProcessMemory {
	return &ProcessMemory{
		PID:               pid,
		allocationOffsets: make(map[int64]int64),
	}
}

func (pm *ProcessMemory) ReadMemory(address int64, size int) ([]byte, error) {
	buf := make([]byte, size)

	local := syscall.Iovec{
		Base: &buf[0],
		Len:  uint64(size),
	}

	remote := syscall.Iovec{
		Base: (*byte)(unsafe.Pointer(uintptr(address))),
		Len:  uint64(size),
	}

	n, _, errno := syscall.Syscall6(
		SYS_PROCESS_VM_READV,
		uintptr(pm.PID),
		uintptr(unsafe.Pointer(&local)),
		1,
		uintptr(unsafe.Pointer(&remote)),
		1,
		0,
	)

	if errno != 0 {
		return nil, fmt.Errorf("process_vm_readv failed: %v", errno)
	}

	if int(n) != size {
		return nil, fmt.Errorf("partial read: got %d/%d bytes", n, size)
	}

	return buf, nil
}

func (pm *ProcessMemory) WriteMemory(address int64, data []byte) error {
	size := len(data)

	local := syscall.Iovec{
		Base: &data[0],
		Len:  uint64(size),
	}

	remote := syscall.Iovec{
		Base: (*byte)(unsafe.Pointer(uintptr(address))),
		Len:  uint64(size),
	}

	n, _, errno := syscall.Syscall6(
		SYS_PROCESS_VM_WRITEV,
		uintptr(pm.PID),
		uintptr(unsafe.Pointer(&local)),
		1,
		uintptr(unsafe.Pointer(&remote)),
		1,
		0,
	)

	if errno == 0 && int(n) == size {
		return nil
	}

	if errno == syscall.EFAULT {
		return pm.writeMemoryViaProcMem(address, data)
	}

	return fmt.Errorf("process_vm_writev failed: %v", errno)
}

func (pm *ProcessMemory) writeMemoryViaProcMem(address int64, data []byte) error {
	memPath := fmt.Sprintf("/proc/%d/mem", pm.PID)
	f, err := os.OpenFile(memPath, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("failed to open %s: %v", memPath, err)
	}
	defer f.Close()

	_, err = f.Seek(address, 0)
	if err != nil {
		return fmt.Errorf("seek failed: %v", err)
	}

	n, err := f.Write(data)
	if err != nil {
		return fmt.Errorf("write failed: %v", err)
	}

	if n != len(data) {
		return fmt.Errorf("partial write: %d/%d bytes", n, len(data))
	}

	return nil
}

type MemoryRegion struct {
	Start       int64
	End         int64
	Permissions string
	Path        string
}

func (mr *MemoryRegion) IsReadable() bool {
	return len(mr.Permissions) > 0 && mr.Permissions[0] == 'r'
}

func (mr *MemoryRegion) IsWritable() bool {
	return len(mr.Permissions) > 1 && mr.Permissions[1] == 'w'
}

func (mr *MemoryRegion) IsExecutable() bool {
	return len(mr.Permissions) > 2 && mr.Permissions[2] == 'x'
}

func (pm *ProcessMemory) ParseMemoryMaps() ([]MemoryRegion, error) {
	mapsPath := fmt.Sprintf("/proc/%d/maps", pm.PID)
	f, err := os.Open(mapsPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var regions []MemoryRegion
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		addrRange := strings.Split(fields[0], "-")
		if len(addrRange) != 2 {
			continue
		}

		start, err := strconv.ParseInt(addrRange[0], 16, 64)
		if err != nil {
			continue
		}

		end, err := strconv.ParseInt(addrRange[1], 16, 64)
		if err != nil {
			continue
		}

		permissions := fields[1]

		path := ""
		if len(fields) >= 6 {
			path = strings.Join(fields[5:], " ")
		}

		regions = append(regions, MemoryRegion{
			Start:       start,
			End:         end,
			Permissions: permissions,
			Path:        path,
		})
	}

	return regions, scanner.Err()
}

func (pm *ProcessMemory) GetModuleBaseAddress(moduleName string) (int64, error) {
	regions, err := pm.ParseMemoryMaps()
	if err != nil {
		return 0, fmt.Errorf("failed to parse maps: %v", err)
	}

	searchSuffix := strings.ToLower(moduleName) + ".exe"
	logger.Log.Debug("Searching for module",
		zap.String("module", moduleName),
		zap.String("suffix", searchSuffix),
		zap.Int("regions", len(regions)))

	for _, region := range regions {
		if region.Path != "" && strings.HasSuffix(strings.ToLower(region.Path), searchSuffix) {
			logger.Log.Info("Found module",
				zap.String("module", moduleName),
				zap.String("address", fmt.Sprintf("0x%X", region.Start)),
				zap.String("path", region.Path))
			return region.Start, nil
		}
	}

	for _, region := range regions {
		if region.Path != "" && strings.Contains(strings.ToLower(region.Path), strings.ToLower(moduleName)) {
			logger.Log.Info("Found module (contains match)",
				zap.String("module", moduleName),
				zap.String("address", fmt.Sprintf("0x%X", region.Start)),
				zap.String("path", region.Path))
			return region.Start, nil
		}
	}

	return 0, fmt.Errorf("module %s not found (searched %d regions)", moduleName, len(regions))
}

func (pm *ProcessMemory) GetModuleSize(moduleName string) (int, error) {
	baseAddress, err := pm.GetModuleBaseAddress(moduleName)
	if err != nil {
		return 0, err
	}

	regions, err := pm.ParseMemoryMaps()
	if err != nil {
		return 0, err
	}

	for _, region := range regions {
		if region.IsExecutable() && region.Start >= baseAddress && region.Start < baseAddress+0x10000 {
			return int(region.End - baseAddress), nil
		}
	}

	return 0, fmt.Errorf("executable section not found for %s", moduleName)
}

func (pm *ProcessMemory) AllocateMemory(nearAddress int64, size int) (int64, error) {
	regions, err := pm.ParseMemoryMaps()
	if err != nil {
		return 0, err
	}

	minAddress := nearAddress - 0x70000000
	maxAddress := nearAddress + 0x70000000

	var candidates []MemoryRegion
	for _, region := range regions {
		if region.IsWritable() &&
			!region.IsExecutable() &&
			region.Start >= minAddress &&
			region.Start < maxAddress &&
			(region.End-region.Start) >= int64(size) {
			candidates = append(candidates, region)
		}
	}

	if len(candidates) == 0 {
		return 0, fmt.Errorf("no suitable writable regions found")
	}

	region := candidates[0]
	regionStart := region.Start

	offset, exists := pm.allocationOffsets[regionStart]
	if !exists {
		offset = 0x1000
	}

	allocationAddress := regionStart + offset

	alignedSize := int64((size + 15) &^ 15)
	pm.allocationOffsets[regionStart] = offset + alignedSize

	return allocationAddress, nil
}

func FindProcessByName(name string) ([]int, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}

	var exactMatches []int
	var cmdlineMatches []int

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		commPath := fmt.Sprintf("/proc/%d/comm", pid)
		commData, err := os.ReadFile(commPath)
		if err != nil {
			continue
		}

		comm := strings.TrimSpace(string(commData))

		if strings.EqualFold(comm, name) || strings.EqualFold(comm, name+".exe") {
			logger.Log.Debug("Found exact process match",
				zap.Int("pid", pid),
				zap.String("comm", comm))
			exactMatches = append(exactMatches, pid)
			continue
		}

		cmdlinePath := fmt.Sprintf("/proc/%d/cmdline", pid)
		cmdlineData, err := os.ReadFile(cmdlinePath)
		if err != nil {
			continue
		}

		cmdline := string(cmdlineData)
		if strings.Contains(strings.ToLower(cmdline), strings.ToLower(name)+".exe") {
			cmdlineMatches = append(cmdlineMatches, pid)
		}
	}

	var result []int
	result = append(result, exactMatches...)
	result = append(result, cmdlineMatches...)

	if len(result) > 0 {
		logger.Log.Info("Found processes",
			zap.String("name", name),
			zap.Int("exact_matches", len(exactMatches)),
			zap.Int("cmdline_matches", len(cmdlineMatches)),
			zap.Ints("exact_pids", exactMatches))
	} else {
		logger.Log.Debug("No processes found", zap.String("name", name))
	}

	return result, nil
}

func (pm *ProcessMemory) ReadInt32(address int64) (int32, error) {
	data, err := pm.ReadMemory(address, 4)
	if err != nil {
		return 0, err
	}
	return int32(data[0]) | int32(data[1])<<8 | int32(data[2])<<16 | int32(data[3])<<24, nil
}

func (pm *ProcessMemory) ReadInt64(address int64) (int64, error) {
	data, err := pm.ReadMemory(address, 8)
	if err != nil {
		return 0, err
	}
	return int64(data[0]) | int64(data[1])<<8 | int64(data[2])<<16 | int64(data[3])<<24 |
		int64(data[4])<<32 | int64(data[5])<<40 | int64(data[6])<<48 | int64(data[7])<<56, nil
}

func (pm *ProcessMemory) ReadFloat32(address int64) (float32, error) {
	data, err := pm.ReadMemory(address, 4)
	if err != nil {
		return 0, err
	}
	bits := uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
	return *(*float32)(unsafe.Pointer(&bits)), nil
}

func (pm *ProcessMemory) WriteFloat32(address int64, value float32) error {
	bits := *(*uint32)(unsafe.Pointer(&value))
	data := []byte{
		byte(bits),
		byte(bits >> 8),
		byte(bits >> 16),
		byte(bits >> 24),
	}
	return pm.WriteMemory(address, data)
}

// DereferenceStaticPointer dereferences a static x64 pointer (RIP-relative addressing)
// This reads the instruction at instructionAddr and calculates the final address
func (pm *ProcessMemory) DereferenceStaticPointer(instructionAddr int64, instructionLength int) (int64, error) {
	// Read the 4-byte offset (at instructionAddr + 3 for most mov instructions)
	offsetData, err := pm.ReadMemory(instructionAddr+3, 4)
	if err != nil {
		return 0, err
	}
	offset := int32(offsetData[0]) | int32(offsetData[1])<<8 | int32(offsetData[2])<<16 | int32(offsetData[3])<<24

	// RIP-relative: address = instructionAddr + instructionLength + offset
	targetAddr := instructionAddr + int64(instructionLength) + int64(offset)

	// Read the pointer at the target address
	return pm.ReadInt64(targetAddr)
}
