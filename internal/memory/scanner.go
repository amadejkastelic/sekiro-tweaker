package memory

import (
	"encoding/hex"
	"fmt"
	"strings"
)

type PatternScanner struct {
	memory     *ProcessMemory
	moduleName string
}

func NewPatternScanner(memory *ProcessMemory, moduleName string) *PatternScanner {
	return &PatternScanner{
		memory:     memory,
		moduleName: moduleName,
	}
}

func (ps *PatternScanner) FindPattern(pattern string) (int64, error) {
	baseAddress, err := ps.memory.GetModuleBaseAddress(ps.moduleName)
	if err != nil {
		return -1, err
	}

	moduleSize, err := ps.memory.GetModuleSize(ps.moduleName)
	if err != nil {
		return -1, err
	}

	moduleData, err := ps.memory.ReadMemory(baseAddress, moduleSize)
	if err != nil {
		return -1, err
	}

	return ps.findPatternInData(pattern, moduleData, baseAddress)
}

func (ps *PatternScanner) FindPatternInRegion(pattern string, address int64, size int) (int64, error) {
	data, err := ps.memory.ReadMemory(address, size)
	if err != nil {
		return -1, err
	}

	return ps.findPatternInData(pattern, data, address)
}

func (ps *PatternScanner) findPatternInData(pattern string, data []byte, baseAddress int64) (int64, error) {
	parts := strings.Split(pattern, " ")
	var bytes []byte
	var mask []bool

	for _, part := range parts {
		if part == "??" {
			bytes = append(bytes, 0)
			mask = append(mask, false)
		} else {
			b, err := hex.DecodeString(part)
			if err != nil {
				return -1, fmt.Errorf("invalid pattern: %v", err)
			}
			bytes = append(bytes, b[0])
			mask = append(mask, true)
		}
	}

	if len(bytes) == 0 {
		return -1, fmt.Errorf("empty pattern")
	}

	var matchIndices []int
	var matchBytes []byte

	for i := len(bytes) - 1; i >= 0; i-- {
		if mask[i] {
			matchIndices = append(matchIndices, i)
			matchBytes = append(matchBytes, bytes[i])
		}
	}

	if len(matchBytes) == 0 {
		return -1, fmt.Errorf("pattern contains only wildcards")
	}

	dataLength := len(data) - len(bytes)

	for offset := 0; offset < dataLength; offset++ {
		if data[offset] != matchBytes[len(matchBytes)-1] {
			continue
		}

		found := true
		for i := 0; i < len(matchIndices); i++ {
			index := matchIndices[i]
			if data[offset+index] != matchBytes[i] {
				found = false
				break
			}
		}

		if found {
			return baseAddress + int64(offset), nil
		}
	}

	return -1, fmt.Errorf("pattern not found")
}
