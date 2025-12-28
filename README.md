# Sekiro Tweaker

A Linux-native game tweaker for Sekiro: Shadows Die Twice running under Proton/Wine. Written in Go with GTK4.

## Features

### Graphics Settings
- **FPS Unlock**: Remove the 60 FPS cap (30-300 FPS)
- **Custom Resolution**: Set any resolution you want
- **FOV Adjustment**: Adjust field of view in degrees (default: 1.0°, range: 0.5-2.5°)

### Camera Settings
- **Disable Camera Reset on Lock-on**: Prevents annoying camera centering when locking on with no target

### Gameplay Modifications
- **Auto-loot**: Automatically pickup and collect all enemy loot/items
- **Prevent Dragonrot**: NPCs won't get the disease on death
- **Disable Death Penalties**: No Sen or Experience loss on death

### Speed Modifiers
- **Game Speed**: Adjust overall game speed (0.1x - 5.0x, default 1.0x) ✅ Works reliably
- **Player Speed**: Adjust player movement speed (0.1x - 5.0x, default 1.0x) ⚠️ Experimental
  - **Known Issue**: May not work reliably on Linux/Proton due to deep pointer chain
  - Requires being fully loaded into game world (not at main menu)
  - Pointer addresses may change on save load/fast travel

### Stats Display
- **Death Counter**: Real-time display of player deaths
- **Kill Counter**: Real-time display of total enemy kills

### Technical Features
- **Automatic Game Detection**: Finds and attaches to Sekiro automatically
- **Memory-Safe Patching**: Uses data caves for pointer redirection
- **Real-time Stats**: Continuous monitoring of player stats (deaths/kills)
- **Configuration Persistence**: Settings are automatically saved and restored between sessions

## Building

```bash
nix build .#
```

## Running

```bash
nix run .#
```

The application will automatically detect when Sekiro is running and enable the patch controls.

## Configuration

Settings are automatically saved when you click "Apply Patches" and restored on next launch.

Configuration file location: `~/.config/sekiro-tweaker/config.yaml`

## How It Works

### Memory Access
- Uses `process_vm_readv/writev` syscalls for fast memory access
- Falls back to `/proc/pid/mem` for writing to read-only code sections
- Parses `/proc/pid/maps` to find module base addresses and sections

### Pattern Scanning
- Scans game memory for byte patterns with wildcard support
- Finds PE sections (.data, .text) for targeted scanning
- Caches scan results for performance

### Memory Caves
- Allocates memory in writable data regions near target code
- Uses relative pointer redirection (DWordRelative style)
- Tracks allocations to prevent collisions

### Patches

**FPS Unlock**:
- Writes new frame time value (1/FPS) to game's frame limiter
- Creates data cave for speed fix lookup table
- Automatically selects correct speed fix value from matrix

**Resolution**:
- Scans .data section for resolution patterns
- Writes custom width/height values
- Patches aspect ratio scaling for widescreen support

**FOV**:
- Creates data cave with FOV value in radians (degrees × 0.0174533)
- Default game FOV is 1.0 degrees
- Redirects FOV calculation pointer to cave
- Values: 0.5° = narrower, 1.0° = default, 2.0° = wider

**Camera Reset**:
- Simple byte patch to disable camera centering on lock-on
- Writes 0x00 to disable, 0x01 to enable

**Auto-loot**:
- 2-byte patch that changes `xor al,al` to `mov al,1`
- Forces loot pickup flag to always be true

**Dragonrot Prevention**:
- 4-byte patch that converts conditional jump to unconditional
- Changes `test al,al; jne` to `nop; nop; nop; jmp`
- Skips dragonrot increase routine on death

**Death Penalties**:
- **Pattern 1** (5 bytes): NOPs function call that reduces Sen
- **Pattern 2** (32 bytes modern / 26 bytes legacy): NOPs AP/Sen virtual decrease display
- **Pattern 3** (3 bytes modern only): NOPs additional penalty routine
- Automatically detects game version (modern vs legacy)
- Falls back gracefully if patterns 2/3 not found
- Only supports disabling (game restart required to restore)

**Game Speed**:
- Reads RIP-relative pointer to timescale manager
- Writes float value directly to game's timescale
- Default: 1.0 (normal speed), 0.5 = half speed, 2.0 = double speed

**Player Speed**:
- Follows 5-level pointer chain to player timescale value
- Pointer chain: base -> +0x88 -> +0x1FF8 -> +0x28 -> +0xD00
- Requires player to be loaded in game world (not available at main menu)
- Can change on save load/fast travel (may need reapplication)

**Death/Kill Counters**:
- Reads from player stats struct via pointer dereferencing
- Deaths: Static pointer + dynamic offset read from instruction
- Kills: Two-level pointer chain + fixed offsets (0x08, 0xDC)
- Updates every second in background thread

## Architecture

```
sekiro-tweaker/
├── cmd/
│   └── sekiro-tweaker/
│       └── main.go       # GTK4 UI and application entry point
├── internal/
│   ├── memory/
│   │   ├── process.go    # Memory read/write, process detection
│   │   ├── scanner.go    # Pattern scanning
│   │   ├── pe.go         # PE header parsing
│   │   ├── cave.go       # Data cave management
│   │   └── syscalls_linux.go  # Linux syscall definitions
│   └── game/
│       ├── constants.go  # Game patterns and speed fix matrix
│       └── patcher.go    # Patch implementations
└── go.mod
```
