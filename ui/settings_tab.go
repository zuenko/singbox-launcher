package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	ttwidget "github.com/dweymouth/fyne-tooltip/widget"

	"singbox-launcher/core"
	"singbox-launcher/internal/debuglog"
	"singbox-launcher/internal/dialogs"
	"singbox-launcher/internal/locale"
	"singbox-launcher/internal/platform"
)

// CreateSettingsTab builds the Settings tab. Collects launcher-wide toggles
// that used to be scattered across Core Dashboard (auto-update, auto-ping)
// and Help (language + download-locales), so there's one obvious place to
// look for "change launcher behavior".
//
// Settings persist to bin/settings.json via locale.LoadSettings /
// locale.SaveSettings with load-mutate-save — we explicitly avoid the
// `Settings{Lang: code}` "fresh struct" anti-pattern which silently wiped
// every other field.
func CreateSettingsTab(ac *core.AppController) fyne.CanvasObject {
	binDir := platform.GetBinDir(ac.FileService.ExecDir)

	// ---- Subscriptions section ---------------------------------------------
	subsTitle := widget.NewLabelWithStyle(locale.T("settings.section_subscriptions"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	autoUpdateCheck := widget.NewCheck(locale.T("core.auto_update_subs_label"), nil)
	autoUpdateCheck.SetChecked(ac.StateService.IsAutoUpdateEnabled())
	autoUpdateCheck.OnChanged = func(enabled bool) {
		ac.StateService.SetAutoUpdateEnabled(enabled)
		if enabled {
			ac.StateService.ResetAutoUpdateFailedAttempts()
		}
		st := locale.LoadSettings(binDir)
		st.SubscriptionAutoUpdateDisabled = !enabled
		if err := locale.SaveSettings(binDir, st); err != nil {
			debuglog.WarnLog("settings_tab: save subscription_auto_update_disabled: %v", err)
		}
	}

	autoPingCheck := widget.NewCheck(locale.T("core.auto_ping_label"), nil)
	autoPingCheck.SetChecked(ac.StateService.IsAutoPingAfterConnectEnabled())
	autoPingCheck.OnChanged = func(enabled bool) {
		ac.StateService.SetAutoPingAfterConnectEnabled(enabled)
		st := locale.LoadSettings(binDir)
		st.AutoPingAfterConnectDisabled = !enabled
		if err := locale.SaveSettings(binDir, st); err != nil {
			debuglog.WarnLog("settings_tab: save auto_ping_after_connect_disabled: %v", err)
		}
	}

	// ---- Language section --------------------------------------------------
	langTitle := widget.NewLabelWithStyle(locale.T("settings.section_language"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	langLabel := widget.NewLabel(locale.T("help.language_label"))
	langSelect := widget.NewSelect(locale.LangDisplayNames(), nil)
	langSelect.Selected = locale.LangDisplayName(locale.GetLang())

	downloadLocalesBtn := ttwidget.NewButton(locale.T("help.download_locales_btn"), nil)
	downloadLocalesBtn.SetToolTip(locale.T("help.download_locales"))

	langSelect.OnChanged = func(selected string) {
		code := locale.LangCodeByDisplayName(selected)
		if code == "" || code == locale.GetLang() {
			return
		}
		locale.SetLang(code)
		// load-mutate-save so we don't clobber other settings fields
		st := locale.LoadSettings(binDir)
		st.Lang = code
		if err := locale.SaveSettings(binDir, st); err != nil {
			debuglog.ErrorLog("settings_tab: save lang: %v", err)
		}
		ShowInfo(ac.UIService.MainWindow, locale.T("help.language_label"),
			fmt.Sprintf("%s\n\n%s", locale.LangDisplayName(code), locale.T("help.language_changed")))
	}

	downloadLocalesBtn.OnTapped = func() {
		downloadLocalesBtn.Disable()
		downloadLocalesBtn.SetText(locale.T("help.downloading_locales_btn"))
		go func() {
			localeDir := locale.GetLocaleDir(binDir)
			count, err := locale.DownloadAllRemoteLocales(localeDir)
			fyne.Do(func() {
				downloadLocalesBtn.Enable()
				downloadLocalesBtn.SetText(locale.T("help.download_locales_btn"))
				if err != nil && count == 0 {
					downloadURL := ""
					if len(locale.RemoteLanguages) > 0 {
						downloadURL = locale.GetLocaleURL(locale.RemoteLanguages[0])
					}
					dialogs.ShowDownloadFailedManual(
						ac.UIService.MainWindow,
						locale.T("help.download_locales_failed"),
						downloadURL,
						localeDir,
					)
					return
				}
				langSelect.Options = locale.LangDisplayNames()
				langSelect.Selected = locale.LangDisplayName(locale.GetLang())
				langSelect.Refresh()
				ShowInfo(ac.UIService.MainWindow, locale.T("help.language_label"),
					locale.Tf("help.download_locales_success", count))
			})
		}()
	}

	// langSelect stretches; button stays compact on the right.
	langRow := container.NewBorder(nil, nil, langLabel, downloadLocalesBtn, langSelect)

	content := container.NewVBox(
		subsTitle,
		autoUpdateCheck,
		autoPingCheck,
		widget.NewSeparator(),
		langTitle,
		langRow,
	)
	return container.NewPadded(content)
}
