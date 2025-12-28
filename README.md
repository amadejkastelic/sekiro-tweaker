# Sekiro Tweaker

A Linux-native game tweaker for Sekiro: Shadows Die Twice running under Proton/Wine. Written in Go with GTK4.

<img width="628" height="678" alt="image" src="https://github.com/user-attachments/assets/9295f378-c8f9-4f40-ad47-8e8be691b0df" />

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


## Cachix

Set up cachix, so you don't have to build the package each time:

```nix
{
  nix.settings = {
    substituters = ["https://amadejkastelic.cachix.org"];
    trusted-public-keys = ["amadejkastelic.cachix.org-1:EiQfTbiT0UKsynF4q3nbNYjNH6/l7zuhrNkQTuXmyOs="];
  };
}
```

## Running

```bash
nix run github:amadejkastelic/sekiro-tweaker --accept-flake-config"
```


## Building

```bash
nix build .#
```

The application will automatically detect when Sekiro is running and enable the patch controls.

## Configuration

Settings are automatically saved when you click "Apply Patches" and restored on next launch.

Configuration file location: `~/.config/sekiro-tweaker/config.yaml`
