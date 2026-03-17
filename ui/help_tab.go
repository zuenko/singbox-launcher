package ui

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	ttwidget "github.com/dweymouth/fyne-tooltip/widget"

	"singbox-launcher/core"
	"singbox-launcher/internal/constants"
	"singbox-launcher/internal/debuglog"
	"singbox-launcher/internal/locale"
	"singbox-launcher/internal/platform"
)

// CreateHelpTab creates and returns the content for the "Help" tab.
func CreateHelpTab(ac *core.AppController) fyne.CanvasObject {
	configButton := widget.NewButton(locale.T("help.open_config_folder"), func() {
		binDir := platform.GetBinDir(ac.FileService.ExecDir)
		if err := platform.OpenFolder(binDir); err != nil {
			debuglog.ErrorLog("toolsTab: Failed to open config folder: %v", err)
			ShowError(ac.UIService.MainWindow, err)
		}
	})
	killButton := widget.NewButton(locale.T("help.kill_singbox"), func() {
		go func() {
			processName := platform.GetProcessNameForCheck()
			_ = platform.KillProcess(processName)
			fyne.Do(func() {
				ShowAutoHideInfo(ac.UIService.Application, ac.UIService.MainWindow,
					locale.T("help.kill_title"), locale.T("help.kill_result"))
				ac.RunningState.Set(false)
			})
		}()
	})

	// Version and links section
	versionLabel := widget.NewLabel(locale.Tf("help.version_label", constants.AppVersion))
	versionLabel.Alignment = fyne.TextAlignCenter

	// Launcher update status
	launcherUpdateLabel := widget.NewLabel(locale.T("help.checking_updates"))
	launcherUpdateLabel.Alignment = fyne.TextAlignCenter
	launcherUpdateLabel.Wrapping = fyne.TextWrapWord

	// Update launcher version info
	updateLauncherVersionInfo := func() {
		latest := ac.GetCachedLauncherVersion()
		current := constants.AppVersion

		if latest == "" {
			launcherUpdateLabel.SetText(locale.T("help.unable_to_check_updates"))
			return
		}

		currentClean := strings.TrimPrefix(current, "v")
		latestClean := strings.TrimPrefix(latest, "v")

		compareResult := core.CompareVersions(currentClean, latestClean)
		if compareResult < 0 {
			launcherUpdateLabel.SetText(locale.Tf("help.update_available_format", latest, current))
		} else if compareResult > 0 {
			launcherUpdateLabel.SetText(locale.Tf("help.dev_build_format", current, latest))
		} else {
			launcherUpdateLabel.SetText(locale.Tf("help.latest_version_format", current))
		}
	}

	updateLauncherVersionInfo()

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for i := 0; i < 10; i++ {
			select {
			case <-ticker.C:
				if platform.IsSleeping() {
					continue
				}
				fyne.Do(func() {
					if ac.GetCachedLauncherVersion() == "" {
						updateLauncherVersionInfo()
					} else {
						updateLauncherVersionInfo()
						return
					}
				})
			}
		}
	}()

	telegramLink := widget.NewHyperlink(locale.T("help.telegram_link"), nil)
	_ = telegramLink.SetURLFromString("https://t.me/singbox_launcher")
	telegramLink.OnTapped = func() {
		if err := platform.OpenURL("https://t.me/singbox_launcher"); err != nil {
			debuglog.ErrorLog("toolsTab: Failed to open Telegram link: %v", err)
			ShowError(ac.UIService.MainWindow, err)
		}
	}

	githubLink := widget.NewHyperlink(locale.T("help.github_link"), nil)
	_ = githubLink.SetURLFromString("https://github.com/Leadaxe/singbox-launcher")
	githubLink.OnTapped = func() {
		if err := platform.OpenURL("https://github.com/Leadaxe/singbox-launcher"); err != nil {
			debuglog.ErrorLog("toolsTab: Failed to open GitHub link: %v", err)
			ShowError(ac.UIService.MainWindow, err)
		}
	}

	// Language selector
	langLabel := widget.NewLabel(locale.T("help.language_label"))
	langSelect := widget.NewSelect(locale.LangDisplayNames(), func(selected string) {
		code := locale.LangCodeByDisplayName(selected)
		if code == "" || code == locale.GetLang() {
			return
		}
		locale.SetLang(code)
		binDir := platform.GetBinDir(ac.FileService.ExecDir)
		if err := locale.SaveSettings(binDir, locale.Settings{Lang: code}); err != nil {
			debuglog.ErrorLog("helpTab: Failed to save language setting: %v", err)
		}
		debuglog.InfoLog("helpTab: Language changed to %q", code)
		ShowInfo(ac.UIService.MainWindow, locale.T("help.language_label"),
			fmt.Sprintf("%s\n\n%s", locale.LangDisplayName(code), locale.T("help.language_changed")))
	})
	langSelect.Selected = locale.LangDisplayName(locale.GetLang())

	// Download translations button (compact icon + tooltip to avoid widening the window)
	downloadLocalesBtn := ttwidget.NewButton(locale.T("help.download_locales_btn"), nil)
	downloadLocalesBtn.SetToolTip(locale.T("help.download_locales"))
	downloadLocalesBtn.OnTapped = func() {
		downloadLocalesBtn.Disable()
		downloadLocalesBtn.SetText(locale.T("help.downloading_locales_btn"))
		go func() {
			binDir := platform.GetBinDir(ac.FileService.ExecDir)
			localeDir := locale.GetLocaleDir(binDir)
			count, err := locale.DownloadAllRemoteLocales(localeDir)
			fyne.Do(func() {
				downloadLocalesBtn.Enable()
				downloadLocalesBtn.SetText(locale.T("help.download_locales_btn"))
				if err != nil && count == 0 {
					// Use unified manual-download dialog, same as for core template/SRS/etc.
					downloadURL := ""
					if len(locale.RemoteLanguages) > 0 {
						downloadURL = locale.GetLocaleURL(locale.RemoteLanguages[0])
					}
					ShowDownloadFailedManual(
						ac.UIService.MainWindow,
						locale.T("help.download_locales_failed"),
						downloadURL,
						localeDir,
					)
					return
				}
				// Refresh language selector with newly loaded languages
				langSelect.Options = locale.LangDisplayNames()
				langSelect.Selected = locale.LangDisplayName(locale.GetLang())
				langSelect.Refresh()
				ShowInfo(ac.UIService.MainWindow, locale.T("help.language_label"),
					locale.Tf("help.download_locales_success", count))
			})
		}()
	}

	langRow := container.NewHBox(langLabel, langSelect, downloadLocalesBtn)

	return container.NewVBox(
		configButton,
		killButton,
		widget.NewSeparator(),
		versionLabel,
		launcherUpdateLabel,
		widget.NewSeparator(),
		container.NewHBox(
			layout.NewSpacer(),
			telegramLink,
			widget.NewLabel(" | "),
			githubLink,
			layout.NewSpacer(),
		),
		widget.NewSeparator(),
		container.NewHBox(layout.NewSpacer(), langRow, layout.NewSpacer()),
	)
}
