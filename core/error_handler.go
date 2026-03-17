package core

import (
	"fmt"

	"singbox-launcher/internal/debuglog"
	"singbox-launcher/internal/dialogs"
	"singbox-launcher/internal/locale"
)

// showErrorUI logs the error and shows it in the UI if available.
// category is used as a log prefix (e.g. "StartupError", "ParserError").
func (ac *AppController) showErrorUI(category string, err error) {
	debuglog.ErrorLog("%s: %v", category, err)
	if ac.hasUI() {
		dialogs.ShowError(ac.UIService.MainWindow, err)
	}
}

// ShowStartupError shows an error when sing-box fails to start.
func (ac *AppController) ShowStartupError(err error) {
	ac.showErrorUI("StartupError", fmt.Errorf("%s", locale.Tf("error.startup", err.Error())))
}

// ShowParserError shows an error when parser fails.
func (ac *AppController) ShowParserError(err error) {
	ac.showErrorUI("ParserError", fmt.Errorf("%s", locale.Tf("error.parser", err.Error())))
}
