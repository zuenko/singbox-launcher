// Package presentation содержит слой представления визарда конфигурации.
//
// Файл presenter_save.go содержит методы для сохранения конфигурации:
//   - SaveConfig - асинхронное сохранение конфигурации с прогресс-баром и проверками
//
// SaveConfig выполняет следующие шаги:
//  1. Проверяет, что ParserConfig заполнен (хотя бы один proxy с Source или Connections)
//  2. Генерирует конфигурацию из текущей модели (BuildTemplateConfig; без ожидания парсинга outbounds)
//  3. Валидирует конфиг через sing-box по временному файлу, при успехе пишет config.json с бэкапом (SaveConfigWithBackup)
//  4. Сохраняет state.json и показывает диалог успеха; в фоне запускается RunParserProcess (update from subscriptions)
//
// Все операции выполняются асинхронно в отдельной горутине с обновлением прогресс-бара.
//
// Сохранение конфигурации - это отдельная ответственность с сложной логикой.
// Содержит координацию нескольких бизнес-операций (парсинг, генерация, сохранение).
// Управляет прогресс-баром и диалогами на разных этапах сохранения.
//
// Используется в:
//   - wizard.go - SaveConfig вызывается при нажатии кнопки "Save" в визарде
//
// Использует:
//   - business/create_config.go - BuildTemplateConfig для генерации конфигурации
//   - business/saver.go - SaveConfigWithBackup для сохранения файла
//   - core.RunParserProcess - обновление конфига из подписок после сохранения (в фоне)
package presentation

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"singbox-launcher/core"
	"singbox-launcher/core/config"
	"singbox-launcher/internal/debuglog"
	"singbox-launcher/internal/dialogs"
	wizardbusiness "singbox-launcher/ui/wizard/business"
	wizardmodels "singbox-launcher/ui/wizard/models"
)

// SaveConfig сохраняет конфигурацию асинхронно с прогресс-баром.
func (p *WizardPresenter) SaveConfig() {
	p.SyncGUIToModel()

	// Validate input before starting save operation
	if !p.validateSaveInput() {
		return
	}

	// Check if save operation is already in progress
	if !p.checkSaveOperationState() {
		return
	}

	debuglog.InfoLog("SaveConfig: starting save operation")

	// Устанавливаем флаг синхронно ДО запуска горутины, чтобы избежать race condition
	p.guiState.SaveInProgress = true
	p.SetSaveState("", 0.0)

	go p.executeSaveOperation()
}

// validateSaveInput проверяет входные данные перед сохранением.
// Only ParserConfig.ParserConfig.Proxies is the source of truth; at least one proxy must have Source or Connections.
func (p *WizardPresenter) validateSaveInput() bool {
	if strings.TrimSpace(p.model.ParserConfigJSON) == "" {
		debuglog.WarnLog("SaveConfig: ParserConfig is empty")
		dialog.ShowError(fmt.Errorf("ParserConfig is empty"), p.guiState.Window)
		return false
	}
	var pc config.ParserConfig
	if err := json.Unmarshal([]byte(p.model.ParserConfigJSON), &pc); err != nil {
		debuglog.WarnLog("SaveConfig: ParserConfig JSON invalid: %v", err)
		dialog.ShowError(fmt.Errorf("ParserConfig is invalid: %w", err), p.guiState.Window)
		return false
	}
	for _, px := range pc.ParserConfig.Proxies {
		if strings.TrimSpace(px.Source) != "" || len(px.Connections) > 0 {
			return true
		}
	}
	debuglog.WarnLog("SaveConfig: no proxy with source or connections in ParserConfig")
	dialog.ShowError(fmt.Errorf("Add at least one source: use the Sources tab (Add) or add proxies in ParserConfig on the Outbounds tab."), p.guiState.Window)
	return false
}

// checkSaveOperationState проверяет состояние операции сохранения.
func (p *WizardPresenter) checkSaveOperationState() bool {
	if p.guiState.SaveInProgress {
		debuglog.WarnLog("SaveConfig: Save operation already in progress")
		dialog.ShowInformation("Saving", "Save operation already in progress... Please wait.", p.guiState.Window)
		return false
	}
	return true
}

// executeSaveOperation выполняет операцию сохранения в отдельной горутине.
// Save does not wait for outbounds parsing: it writes current model state (existing GeneratedOutbounds or empty).
// After save, config update from subscriptions is triggered in background (RunParserProcess).
func (p *WizardPresenter) executeSaveOperation() {
	defer p.finalizeSaveOperation()

	// Step 1: Build config from current model (0.2-0.4)
	configText, err := p.buildConfigForSave()
	if err != nil {
		return
	}

	// Step 2: Validate (по временному файлу) и запись config.json (0.4-0.6)
	configPath, err := p.saveConfigFile(configText)
	if err != nil {
		return
	}

	// Step 3: Save state.json and show success dialog (0.6-0.9)
	p.saveStateAndShowSuccessDialog(configPath)

	// Step 4: Completion (0.9-1.0)
	p.completeSaveOperation()
}

// finalizeSaveOperation завершает операцию сохранения и восстанавливает UI.
func (p *WizardPresenter) finalizeSaveOperation() {
	debuglog.InfoLog("SaveConfig: save operation completed (or failed)")
	p.UpdateSaveStatusText("")
	// Всегда восстанавливаем кнопку Save, даже при ошибке
	p.SetSaveState("Save", -1)
	// Сбрасываем флаг парсинга на случай, если он завис
	if p.model.AutoParseInProgress {
		p.model.AutoParseInProgress = false
	}
}

// buildConfigForSave строит конфигурацию из шаблона и модели.
// Возвращает текст конфигурации или ошибку.
func (p *WizardPresenter) buildConfigForSave() (string, error) {
	p.UpdateSaveStatusText("Building config...")
	p.UpdateSaveProgress(0.2)

	// Check if save operation was cancelled
	if !p.guiState.SaveInProgress {
		debuglog.DebugLog("presenter_save: Save operation cancelled before building config")
		return "", fmt.Errorf("save operation cancelled")
	}

	debuglog.InfoLog("SaveConfig: building template config")
	text, err := wizardbusiness.BuildTemplateConfig(p.model, false)
	if err != nil {
		debuglog.ErrorLog("SaveConfig: BuildTemplateConfig failed: %v", err)
		p.UpdateUI(func() {
			dialog.ShowError(err, p.guiState.Window)
		})
		return "", err
	}

	debuglog.InfoLog("SaveConfig: template config built successfully, length: %d", len(text))
	p.UpdateSaveStatusText("Preparing...")
	p.UpdateSaveProgress(0.4)
	return text, nil
}

// saveConfigFile валидирует конфиг через sing-box (по временному файлу) и при успехе пишет config.json с бэкапом.
// Возвращает путь к сохранённому файлу или ошибку (в т.ч. ошибку валидации).
func (p *WizardPresenter) saveConfigFile(configText string) (string, error) {
	if !p.guiState.SaveInProgress {
		debuglog.DebugLog("presenter_save: Save operation cancelled before saving file")
		return "", fmt.Errorf("save operation cancelled")
	}

	p.UpdateSaveStatusText("Validating...")
	p.UpdateSaveProgress(0.45)

	ac := core.GetController()
	if ac == nil || ac.FileService == nil {
		debuglog.WarnLog("SaveConfig: controller or FileService not available")
		p.UpdateUI(func() {
			dialog.ShowError(fmt.Errorf("controller not available"), p.guiState.Window)
		})
		return "", fmt.Errorf("controller not available")
	}

	fileService := &wizardbusiness.FileServiceAdapter{FileService: ac.FileService}
	debuglog.InfoLog("SaveConfig: validating then saving config file")
	path, err := wizardbusiness.SaveConfigWithBackup(fileService, configText)
	if err != nil {
		debuglog.ErrorLog("SaveConfig: SaveConfigWithBackup failed: %v", err)
		p.UpdateUI(func() {
			p.showSaveErrorDialog(err)
		})
		return "", err
	}

	debuglog.InfoLog("SaveConfig: config saved to %s", path)
	p.UpdateSaveStatusText("Saving state...")
	p.UpdateSaveProgress(0.6)
	return path, nil
}

// showSaveErrorDialog показывает ошибку сохранения; при ValidationError — диалог с кнопкой «Копировать конфиг».
func (p *WizardPresenter) showSaveErrorDialog(err error) {
	var valErr *wizardbusiness.ValidationError
	if errors.As(err, &valErr) && valErr.ConfigText != "" {
		p.showValidationErrorDialog(valErr)
		return
	}
	dialog.ShowError(err, p.guiState.Window)
}

// showValidationErrorDialog показывает ошибку валидации и кнопку «Копировать конфиг» в буфер обмена.
func (p *WizardPresenter) showValidationErrorDialog(valErr *wizardbusiness.ValidationError) {
	if p.guiState.Window == nil {
		return
	}
	msg := valErr.Error()
	messageLabel := widget.NewLabel(msg)
	messageLabel.Wrapping = fyne.TextWrapWord

	var d dialog.Dialog
	copyBtn := widget.NewButton("Copy config", func() {
		if app := fyne.CurrentApp(); app != nil && app.Clipboard() != nil {
			app.Clipboard().SetContent(valErr.ConfigText)
			if p.guiState.Window != nil {
				dialogs.ShowAutoHideInfo(app, p.guiState.Window, "Copied", "Config copied to clipboard.")
			}
		}
	})
	copyBtn.Importance = widget.MediumImportance
	closeBtn := widget.NewButton("Close", func() {
		if d != nil {
			d.Hide()
		}
	})
	closeBtn.Importance = widget.HighImportance

	buttons := container.NewHBox(layout.NewSpacer(), copyBtn, closeBtn)
	content := container.NewBorder(nil, buttons, nil, nil, messageLabel)
	d = dialog.NewCustomWithoutButtons("Validation failed", content, p.guiState.Window)
	d.Show()
}

// saveStateAndShowSuccessDialog сохраняет state.json и показывает диалог успешного сохранения.
// Вызывается только после успешной валидации и записи config.json, поэтому диалог всегда с «Validation: Passed».
func (p *WizardPresenter) saveStateAndShowSuccessDialog(configPath string) {
	// Check if save operation was cancelled
	if !p.guiState.SaveInProgress {
		debuglog.DebugLog("presenter_save: Save operation cancelled before saving state")
		return
	}
	p.UpdateSaveStatusText("Saving state...")
	p.UpdateSaveProgress(0.7)

	ac := core.GetController()
	// Получаем путь к state.json для логирования
	statesDir := filepath.Join(ac.FileService.ExecDir, "bin", wizardbusiness.WizardStatesDir)
	statePath := filepath.Join(statesDir, wizardmodels.StateFileName)

	p.UpdateUI(func() {
		// Update config status in Core Dashboard
		if ac.UIService != nil && ac.UIService.UpdateConfigStatusFunc != nil {
			ac.UIService.UpdateConfigStatusFunc()
		}

		// Сохраняем текущее состояние в state.json после успешного сохранения конфигурации
		// Сохранение происходит всегда, независимо от hasChanges
		debuglog.InfoLog("SaveConfig: saving state.json to %s", statePath)
		if err := p.SaveCurrentState(); err != nil {
			debuglog.WarnLog("presenter_save: failed to save state after config save: %v", err)
		} else {
			debuglog.InfoLog("SaveConfig: state.json saved successfully to %s", statePath)
		}

		// Логируем итоговую информацию о сохранении
		debuglog.InfoLog("SaveConfig: completed - config.json=%s, state.json=%s", configPath, statePath)

		p.showSaveSuccessDialog(configPath)
	})
	p.UpdateSaveStatusText("Done")
	p.UpdateSaveProgress(0.9)
}

// showSaveSuccessDialog показывает диалог успешного сохранения (вызывается только после успешной валидации).
func (p *WizardPresenter) showSaveSuccessDialog(configPath string) {
	message := fmt.Sprintf("Config written to %s\n\n✅ Validation: Passed", configPath)
	title := "Config Saved"

	// Create dialog with OK button that closes both dialog and wizard
	var d dialog.Dialog
	okButton := widget.NewButton("OK", func() {
		// Close dialog first
		if d != nil {
			d.Hide()
		}
		// Close wizard window only (not the main application)
		if p.guiState.Window != nil {
			p.guiState.Window.Close()
		}
	})
	okButton.Importance = widget.HighImportance

	buttonsRow := container.NewHBox(
		layout.NewSpacer(),
		okButton,
	)

	messageLabel := widget.NewLabel(message)

	d = dialogs.NewCustom(title, messageLabel, buttonsRow, "", p.guiState.Window)
	d.Show()
}

// completeSaveOperation завершает операцию сохранения с небольшой задержкой.
// Triggers config update from subscriptions in background (same as "Update" on main tab).
func (p *WizardPresenter) completeSaveOperation() {
	debuglog.InfoLog("SaveConfig: triggering config update from subscriptions (background)")
	go core.RunParserProcess()
	<-time.After(100 * time.Millisecond)
	p.UpdateSaveProgress(1.0)
	<-time.After(200 * time.Millisecond)
}
