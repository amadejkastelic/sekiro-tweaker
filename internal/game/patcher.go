package game

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/amadejkastelic/sekiro-tweaker/internal/memory"
)

type Patcher struct {
	mem         *memory.ProcessMemory
	scanner     *memory.PatternScanner
	peParser    *memory.PEParser
	caveManager *memory.CaveManager
	baseAddress int64
}

func NewPatcher(pid int) (*Patcher, error) {
	mem := memory.NewProcessMemory(pid)

	baseAddress, err := mem.GetModuleBaseAddress(ProcessName)
	if err != nil {
		return nil, fmt.Errorf("failed to find module: %v", err)
	}

	scanner := memory.NewPatternScanner(mem, ProcessName)
	peParser := memory.NewPEParser(mem, baseAddress)
	caveManager := memory.NewCaveManager(mem, baseAddress)

	return &Patcher{
		mem:         mem,
		scanner:     scanner,
		peParser:    peParser,
		caveManager: caveManager,
		baseAddress: baseAddress,
	}, nil
}

func (p *Patcher) ApplyFPSPatch(targetFPS int) error {
	address, err := p.scanner.FindPattern(PatternFramelockFuzzy)
	if err != nil {
		return fmt.Errorf("failed to find FPS pattern: %v", err)
	}

	targetAddress := address + PatternFramelockFuzzyOffset
	fpsValue := 1.0 / float32(targetFPS)

	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, math.Float32bits(fpsValue))

	if err := p.mem.WriteMemory(targetAddress, data); err != nil {
		return fmt.Errorf("failed to write FPS value: %v", err)
	}

	speedFixAddress, err := p.scanner.FindPattern(PatternFramelockSpeedFix)
	if err != nil {
		return nil
	}

	speedFixValue := FindSpeedFixForFrameRate(targetFPS)
	speedFixPointer := speedFixAddress + PatternFramelockSpeedFixOffset

	speedFixData := make([]byte, 4)
	binary.LittleEndian.PutUint32(speedFixData, math.Float32bits(speedFixValue))

	if err := p.caveManager.CreateDataCave("speedfix", speedFixPointer, speedFixData, memory.DWordRelative); err != nil {
		return fmt.Errorf("failed to create speed fix cave: %v", err)
	}

	if err := p.caveManager.ActivateDataCave("speedfix"); err != nil {
		return fmt.Errorf("failed to activate speed fix cave: %v", err)
	}

	return nil
}

func (p *Patcher) ApplyResolutionPatch(width, height int) error {
	dataAddress, dataSize, err := p.peParser.FindSection(".data")
	if err != nil {
		return fmt.Errorf("failed to find .data section: %v", err)
	}

	address, err := p.scanner.FindPatternInRegion(PatternResolutionDefault, dataAddress, dataSize)
	if err != nil {
		address, err = p.scanner.FindPatternInRegion(PatternResolutionDefault720, dataAddress, dataSize)
		if err != nil {
			return fmt.Errorf("failed to find resolution pattern: %v", err)
		}
	}

	data := make([]byte, 8)
	binary.LittleEndian.PutUint32(data[0:], uint32(width))
	binary.LittleEndian.PutUint32(data[4:], uint32(height))

	if err := p.mem.WriteMemory(address, data); err != nil {
		return fmt.Errorf("failed to write resolution: %v", err)
	}

	widescreenAddress, err := p.scanner.FindPattern(PatternResolutionScalingFix)
	if err == nil {
		p.mem.WriteMemory(widescreenAddress, PatchResolutionScalingFixEnable)
	}

	return nil
}

func (p *Patcher) ApplyFOVPatch(fovDegrees float32) error {
	address, err := p.scanner.FindPattern(PatternFovSetting)
	if err != nil {
		return fmt.Errorf("failed to find FOV pattern: %v", err)
	}

	fovPointer := address + PatternFovSettingOffset

	fovRadians := fovDegrees * DegreesToRadians
	fovData := make([]byte, 4)
	binary.LittleEndian.PutUint32(fovData, math.Float32bits(fovRadians))

	if err := p.caveManager.CreateDataCave("fov", fovPointer, fovData, memory.DWordRelative); err != nil {
		return fmt.Errorf("failed to create FOV cave: %v", err)
	}

	if err := p.caveManager.ActivateDataCave("fov"); err != nil {
		return fmt.Errorf("failed to activate FOV cave: %v", err)
	}

	return nil
}

func (p *Patcher) RemoveFPSPatch() error {
	if err := p.caveManager.DeactivateDataCave("speedfix"); err != nil {
		return err
	}

	address, err := p.scanner.FindPattern(PatternFramelockFuzzy)
	if err != nil {
		return err
	}

	targetAddress := address + PatternFramelockFuzzyOffset
	fpsValue := float32(1.0 / 60.0)

	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, math.Float32bits(fpsValue))

	return p.mem.WriteMemory(targetAddress, data)
}

func (p *Patcher) RemoveFOVPatch() error {
	return p.caveManager.DeactivateDataCave("fov")
}

func (p *Patcher) ApplyCameraResetPatch(disable bool) error {
	address, err := p.scanner.FindPattern(PatternCameraResetLockOn)
	if err != nil {
		return fmt.Errorf("failed to find camera reset pattern: %v", err)
	}

	targetAddress := address + PatternCameraResetLockOnOffset

	var value byte
	if disable {
		value = 0x00
	} else {
		value = 0x01
	}

	return p.mem.WriteMemory(targetAddress, []byte{value})
}

func (p *Patcher) ApplyAutoLootPatch(enable bool) error {
	address, err := p.scanner.FindPattern(PatternAutoLoot)
	if err != nil {
		return fmt.Errorf("failed to find auto-loot pattern: %v", err)
	}

	targetAddress := address + PatternAutoLootOffset

	var patch []byte
	if enable {
		patch = PatchAutoLootEnable
	} else {
		patch = PatchAutoLootDisable
	}

	return p.mem.WriteMemory(targetAddress, patch)
}

func (p *Patcher) ApplyDragonrotPatch(disable bool) error {
	address, err := p.scanner.FindPattern(PatternDragonrotEffect)
	if err != nil {
		return fmt.Errorf("failed to find dragonrot pattern: %v", err)
	}

	targetAddress := address + PatternDragonrotEffectOffset

	var patch []byte
	if disable {
		patch = PatchDragonrotEffectDisable
	} else {
		patch = PatchDragonrotEffectEnable
	}

	return p.mem.WriteMemory(targetAddress, patch)
}

func (p *Patcher) ApplyDeathPenaltyPatch(disable bool) error {
	if !disable {
		// Enabling requires restoring original bytes
		// Game restart required to restore
		return nil
	}

	// Patch 1: Disable Sen loss function call (5 bytes)
	address1, err := p.scanner.FindPattern(PatternDeathPenalties1)
	if err != nil {
		return fmt.Errorf("failed to find death penalty pattern 1: %v", err)
	}
	targetAddress1 := address1 + PatternDeathPenalties1Offset

	if err := p.mem.WriteMemory(targetAddress1, PatchDeathPenalties1Disable); err != nil {
		return fmt.Errorf("failed to apply death penalty patch 1: %v", err)
	}

	// Patch 2: Try modern pattern first, then legacy
	address2, err := p.scanner.FindPattern(PatternDeathPenalties2)
	isLegacy := false
	if err != nil {
		// Try legacy pattern
		address2, err = p.scanner.FindPattern(PatternDeathPenalties2Legacy)
		if err != nil {
			// Pattern 2 not found - not critical, pattern 1 already applied
			return nil
		}
		isLegacy = true
	}

	if isLegacy {
		targetAddress2 := address2 + PatternDeathPenalties2OffsetLegacy
		if err := p.mem.WriteMemory(targetAddress2, PatchDeathPenalties2DisableLegacy); err != nil {
			return fmt.Errorf("failed to apply death penalty patch 2 (legacy): %v", err)
		}
	} else {
		targetAddress2 := address2 + PatternDeathPenalties2Offset
		if err := p.mem.WriteMemory(targetAddress2, PatchDeathPenalties2Disable); err != nil {
			return fmt.Errorf("failed to apply death penalty patch 2: %v", err)
		}

		// Patch 3: Only exists in modern version (offset from pattern 2)
		targetAddress3 := address2 + PatternDeathPenalties3Offset
		if err := p.mem.WriteMemory(targetAddress3, PatchDeathPenalties3Disable); err != nil {
			return fmt.Errorf("failed to apply death penalty patch 3: %v", err)
		}
	}

	return nil
}

func (p *Patcher) GetGameSpeedAddress() (int64, error) {
	refAddress, err := p.scanner.FindPattern(PatternGameSpeed)
	if err != nil {
		return 0, fmt.Errorf("failed to find game speed pattern: %v", err)
	}

	// Dereference the static pointer
	timescaleManager, err := p.mem.DereferenceStaticPointer(refAddress, PatternGameSpeedInstructionLength)
	if err != nil {
		return 0, fmt.Errorf("failed to dereference timescale manager: %v", err)
	}

	// Read the offset to the actual timescale value
	offset, err := p.mem.ReadInt32(refAddress + PatternGameSpeedPointerOffsetOffset)
	if err != nil {
		return 0, fmt.Errorf("failed to read timescale offset: %v", err)
	}

	return timescaleManager + int64(offset), nil
}

func (p *Patcher) SetGameSpeed(speed float32) error {
	address, err := p.GetGameSpeedAddress()
	if err != nil {
		return err
	}

	return p.mem.WriteFloat32(address, speed)
}

func (p *Patcher) GetGameSpeed() (float32, error) {
	address, err := p.GetGameSpeedAddress()
	if err != nil {
		return 0, err
	}

	return p.mem.ReadFloat32(address)
}

func (p *Patcher) GetPlayerSpeedAddress() (int64, error) {
	// Find the pattern for the first pointer
	lpPlayerStructRelated1, err := p.scanner.FindPattern(PatternPlayerSpeed)
	if err != nil {
		return 0, fmt.Errorf("failed to find player speed pattern: %v", err)
	}

	// Dereference pointer 1 -> pointer 2
	lpPlayerStructRelated2, err := p.mem.DereferenceStaticPointer(lpPlayerStructRelated1, PatternPlayerSpeedInstructionLength)
	if err != nil {
		return 0, fmt.Errorf("failed to dereference player struct pointer 1: %v", err)
	}

	// Validate pointer 2 is not null
	if lpPlayerStructRelated2 == 0 {
		return 0, fmt.Errorf("player not loaded (pointer 2 is null)")
	}

	// Follow pointer chain: 2 -> 3
	lpPlayerStructRelated3, err := p.mem.ReadInt64(lpPlayerStructRelated2)
	if err != nil {
		return 0, fmt.Errorf("failed to read player struct pointer 2: %v", err)
	}
	if lpPlayerStructRelated3 == 0 {
		return 0, fmt.Errorf("player not loaded (pointer 3 is null)")
	}
	lpPlayerStructRelated3 += PatternPlayerSpeedPointer2Offset

	// 3 -> 4
	lpPlayerStructRelated4, err := p.mem.ReadInt64(lpPlayerStructRelated3)
	if err != nil {
		return 0, fmt.Errorf("failed to read player struct pointer 3: %v", err)
	}
	if lpPlayerStructRelated4 == 0 {
		return 0, fmt.Errorf("player not loaded (pointer 4 is null)")
	}
	lpPlayerStructRelated4 += PatternPlayerSpeedPointer3Offset

	// 4 -> 5
	lpPlayerStructRelated5, err := p.mem.ReadInt64(lpPlayerStructRelated4)
	if err != nil {
		return 0, fmt.Errorf("failed to read player struct pointer 4: %v", err)
	}
	if lpPlayerStructRelated5 == 0 {
		return 0, fmt.Errorf("player not loaded (pointer 5 is null)")
	}
	lpPlayerStructRelated5 += PatternPlayerSpeedPointer4Offset

	// 5 -> final address
	playerSpeedBase, err := p.mem.ReadInt64(lpPlayerStructRelated5)
	if err != nil {
		return 0, fmt.Errorf("failed to read player struct pointer 5: %v", err)
	}
	if playerSpeedBase == 0 {
		return 0, fmt.Errorf("player not loaded (final pointer is null)")
	}

	return playerSpeedBase + PatternPlayerSpeedPointer5Offset, nil
}

func (p *Patcher) SetPlayerSpeed(speed float32) error {
	address, err := p.GetPlayerSpeedAddress()
	if err != nil {
		return err
	}

	return p.mem.WriteFloat32(address, speed)
}

func (p *Patcher) GetPlayerSpeed() (float32, error) {
	address, err := p.GetPlayerSpeedAddress()
	if err != nil {
		return 0, err
	}

	return p.mem.ReadFloat32(address)
}

func (p *Patcher) GetPlayerDeathsAddress() (int64, error) {
	refAddress, err := p.scanner.FindPattern(PatternPlayerDeaths)
	if err != nil {
		return 0, fmt.Errorf("failed to find player deaths pattern: %v", err)
	}
	refAddress += PatternPlayerDeathsOffset

	// Dereference the static pointer
	lpPlayerStatsRelated, err := p.mem.DereferenceStaticPointer(refAddress, PatternPlayerDeathsInstructionLength)
	if err != nil {
		return 0, fmt.Errorf("failed to dereference player stats: %v", err)
	}

	// Read the offset to the death counter
	offset, err := p.mem.ReadInt32(refAddress + PatternPlayerDeathsPointerOffsetOffset)
	if err != nil {
		return 0, fmt.Errorf("failed to read deaths offset: %v", err)
	}

	return lpPlayerStatsRelated + int64(offset), nil
}

func (p *Patcher) GetPlayerDeaths() (int32, error) {
	address, err := p.GetPlayerDeathsAddress()
	if err != nil {
		return 0, err
	}

	return p.mem.ReadInt32(address)
}

func (p *Patcher) GetTotalKillsAddress() (int64, error) {
	refAddress, err := p.scanner.FindPattern(PatternTotalKills)
	if err != nil {
		return 0, fmt.Errorf("failed to find total kills pattern: %v", err)
	}
	refAddress += PatternTotalKillsOffset

	// Dereference pointer 1
	lpPlayerStatsRelatedKills1, err := p.mem.DereferenceStaticPointer(refAddress, PatternTotalKillsInstructionLength)
	if err != nil {
		return 0, fmt.Errorf("failed to dereference kills pointer 1: %v", err)
	}

	// Follow pointer chain: pointer 1 -> 2
	lpPlayerStructRelatedKills2, err := p.mem.ReadInt64(lpPlayerStatsRelatedKills1)
	if err != nil {
		return 0, fmt.Errorf("failed to read kills pointer 1: %v", err)
	}
	lpPlayerStructRelatedKills2 += PatternTotalKillsPointer1Offset

	// 2 -> final address
	totalKillsBase, err := p.mem.ReadInt64(lpPlayerStructRelatedKills2)
	if err != nil {
		return 0, fmt.Errorf("failed to read kills pointer 2: %v", err)
	}

	return totalKillsBase + PatternTotalKillsPointer2Offset, nil
}

func (p *Patcher) GetTotalKills() (int32, error) {
	address, err := p.GetTotalKillsAddress()
	if err != nil {
		return 0, err
	}

	return p.mem.ReadInt32(address)
}

func (p *Patcher) ApplyCameraAutoRotatePatch(disable bool) error {
	if !disable {
		// Restore original behavior by deactivating code caves
		errors := []error{}
		if err := p.caveManager.DeactivateCodeCave("camera_pitch"); err != nil {
			errors = append(errors, fmt.Errorf("pitch: %v", err))
		}
		if err := p.caveManager.DeactivateCodeCave("camera_yaw_z"); err != nil {
			errors = append(errors, fmt.Errorf("yaw_z: %v", err))
		}
		if err := p.caveManager.DeactivateCodeCave("camera_pitch_xy"); err != nil {
			errors = append(errors, fmt.Errorf("pitch_xy: %v", err))
		}
		if err := p.caveManager.DeactivateCodeCave("camera_yaw_xy"); err != nil {
			errors = append(errors, fmt.Errorf("yaw_xy: %v", err))
		}

		if len(errors) > 0 {
			return fmt.Errorf("failed to deactivate camera caves: %v", errors)
		}
		return nil
	}

	// Create all 4 code caves first (without activating)
	// This validates patterns and allocates memory without modifying game code
	errors := []error{}
	successCount := 0

	// 1. Camera Pitch
	pitchAddr, err := p.scanner.FindPattern(PatternCameraAdjustPitch)
	if err == nil {
		if err := p.caveManager.CreateCodeCave("camera_pitch", pitchAddr, PatternCameraAdjustPitchOverwrite, ShellcodeCameraAdjustPitch); err != nil {
			errors = append(errors, fmt.Errorf("create camera_pitch: %v", err))
		} else {
			successCount++
		}
	} else {
		errors = append(errors, fmt.Errorf("find camera_pitch pattern: %v", err))
	}

	// 2. Camera Yaw Z
	yawZAddr, err := p.scanner.FindPattern(PatternCameraAdjustYawZ)
	if err == nil {
		yawZAddr += PatternCameraAdjustYawZOffset
		if err := p.caveManager.CreateCodeCave("camera_yaw_z", yawZAddr, PatternCameraAdjustYawZOverwrite, ShellcodeCameraAdjustYawZ); err != nil {
			errors = append(errors, fmt.Errorf("create camera_yaw_z: %v", err))
		} else {
			successCount++
		}
	} else {
		errors = append(errors, fmt.Errorf("find camera_yaw_z pattern: %v", err))
	}

	// 3. Camera Pitch XY
	pitchXYAddr, err := p.scanner.FindPattern(PatternCameraAdjustPitchXY)
	if err == nil {
		if err := p.caveManager.CreateCodeCave("camera_pitch_xy", pitchXYAddr, PatternCameraAdjustPitchXYOverwrite, ShellcodeCameraAdjustPitchXY); err != nil {
			errors = append(errors, fmt.Errorf("create camera_pitch_xy: %v", err))
		} else {
			successCount++
		}
	} else {
		errors = append(errors, fmt.Errorf("find camera_pitch_xy pattern: %v", err))
	}

	// 4. Camera Yaw XY
	yawXYAddr, err := p.scanner.FindPattern(PatternCameraAdjustYawXY)
	if err == nil {
		yawXYAddr += PatternCameraAdjustYawXYOffset
		if err := p.caveManager.CreateCodeCave("camera_yaw_xy", yawXYAddr, PatternCameraAdjustYawXYOverwrite, ShellcodeCameraAdjustYawXY); err != nil {
			errors = append(errors, fmt.Errorf("create camera_yaw_xy: %v", err))
		} else {
			successCount++
		}
	} else {
		errors = append(errors, fmt.Errorf("find camera_yaw_xy pattern: %v", err))
	}

	// If we couldn't create any caves, fail early
	if successCount == 0 {
		return fmt.Errorf("camera auto-rotate: failed to create any code caves: %v", errors)
	}

	// Now activate the caves that were successfully created
	// Do this carefully and check for process validity between activations
	activationErrors := []error{}

	if p.caveManager.CodeCaveExists("camera_pitch") {
		if err := p.caveManager.ActivateCodeCave("camera_pitch"); err != nil {
			activationErrors = append(activationErrors, fmt.Errorf("activate camera_pitch: %v", err))
		}
	}

	if p.caveManager.CodeCaveExists("camera_yaw_z") {
		if err := p.caveManager.ActivateCodeCave("camera_yaw_z"); err != nil {
			activationErrors = append(activationErrors, fmt.Errorf("activate camera_yaw_z: %v", err))
		}
	}

	if p.caveManager.CodeCaveExists("camera_pitch_xy") {
		if err := p.caveManager.ActivateCodeCave("camera_pitch_xy"); err != nil {
			activationErrors = append(activationErrors, fmt.Errorf("activate camera_pitch_xy: %v", err))
		}
	}

	if p.caveManager.CodeCaveExists("camera_yaw_xy") {
		if err := p.caveManager.ActivateCodeCave("camera_yaw_xy"); err != nil {
			activationErrors = append(activationErrors, fmt.Errorf("activate camera_yaw_xy: %v", err))
		}
	}

	// Combine all errors for reporting
	allErrors := append(errors, activationErrors...)
	if len(allErrors) > 0 {
		return fmt.Errorf("camera auto-rotate partial success (%d/4 caves): %v", successCount, allErrors)
	}

	return nil
}
