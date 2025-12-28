package game

import "math"

const (
	ProcessName = "sekiro"

	PatternFramelockFuzzy          = "C7 43 ?? ?? ?? ?? ?? 4C 89 AB"
	PatternFramelockFuzzyOffset    = 3
	PatternFramelockSpeedFix       = "F3 0F 58 ?? 0F C6 ?? 00 0F 51 ?? F3 0F 59 ?? ?? ?? ?? ?? 0F 2F"
	PatternFramelockSpeedFixOffset = 15

	PatternResolutionDefault    = "80 07 00 00 38 04 00 00 00 08 00 00 80 04 00 00"
	PatternResolutionDefault720 = "00 05 00 00 D0 02 00 00 A0 05 00 00 2A 03 00 00"
	PatternResolutionScalingFix = "85 C9 74 ?? 47 8B ?? ?? ?? ?? ?? ?? 45 ?? ?? 74"

	PatternFovSetting       = "F3 0F 10 08 F3 0F 59 0D ?? ?? ?? ?? F3 0F 5C 4E"
	PatternFovSettingOffset = 8

	DefaultFOVDegrees = 1.0
	DegreesToRadians  = 0.0174533

	PatternCameraResetLockOn       = "C6 86 ?? ?? 00 00 ?? F3 0F 10 8E ?? ?? 00 00"
	PatternCameraResetLockOnOffset = 6

	PatternCameraAdjustPitch          = "0F 29 ?? ?? ?? 00 00 0F 29 ?? ?? ?? 00 00 0F 29 ?? ?? ?? 00 00 EB ?? F3"
	PatternCameraAdjustPitchOverwrite = 7

	PatternCameraAdjustYawZ          = "E8 ?? ?? ?? ?? F3 ?? ?? ?? ?? ?? 00 00 80 ?? ?? ?? 00 00 00 0F 84"
	PatternCameraAdjustYawZOffset    = 5
	PatternCameraAdjustYawZOverwrite = 8

	PatternCameraAdjustPitchXY          = "F3 ?? ?? ?? F3 ?? ?? ?? 70 01 00 00 F3 ?? ?? ?? ?? ?? ?? ?? E8 ?? ?? ?? ?? 0F"
	PatternCameraAdjustPitchXYOverwrite = 12

	PatternCameraAdjustYawXY          = "E8 ?? ?? ?? ?? F3 0F 11 86 ?? ?? 00 00 E9"
	PatternCameraAdjustYawXYOffset    = 5
	PatternCameraAdjustYawXYOverwrite = 8

	PatternAutoLoot       = "C6 85 ?? ?? ?? ?? ?? B0 01 EB ?? C6 85 ?? ?? ?? ?? ?? 32 C0"
	PatternAutoLootOffset = 18

	PatternDragonrotEffect       = "45 ?? ?? BA ?? ?? ?? ?? E8 ?? ?? ?? ?? 84 C0 0F 85 ?? ?? ?? ?? 48 8B 0D ?? ?? ?? ?? 48 85 C9 75 ?? 48 8D 0D ?? ?? ?? ?? E8 ?? ?? ?? ?? 4C ?? ?? 4C ?? ?? ?? ?? ?? ?? BA ?? ?? ?? ?? 48 8D 0D ?? ?? ?? ?? E8 ?? ?? ?? ?? 48 8B 0D ?? ?? ?? ?? 45 ?? ?? BA ?? ?? ?? ?? E8 ?? ?? ?? ?? 84 C0 0F 84 ?? ?? ?? ?? 48 8D"
	PatternDragonrotEffectOffset = 13

	PatternDeathPenalties1       = "F3 ?? 0F 2C ?? 41 ?? ?? 48 ?? ?? E8 ?? ?? ?? ?? 8B"
	PatternDeathPenalties1Offset = 11

	PatternDeathPenalties2       = "E8 ?? ?? ?? ?? 45 ?? ?? 44 89 ?? 24 ?? ?? 00 00 8B ?? 24 ?? ?? 00 00 2B ?? 89 ?? 24 ?? ?? 00 00 E8 ?? ?? ?? ?? 48 ?? ?? 24 ?? ?? 00 00 48 ?? ?? 48"
	PatternDeathPenalties2Offset = 0
	PatternDeathPenalties3Offset = 45

	PatternDeathPenalties2Legacy       = "8B ?? 89 83 ?? ?? ?? ?? 45 ?? ?? 44 89 ?? 24 ?? ?? 00 00 2B ?? 89 ?? 24 ?? ?? 00 00 E8"
	PatternDeathPenalties2OffsetLegacy = 2

	PatternGameSpeed                    = "48 8B 05 ?? ?? ?? ?? F3 0F 10 88 ?? ?? ?? ?? F3 0F"
	PatternGameSpeedInstructionLength   = 7
	PatternGameSpeedPointerOffsetOffset = 11

	PatternPlayerSpeed                  = "48 8B 1D ?? ?? ?? ?? 48 85 DB 74 ?? 8B ?? 81 FA"
	PatternPlayerSpeedInstructionLength = 7
	PatternPlayerSpeedPointer2Offset    = 0x0088
	PatternPlayerSpeedPointer3Offset    = 0x1FF8
	PatternPlayerSpeedPointer4Offset    = 0x0028
	PatternPlayerSpeedPointer5Offset    = 0x0D00

	PatternPlayerDeaths                    = "0F B6 48 ?? 88 8B ?? ?? 00 00 48 8B 05 ?? ?? ?? ?? 8B 88 ?? ?? 00 00 89 8B ?? ?? 00 00 48 8B 05 ?? ?? ?? ?? 8B 88 ?? ?? 00 00"
	PatternPlayerDeathsOffset              = 29
	PatternPlayerDeathsInstructionLength   = 7
	PatternPlayerDeathsPointerOffsetOffset = 9

	PatternTotalKills                  = "48 ?? D8 ?? ?? ?? ?? 48 8B 05 ?? ?? ?? ?? 48 ?? ?? 48 89 ?? ?? ?? 48 8B ?? 08"
	PatternTotalKillsOffset            = 7
	PatternTotalKillsInstructionLength = 7
	PatternTotalKillsPointer1Offset    = 0x0008
	PatternTotalKillsPointer2Offset    = 0x00DC
)

var (
	PatchResolutionScalingFixEnable  = []byte{0x90, 0x90, 0xEB}
	PatchResolutionScalingFixDisable = []byte{0x85, 0xC9, 0x74}

	PatchAutoLootEnable  = []byte{0xB0, 0x01} // mov al,1
	PatchAutoLootDisable = []byte{0x32, 0xC0} // xor al,al

	PatchDragonrotEffectDisable = []byte{0x90, 0x90, 0x90, 0xE9} // nop; nop; nop; jmp
	PatchDragonrotEffectEnable  = []byte{0x84, 0xC0, 0x0F, 0x85} // test al,al; jne

	PatchDeathPenalties1Disable = []byte{0x90, 0x90, 0x90, 0x90, 0x90} // nop (5 bytes)

	PatchDeathPenalties2Disable = []byte{ // nop (32 bytes)
		0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90,
		0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90,
		0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90,
		0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90,
	}

	PatchDeathPenalties3Disable = []byte{0x90, 0x90, 0x90} // nop (3 bytes)

	PatchDeathPenalties2DisableLegacy = []byte{ // nop (26 bytes)
		0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90,
		0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90,
		0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90,
		0x90, 0x90,
	}

	// Camera adjust shellcode
	ShellcodeCameraAdjustPitch = []byte{
		0x0F, 0x28, 0xA6, 0x70, 0x01, 0x00, 0x00, // movaps xmm4,xmmword ptr ds:[rsi+170]
		0x0F, 0x29, 0xA5, 0x70, 0x08, 0x00, 0x00, // movaps xmmword ptr ss:[rbp+870],xmm4
	}

	ShellcodeCameraAdjustYawZ = []byte{
		0xF3, 0x0F, 0x10, 0x86, 0x74, 0x01, 0x00, 0x00, // movss xmm0,dword ptr ds:[rsi+174]
		0xF3, 0x0F, 0x11, 0x86, 0x74, 0x01, 0x00, 0x00, // movss dword ptr ds:[rsi+174],xmm0
	}

	ShellcodeCameraAdjustPitchXY = []byte{
		0xF3, 0x0F, 0x10, 0x86, 0x70, 0x01, 0x00, 0x00, // movss xmm0,dword ptr ds:[rsi+170]
		0xF3, 0x0F, 0x11, 0x00, // movss dword ptr ds:[rax],xmm0
		0xF3, 0x0F, 0x10, 0x00, // movss xmm0,dword ptr ds:[rax]
		0xF3, 0x0F, 0x11, 0x86, 0x70, 0x01, 0x00, 0x00, // movss dword ptr ds:[rsi+170],xmm0
	}

	ShellcodeCameraAdjustYawXY = []byte{
		0xF3, 0x0F, 0x10, 0x86, 0x74, 0x01, 0x00, 0x00, // movss xmm0,dword ptr ds:[rsi+174]
		0xF3, 0x0F, 0x11, 0x86, 0x74, 0x01, 0x00, 0x00, // movss dword ptr ds:[rsi+174],xmm0
	}
)

var SpeedFixMatrix = []float32{
	15, 16, 16.6667, 18, 18.6875, 18.8516, 20, 24, 25, 27.5,
	30, 32, 38.5, 40, 48, 49.5, 50, 57.2958, 60, 64,
	66.75, 67, 78.8438, 80, 84, 90, 93.8, 100, 120, 127,
	128, 130, 140, 144, 150,
}

const SpeedFixDefaultValue = 30.0

func FindSpeedFixForFrameRate(frameRate int) float32 {
	idealSpeedFix := float32(frameRate) / 2.0
	closestSpeedFix := float32(SpeedFixDefaultValue)

	for _, speedFix := range SpeedFixMatrix {
		if math.Abs(float64(idealSpeedFix-speedFix)) < math.Abs(float64(idealSpeedFix-closestSpeedFix)) {
			closestSpeedFix = speedFix
		}
	}

	return closestSpeedFix
}
