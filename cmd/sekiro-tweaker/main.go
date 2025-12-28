package main

import (
	"fmt"
	"os"
	"time"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"go.uber.org/zap"

	"github.com/amadejkastelic/sekiro-tweaker/internal/config"
	"github.com/amadejkastelic/sekiro-tweaker/internal/game"
	"github.com/amadejkastelic/sekiro-tweaker/internal/logger"
	"github.com/amadejkastelic/sekiro-tweaker/internal/memory"
)

const appID = "com.github.amadejkastelic.sekiro-tweaker"

type Application struct {
	app    *gtk.Application
	window *gtk.ApplicationWindow

	statusLabel *gtk.Label
	pidLabel    *gtk.Label
	applyButton *gtk.Button

	fpsCheck      *gtk.CheckButton
	fpsSpin       *gtk.SpinButton
	resCheck      *gtk.CheckButton
	widthSpin     *gtk.SpinButton
	heightSpin    *gtk.SpinButton
	fovCheck      *gtk.CheckButton
	fovSpin       *gtk.SpinButton
	camResetCheck *gtk.CheckButton

	autoLootCheck     *gtk.CheckButton
	dragonrotCheck    *gtk.CheckButton
	deathPenaltyCheck *gtk.CheckButton

	gameSpeedCheck   *gtk.CheckButton
	gameSpeedSpin    *gtk.SpinButton
	playerSpeedCheck *gtk.CheckButton
	playerSpeedSpin  *gtk.SpinButton

	deathsLabel *gtk.Label
	killsLabel  *gtk.Label

	errorsExpander *gtk.Expander
	errorsView     *gtk.TextView
	errorsBuffer   *gtk.TextBuffer

	patcher *game.Patcher
	gamePID int
}

func main() {
	app := gtk.NewApplication(appID, gio.ApplicationFlagsNone)
	appState := &Application{app: app}

	app.ConnectActivate(func() { appState.activate() })

	if code := app.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}

func (a *Application) activate() {
	a.window = gtk.NewApplicationWindow(a.app)
	a.window.SetTitle("Sekiro Tweaker")
	a.window.SetDefaultSize(400, 500)

	mainBox := gtk.NewBox(gtk.OrientationVertical, 10)
	mainBox.SetMarginTop(20)
	mainBox.SetMarginBottom(20)
	mainBox.SetMarginStart(20)
	mainBox.SetMarginEnd(20)

	a.statusLabel = gtk.NewLabel("Waiting for Sekiro...")
	a.statusLabel.AddCSSClass("title-2")
	mainBox.Append(a.statusLabel)

	a.pidLabel = gtk.NewLabel("PID: -")
	a.pidLabel.AddCSSClass("dim-label")
	mainBox.Append(a.pidLabel)

	separator1 := gtk.NewSeparator(gtk.OrientationHorizontal)
	separator1.SetMarginTop(10)
	separator1.SetMarginBottom(10)
	mainBox.Append(separator1)

	graphicsExpander := gtk.NewExpander("Graphics Settings")
	graphicsExpander.SetExpanded(true)
	graphicsBox := gtk.NewBox(gtk.OrientationVertical, 10)
	graphicsBox.SetMarginStart(15)
	graphicsBox.SetMarginTop(10)
	graphicsBox.SetMarginBottom(10)

	fpsBox := gtk.NewBox(gtk.OrientationHorizontal, 10)
	a.fpsCheck = gtk.NewCheckButtonWithLabel("FPS Unlock")
	a.fpsCheck.SetActive(true)
	fpsBox.Append(a.fpsCheck)
	a.fpsSpin = gtk.NewSpinButtonWithRange(30, 300, 1)
	a.fpsSpin.SetValue(144)
	fpsBox.Append(a.fpsSpin)
	graphicsBox.Append(fpsBox)

	resBox := gtk.NewBox(gtk.OrientationVertical, 5)
	a.resCheck = gtk.NewCheckButtonWithLabel("Custom Resolution")
	a.resCheck.SetActive(false)
	resBox.Append(a.resCheck)

	resInputBox := gtk.NewBox(gtk.OrientationHorizontal, 10)
	resInputBox.SetMarginStart(20)
	widthLabel := gtk.NewLabel("Width:")
	resInputBox.Append(widthLabel)
	a.widthSpin = gtk.NewSpinButtonWithRange(800, 7680, 1)
	a.widthSpin.SetValue(2560)
	resInputBox.Append(a.widthSpin)
	heightLabel := gtk.NewLabel("Height:")
	resInputBox.Append(heightLabel)
	a.heightSpin = gtk.NewSpinButtonWithRange(600, 4320, 1)
	a.heightSpin.SetValue(1440)
	resInputBox.Append(a.heightSpin)
	resBox.Append(resInputBox)
	graphicsBox.Append(resBox)

	fovBox := gtk.NewBox(gtk.OrientationHorizontal, 10)
	a.fovCheck = gtk.NewCheckButtonWithLabel("Custom FOV")
	a.fovCheck.SetActive(false)
	fovBox.Append(a.fovCheck)
	a.fovSpin = gtk.NewSpinButtonWithRange(0.5, 2.5, 0.05)
	a.fovSpin.SetValue(1.0)
	a.fovSpin.SetDigits(2)
	fovSuffix := gtk.NewLabel("degrees")
	fovBox.Append(a.fovSpin)
	fovBox.Append(fovSuffix)
	graphicsBox.Append(fovBox)

	graphicsExpander.SetChild(graphicsBox)
	mainBox.Append(graphicsExpander)

	cameraExpander := gtk.NewExpander("Camera Settings")
	cameraExpander.SetExpanded(false)
	cameraBox := gtk.NewBox(gtk.OrientationVertical, 10)
	cameraBox.SetMarginStart(15)
	cameraBox.SetMarginTop(10)
	cameraBox.SetMarginBottom(10)

	a.camResetCheck = gtk.NewCheckButtonWithLabel("Disable camera reset on lock-on")
	a.camResetCheck.SetActive(false)
	cameraBox.Append(a.camResetCheck)

	cameraExpander.SetChild(cameraBox)
	mainBox.Append(cameraExpander)

	gameplayExpander := gtk.NewExpander("Gameplay Modifications")
	gameplayExpander.SetExpanded(false)
	gameplayBox := gtk.NewBox(gtk.OrientationVertical, 10)
	gameplayBox.SetMarginStart(15)
	gameplayBox.SetMarginTop(10)
	gameplayBox.SetMarginBottom(10)

	a.autoLootCheck = gtk.NewCheckButtonWithLabel("Automatically loot enemies")
	a.autoLootCheck.SetActive(false)
	gameplayBox.Append(a.autoLootCheck)

	a.dragonrotCheck = gtk.NewCheckButtonWithLabel("Prevent dragonrot increase on death")
	a.dragonrotCheck.SetActive(false)
	gameplayBox.Append(a.dragonrotCheck)

	a.deathPenaltyCheck = gtk.NewCheckButtonWithLabel("Disable death penalties (Sen/XP loss)")
	a.deathPenaltyCheck.SetActive(false)
	gameplayBox.Append(a.deathPenaltyCheck)

	gameplayExpander.SetChild(gameplayBox)
	mainBox.Append(gameplayExpander)

	speedExpander := gtk.NewExpander("Speed Modifiers")
	speedExpander.SetExpanded(false)
	speedBox := gtk.NewBox(gtk.OrientationVertical, 10)
	speedBox.SetMarginStart(15)
	speedBox.SetMarginTop(10)
	speedBox.SetMarginBottom(10)

	gameSpeedBox := gtk.NewBox(gtk.OrientationHorizontal, 10)
	a.gameSpeedCheck = gtk.NewCheckButtonWithLabel("Game Speed")
	a.gameSpeedCheck.SetActive(false)
	gameSpeedBox.Append(a.gameSpeedCheck)
	a.gameSpeedSpin = gtk.NewSpinButtonWithRange(0.1, 5.0, 0.1)
	a.gameSpeedSpin.SetValue(1.0)
	a.gameSpeedSpin.SetDigits(1)
	gameSpeedBox.Append(a.gameSpeedSpin)
	speedBox.Append(gameSpeedBox)

	playerSpeedBox := gtk.NewBox(gtk.OrientationHorizontal, 10)
	a.playerSpeedCheck = gtk.NewCheckButtonWithLabel("Player Speed (experimental)")
	a.playerSpeedCheck.SetActive(false)
	a.playerSpeedCheck.SetTooltipText("May not work reliably on Linux/Proton. Use at your own risk.")
	playerSpeedBox.Append(a.playerSpeedCheck)
	a.playerSpeedSpin = gtk.NewSpinButtonWithRange(0.1, 5.0, 0.1)
	a.playerSpeedSpin.SetValue(1.0)
	a.playerSpeedSpin.SetDigits(1)
	playerSpeedBox.Append(a.playerSpeedSpin)
	speedBox.Append(playerSpeedBox)

	speedExpander.SetChild(speedBox)
	mainBox.Append(speedExpander)

	separator2 := gtk.NewSeparator(gtk.OrientationHorizontal)
	separator2.SetMarginTop(10)
	separator2.SetMarginBottom(10)
	mainBox.Append(separator2)

	statsBox := gtk.NewBox(gtk.OrientationHorizontal, 20)
	statsBox.SetMarginTop(5)
	statsBox.SetMarginBottom(5)
	statsBox.SetHAlign(gtk.AlignCenter)

	a.deathsLabel = gtk.NewLabel("Deaths: -")
	a.deathsLabel.AddCSSClass("dim-label")
	statsBox.Append(a.deathsLabel)

	a.killsLabel = gtk.NewLabel("Kills: -")
	a.killsLabel.AddCSSClass("dim-label")
	statsBox.Append(a.killsLabel)

	mainBox.Append(statsBox)

	a.errorsExpander = gtk.NewExpander("Errors")
	a.errorsExpander.SetExpanded(false)
	a.errorsExpander.SetVisible(false)

	errorsScrolled := gtk.NewScrolledWindow()
	errorsScrolled.SetVExpand(false)
	errorsScrolled.SetHExpand(true)
	errorsScrolled.SetSizeRequest(-1, 150)
	errorsScrolled.SetPolicy(gtk.PolicyAutomatic, gtk.PolicyAutomatic)

	a.errorsBuffer = gtk.NewTextBuffer(nil)
	a.errorsView = gtk.NewTextViewWithBuffer(a.errorsBuffer)
	a.errorsView.SetEditable(false)
	a.errorsView.SetMonospace(true)
	a.errorsView.SetWrapMode(gtk.WrapWord)
	a.errorsView.SetMarginTop(5)
	a.errorsView.SetMarginBottom(5)
	a.errorsView.SetMarginStart(5)
	a.errorsView.SetMarginEnd(5)

	errorsScrolled.SetChild(a.errorsView)
	a.errorsExpander.SetChild(errorsScrolled)
	mainBox.Append(a.errorsExpander)

	a.applyButton = gtk.NewButtonWithLabel("Apply Patches")
	a.applyButton.AddCSSClass("suggested-action")
	a.applyButton.SetSensitive(false)
	a.applyButton.ConnectClicked(func() { a.applyPatches() })
	mainBox.Append(a.applyButton)

	a.window.SetChild(mainBox)

	a.loadConfig()

	a.window.SetVisible(true)

	go a.detectGame()
	go a.updateStats()
}

func (a *Application) detectGame() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		pids, err := memory.FindProcessByName(game.ProcessName)
		if err != nil || len(pids) == 0 {
			glib.IdleAdd(func() {
				a.statusLabel.SetText("Waiting for Sekiro...")
				a.pidLabel.SetText("PID: -")
				a.applyButton.SetSensitive(false)
				a.patcher = nil
				a.gamePID = 0
			})
			continue
		}

		pid := pids[0]
		if pid == a.gamePID {
			continue
		}

		patcher, err := game.NewPatcher(pid)
		if err != nil {
			logger.Log.Error("Failed to create patcher", zap.Error(err), zap.Int("pid", pid))
			continue
		}

		a.gamePID = pid
		a.patcher = patcher

		glib.IdleAdd(func() {
			a.statusLabel.SetText("Sekiro detected!")
			a.pidLabel.SetText(fmt.Sprintf("PID: %d", pid))
			a.applyButton.SetSensitive(true)
		})
	}
}

func (a *Application) applyPatches() {
	if a.patcher == nil {
		a.showError("No game detected")
		return
	}

	a.applyButton.SetSensitive(false)
	a.statusLabel.SetText("Applying patches...")

	// Clear previous errors
	glib.IdleAdd(func() {
		a.errorsBuffer.SetText("")
		a.errorsExpander.SetVisible(false)
		a.errorsExpander.SetExpanded(false)
	})

	a.saveConfig()

	go func() {
		var errors []string

		if a.fpsCheck.Active() {
			fps := int(a.fpsSpin.Value())
			if err := a.patcher.ApplyFPSPatch(fps); err != nil {
				errors = append(errors, fmt.Sprintf("FPS: %v", err))
			}
		}

		if a.resCheck.Active() {
			width := int(a.widthSpin.Value())
			height := int(a.heightSpin.Value())
			if err := a.patcher.ApplyResolutionPatch(width, height); err != nil {
				errors = append(errors, fmt.Sprintf("Resolution: %v", err))
			}
		}

		if a.fovCheck.Active() {
			fov := float32(a.fovSpin.Value())
			if err := a.patcher.ApplyFOVPatch(fov); err != nil {
				errors = append(errors, fmt.Sprintf("FOV: %v", err))
			}
		}

		if a.camResetCheck.Active() {
			if err := a.patcher.ApplyCameraResetPatch(true); err != nil {
				errors = append(errors, fmt.Sprintf("Camera reset: %v", err))
			}
		}

		if a.autoLootCheck.Active() {
			if err := a.patcher.ApplyAutoLootPatch(true); err != nil {
				errors = append(errors, fmt.Sprintf("Auto-loot: %v", err))
			}
		}

		if a.dragonrotCheck.Active() {
			if err := a.patcher.ApplyDragonrotPatch(true); err != nil {
				errors = append(errors, fmt.Sprintf("Dragonrot: %v", err))
			}
		}

		if a.deathPenaltyCheck.Active() {
			if err := a.patcher.ApplyDeathPenaltyPatch(true); err != nil {
				errors = append(errors, fmt.Sprintf("Death penalty: %v", err))
			}
		}

		if a.gameSpeedCheck.Active() {
			speed := float32(a.gameSpeedSpin.Value())
			if err := a.patcher.SetGameSpeed(speed); err != nil {
				logger.Log.Warn("Failed to set game speed (may not be available yet)", zap.Error(err))
				errors = append(errors, fmt.Sprintf("Game speed: %v (try again after loading into game)", err))
			}
		}

		if a.playerSpeedCheck.Active() {
			speed := float32(a.playerSpeedSpin.Value())
			if err := a.patcher.SetPlayerSpeed(speed); err != nil {
				logger.Log.Warn("Failed to set player speed (may not be available yet)", zap.Error(err))
				errors = append(errors, "Player speed: not available (load into game first)")
			}
		}

		glib.IdleAdd(func() {
			if len(errors) > 0 {
				a.statusLabel.SetText("Some patches failed (see errors below)")

				// Build error text
				var errorText string
				for i, err := range errors {
					errorText += fmt.Sprintf("[%d] %s\n", i+1, err)
				}

				a.errorsBuffer.SetText(errorText)
				a.errorsExpander.SetVisible(true)
				a.errorsExpander.SetExpanded(true)

				logger.Log.Warn("Patches completed with errors", zap.Int("error_count", len(errors)))
			} else {
				a.statusLabel.SetText("Patches applied successfully!")
				a.errorsExpander.SetVisible(false)
			}
			a.applyButton.SetSensitive(true)
		})
	}()
}

func (a *Application) showError(message string) {
	logger.Log.Error("UI error", zap.String("message", message))
	a.statusLabel.SetText("Error occurred")

	a.errorsBuffer.SetText(message)
	a.errorsExpander.SetVisible(true)
	a.errorsExpander.SetExpanded(true)
}

func (a *Application) updateStats() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if a.patcher == nil {
			glib.IdleAdd(func() {
				a.deathsLabel.SetText("Deaths: -")
				a.killsLabel.SetText("Kills: -")
			})
			continue
		}

		deaths, deathsErr := a.patcher.GetPlayerDeaths()
		kills, killsErr := a.patcher.GetTotalKills()

		glib.IdleAdd(func() {
			if deathsErr == nil {
				a.deathsLabel.SetText(fmt.Sprintf("Deaths: %d", deaths))
			} else {
				a.deathsLabel.SetText("Deaths: -")
			}

			if killsErr == nil {
				a.killsLabel.SetText(fmt.Sprintf("Kills: %d", kills))
			} else {
				a.killsLabel.SetText("Kills: -")
			}
		})
	}
}

func (a *Application) loadConfig() {
	cfg := config.Load()

	a.fpsCheck.SetActive(cfg.FPSUnlock)
	a.fpsSpin.SetValue(float64(cfg.FPS))
	a.resCheck.SetActive(cfg.Resolution)
	a.widthSpin.SetValue(float64(cfg.Width))
	a.heightSpin.SetValue(float64(cfg.Height))
	a.fovCheck.SetActive(cfg.FOV)
	a.fovSpin.SetValue(cfg.FOVValue)
	a.camResetCheck.SetActive(cfg.CameraReset)
	a.autoLootCheck.SetActive(cfg.AutoLoot)
	a.dragonrotCheck.SetActive(cfg.Dragonrot)
	a.deathPenaltyCheck.SetActive(cfg.DeathPenalty)
	a.gameSpeedCheck.SetActive(cfg.GameSpeed)
	a.gameSpeedSpin.SetValue(cfg.GameSpeedVal)
	a.playerSpeedCheck.SetActive(cfg.PlayerSpeed)
	a.playerSpeedSpin.SetValue(cfg.PlayerSpeedVal)
}

func (a *Application) saveConfig() {
	cfg := &config.Config{
		FPSUnlock:      a.fpsCheck.Active(),
		FPS:            int(a.fpsSpin.Value()),
		Resolution:     a.resCheck.Active(),
		Width:          int(a.widthSpin.Value()),
		Height:         int(a.heightSpin.Value()),
		FOV:            a.fovCheck.Active(),
		FOVValue:       a.fovSpin.Value(),
		CameraReset:    a.camResetCheck.Active(),
		AutoLoot:       a.autoLootCheck.Active(),
		Dragonrot:      a.dragonrotCheck.Active(),
		DeathPenalty:   a.deathPenaltyCheck.Active(),
		GameSpeed:      a.gameSpeedCheck.Active(),
		GameSpeedVal:   a.gameSpeedSpin.Value(),
		PlayerSpeed:    a.playerSpeedCheck.Active(),
		PlayerSpeedVal: a.playerSpeedSpin.Value(),
	}
	cfg.Save()
}
