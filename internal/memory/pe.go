package memory

import (
	"encoding/binary"
	"fmt"
	"strings"
)

type PEParser struct {
	memory      *ProcessMemory
	baseAddress int64
}

func NewPEParser(memory *ProcessMemory, baseAddress int64) *PEParser {
	return &PEParser{
		memory:      memory,
		baseAddress: baseAddress,
	}
}

func (pe *PEParser) FindSection(sectionName string) (int64, int, error) {
	dosHeader, err := pe.memory.ReadMemory(pe.baseAddress, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read DOS header: %v", err)
	}

	if dosHeader[0] != 0x4D || dosHeader[1] != 0x5A {
		return 0, 0, fmt.Errorf("invalid DOS signature")
	}

	peHeaderOffset := binary.LittleEndian.Uint32(dosHeader[0x3C:])

	coffHeader, err := pe.memory.ReadMemory(pe.baseAddress+int64(peHeaderOffset), 24)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read COFF header: %v", err)
	}

	if coffHeader[0] != 0x50 || coffHeader[1] != 0x45 || coffHeader[2] != 0x00 || coffHeader[3] != 0x00 {
		return 0, 0, fmt.Errorf("invalid PE signature")
	}

	numberOfSections := binary.LittleEndian.Uint16(coffHeader[6:])
	sizeOfOptionalHeader := binary.LittleEndian.Uint16(coffHeader[20:])

	sectionHeadersOffset := pe.baseAddress + int64(peHeaderOffset) + 4 + 20 + int64(sizeOfOptionalHeader)

	for i := uint16(0); i < numberOfSections; i++ {
		sectionHeader, err := pe.memory.ReadMemory(sectionHeadersOffset+int64(i)*40, 40)
		if err != nil {
			continue
		}

		name := string(sectionHeader[0:8])
		name = strings.TrimRight(name, "\x00")

		if strings.EqualFold(name, sectionName) {
			virtualSize := binary.LittleEndian.Uint32(sectionHeader[8:])
			virtualAddress := binary.LittleEndian.Uint32(sectionHeader[12:])

			sectionAddr := pe.baseAddress + int64(virtualAddress)
			return sectionAddr, int(virtualSize), nil
		}
	}

	return 0, 0, fmt.Errorf("section %s not found", sectionName)
}
