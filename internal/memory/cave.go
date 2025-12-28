package memory

import (
	"encoding/binary"
	"fmt"
)

type PointerStyle int

const (
	DWordRelative PointerStyle = iota
)

type DataCave struct {
	name           string
	pointerAddress int64
	caveAddress    int64
	data           []byte
	pointerStyle   PointerStyle
}

type CodeCave struct {
	name              string
	injectionAddress  int64
	caveAddress       int64
	shellcode         []byte
	overwriteLength   int
	originalBytes     []byte
	active            bool
}

type CaveManager struct {
	memory      *ProcessMemory
	baseAddress int64
	caves       map[string]*DataCave
	codeCaves   map[string]*CodeCave
}

func NewCaveManager(memory *ProcessMemory, baseAddress int64) *CaveManager {
	return &CaveManager{
		memory:      memory,
		baseAddress: baseAddress,
		caves:       make(map[string]*DataCave),
		codeCaves:   make(map[string]*CodeCave),
	}
}

func (cm *CaveManager) CreateDataCave(name string, pointerAddress int64, data []byte, pointerStyle PointerStyle) error {
	caveAddress, err := cm.memory.AllocateMemory(pointerAddress, len(data))
	if err != nil {
		return fmt.Errorf("failed to allocate memory: %v", err)
	}

	cave := &DataCave{
		name:           name,
		pointerAddress: pointerAddress,
		caveAddress:    caveAddress,
		data:           data,
		pointerStyle:   pointerStyle,
	}

	cm.caves[name] = cave
	return nil
}

func (cm *CaveManager) ActivateDataCave(name string) error {
	cave, exists := cm.caves[name]
	if !exists {
		return fmt.Errorf("cave %s not found", name)
	}

	if err := cm.memory.WriteMemory(cave.caveAddress, cave.data); err != nil {
		return fmt.Errorf("failed to write cave data: %v", err)
	}

	switch cave.pointerStyle {
	case DWordRelative:
		offset := int32(cave.caveAddress - (cave.pointerAddress + 4))
		pointer := make([]byte, 4)
		binary.LittleEndian.PutUint32(pointer, uint32(offset))

		if err := cm.memory.WriteMemory(cave.pointerAddress, pointer); err != nil {
			return fmt.Errorf("failed to write pointer: %v", err)
		}

	default:
		return fmt.Errorf("unsupported pointer style")
	}

	return nil
}

func (cm *CaveManager) DeactivateDataCave(name string) error {
	cave, exists := cm.caves[name]
	if !exists {
		return fmt.Errorf("cave %s not found", name)
	}

	switch cave.pointerStyle {
	case DWordRelative:
		pointer := make([]byte, 4)
		if err := cm.memory.WriteMemory(cave.pointerAddress, pointer); err != nil {
			return fmt.Errorf("failed to clear pointer: %v", err)
		}

	default:
		return fmt.Errorf("unsupported pointer style")
	}

	return nil
}

// CreateCodeCave creates a code cave for assembly injection
// injectionAddress: where to place the JMP to cave
// overwriteLength: how many bytes to overwrite (must be >= 5 for JMP)
// shellcode: assembly code to execute in the cave
func (cm *CaveManager) CreateCodeCave(name string, injectionAddress int64, overwriteLength int, shellcode []byte) error {
	if overwriteLength < 5 {
		return fmt.Errorf("overwrite length must be at least 5 bytes for JMP instruction")
	}

	// Read original bytes that will be overwritten
	originalBytes, err := cm.memory.ReadMemory(injectionAddress, overwriteLength)
	if err != nil {
		return fmt.Errorf("failed to read original bytes: %v", err)
	}

	// Calculate total cave size:
	// shellcode + original instructions + JMP back (5 bytes)
	caveSize := len(shellcode) + overwriteLength + 5

	// Allocate executable memory near the injection point
	caveAddress, err := cm.memory.AllocateMemory(injectionAddress, caveSize)
	if err != nil {
		return fmt.Errorf("failed to allocate code cave: %v", err)
	}

	// Build cave contents:
	// 1. Shellcode
	// 2. Original overwritten instructions
	// 3. JMP back to (injectionAddress + overwriteLength)
	caveData := make([]byte, 0, caveSize)
	caveData = append(caveData, shellcode...)
	caveData = append(caveData, originalBytes...)

	// Generate JMP back: E9 [offset]
	returnAddress := injectionAddress + int64(overwriteLength)
	jmpBackOffset := int32(returnAddress - (caveAddress + int64(len(shellcode)) + int64(overwriteLength) + 5))
	caveData = append(caveData, 0xE9) // JMP opcode
	jmpBackBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(jmpBackBytes, uint32(jmpBackOffset))
	caveData = append(caveData, jmpBackBytes...)

	cave := &CodeCave{
		name:             name,
		injectionAddress: injectionAddress,
		caveAddress:      caveAddress,
		shellcode:        caveData,
		overwriteLength:  overwriteLength,
		originalBytes:    originalBytes,
		active:           false,
	}

	cm.codeCaves[name] = cave
	return nil
}

// ActivateCodeCave writes the shellcode to the cave and redirects execution
func (cm *CaveManager) ActivateCodeCave(name string) error {
	cave, exists := cm.codeCaves[name]
	if !exists {
		return fmt.Errorf("code cave %s not found", name)
	}

	if cave.active {
		return nil // Already active
	}

	// Write shellcode to cave
	if err := cm.memory.WriteMemory(cave.caveAddress, cave.shellcode); err != nil {
		return fmt.Errorf("failed to write shellcode: %v", err)
	}

	// Generate JMP to cave at injection point: E9 [offset]
	jmpToCaveOffset := int32(cave.caveAddress - (cave.injectionAddress + 5))
	jmpInstruction := make([]byte, cave.overwriteLength)
	jmpInstruction[0] = 0xE9 // JMP opcode
	binary.LittleEndian.PutUint32(jmpInstruction[1:5], uint32(jmpToCaveOffset))
	// Fill remaining bytes with NOP (0x90) if overwrite length > 5
	for i := 5; i < cave.overwriteLength; i++ {
		jmpInstruction[i] = 0x90
	}

	if err := cm.memory.WriteMemory(cave.injectionAddress, jmpInstruction); err != nil {
		return fmt.Errorf("failed to write JMP to cave: %v", err)
	}

	cave.active = true
	return nil
}

// DeactivateCodeCave restores original bytes at injection point
func (cm *CaveManager) DeactivateCodeCave(name string) error {
	cave, exists := cm.codeCaves[name]
	if !exists {
		return fmt.Errorf("code cave %s not found", name)
	}

	if !cave.active {
		return nil // Already deactivated
	}

	// Restore original bytes
	if err := cm.memory.WriteMemory(cave.injectionAddress, cave.originalBytes); err != nil {
		return fmt.Errorf("failed to restore original bytes: %v", err)
	}

	cave.active = false
	return nil
}

// CodeCaveExists checks if a code cave with the given name exists
func (cm *CaveManager) CodeCaveExists(name string) bool {
	_, exists := cm.codeCaves[name]
	return exists
}
