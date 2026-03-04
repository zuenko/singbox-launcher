// Package tabs содержит UI компоненты для табов визарда конфигурации.
//
// Файл source_tab.go содержит функции, создающие UI табов визарда:
//   - Вкладка Sources: ввод URL, проверка, список источников и Preview сгенерированных нод/селекторов
//   - Вкладка Outbounds and ParserConfig: редактор ParserConfig JSON и вход в конфигуратор outbounds
//
// Каждый таб визарда имеет свою отдельную ответственность и логику UI.
//
// Используется в:
//   - wizard.go - при создании окна визарда, вызывается CreateSourceTab(presenter)
//
// Взаимодействует с:
//   - presenter - все действия пользователя (нажатия кнопок, ввод текста) обрабатываются через методы presenter
//   - business - вызывает CheckURL, ParseAndPreview через presenter
package tabs

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"singbox-launcher/core/config"
	"singbox-launcher/internal/debuglog"
	"singbox-launcher/internal/platform"
	wizardbusiness "singbox-launcher/ui/wizard/business"
	wizarddialogs "singbox-launcher/ui/wizard/dialogs"
	"singbox-launcher/ui/wizard/outbounds_configurator"
	wizardpresentation "singbox-launcher/ui/wizard/presentation"
)

// CreateSourcesTab creates the Sources tab UI (URLs, URL status and preview).
func CreateSourcesTab(presenter *wizardpresentation.WizardPresenter) fyne.CanvasObject {
	guiState := presenter.GUIState()
	model := presenter.Model()

	// Section 1: Subscription URL or Direct Links
	// We perform automatic URL checking on input change (debounced) instead of
	// requiring the user to click a "Check" button.
	// Add a padding placeholder container on the right to keep layout similar.
	paddingRect := canvas.NewRectangle(color.Transparent)
	paddingRect.SetMinSize(fyne.NewSize(10, 0)) // 10px padding on right
	guiState.CheckURLContainer = container.NewHBox(
		paddingRect,
	)

	urlLabel := widget.NewLabel("Subscription URL or Direct Links:")
	urlLabel.Importance = widget.MediumImportance

	guiState.SourceURLEntry = widget.NewMultiLineEntry()
	guiState.SourceURLEntry.SetPlaceHolder("https://example.com/subscription\nor\nvless://...\nvmess://...\nhysteria2://...\nssh://...")
	guiState.SourceURLEntry.Wrapping = fyne.TextWrapOff
	guiState.SourceURLEntry.OnChanged = func(value string) {
		model := presenter.Model()
		model.PreviewNeedsParse = true
		trimmed := strings.TrimSpace(value)
		if err := wizardbusiness.ApplyURLToParserConfig(model, presenter, trimmed); err != nil {
			debuglog.ErrorLog("source_tab: error applying URL to ParserConfig: %v", err)
		}

		// Debounce CheckURL: cancel previous timer and set a new one (2s after last change)
		if guiState.CheckURLTimer != nil {
			guiState.CheckURLTimer.Stop()
			guiState.CheckURLTimer = nil
		}

		// Define the actual check logic as a reusable closure so we can reschedule
		var doCheck func(string)
		doCheck = func(v string) {
			// This runs in goroutine from timer - coordinate with UI thread for state
			fyne.Do(func() {
				// If a check is currently in progress, reschedule after delay
				if guiState.CheckURLInProgress {
					// reschedule
					guiState.CheckURLTimer = time.AfterFunc(2*time.Second, func() { doCheck(v) })
					return
				}
				// Mark in-progress and sync
				guiState.CheckURLInProgress = true
				presenter.SyncGUIToModel()
				// Run the check in background
				go func() {
					if err := wizardbusiness.CheckURL(presenter.Model(), presenter); err != nil {
						debuglog.ErrorLog("source_tab: CheckURL failed: %v", err)
					}
					// Clear in-progress flag
					fyne.Do(func() { guiState.CheckURLInProgress = false })
				}()
			})
		}

		// Schedule the check after debounce interval
		guiState.CheckURLTimer = time.AfterFunc(2*time.Second, func() { doCheck(trimmed) })
	}

	// Hint under input field with Check button on right
	hintLabel := widget.NewLabel("Supports subscription URLs (http/https) or direct links (vless://, vmess://, trojan://, ss://, hysteria2://, ssh://). For multiple links, use a new line for each.")
	hintLabel.Wrapping = fyne.TextWrapWord

	getFreeVPNButton := widget.NewButton("Get free VPN!", func() {
		wizarddialogs.ShowGetFreeVPNDialog(presenter)
	})

	hintRow := container.NewBorder(
		nil,                        // top
		nil,                        // bottom
		nil,                        // left
		guiState.CheckURLContainer, // right - actions
		hintLabel,                  // center - hint takes all available space
	)

	guiState.URLStatusLabel = widget.NewLabel("")
	guiState.URLStatusLabel.Wrapping = fyne.TextWrapWord

	// Limit width and height of URL input field (3 lines)
	// Wrap MultiLineEntry in Scroll container to show scrollbars
	urlEntryScroll := container.NewScroll(guiState.SourceURLEntry)
	urlEntryScroll.Direction = container.ScrollBoth
	// Create dummy Rectangle to set size (height 3 lines, width limited)
	urlEntrySizeRect := canvas.NewRectangle(color.Transparent)
	urlEntrySizeRect.SetMinSize(fyne.NewSize(0, 60)) // Width 900px, height ~3 lines (approx 20px per line)
	// Wrap in Max container with Rectangle to fix size
	// Scroll container will be limited by this size and show scrollbars when content doesn't fit
	urlEntryWithSize := container.NewMax(
		urlEntrySizeRect,
		urlEntryScroll,
	)

	// Header row with action on the right
	urlHeader := container.NewHBox(
		urlLabel,
		layout.NewSpacer(),
		getFreeVPNButton,
	)

	urlContainer := container.NewVBox(
		urlHeader,               // Header with action
		urlEntryWithSize,        // Input field with size limit (3 lines)
		hintRow,                 // Hint with button on right
		guiState.URLStatusLabel, // Status
	)

	// Section 2: Sources list (based on ParserConfig.ParserConfig.Proxies)
	sourcesLabel := widget.NewLabel("Sources")
	sourcesLabel.Importance = widget.MediumImportance

	sourcesBox := container.NewVBox()

	refreshSourcesList := func() {
		sourcesBox.Objects = sourcesBox.Objects[:0]

		if model.ParserConfig == nil || len(model.ParserConfig.ParserConfig.Proxies) == 0 {
			sourcesBox.Add(widget.NewLabel("No sources defined in ParserConfig."))
			sourcesBox.Refresh()
			return
		}

		for i, proxy := range model.ParserConfig.ParserConfig.Proxies {
			label := proxy.Source
			if label == "" {
				label = fmt.Sprintf("Source %d", i+1)
			}
			if len(label) > 40 {
				label = label[:37] + "..."
			}

			fullURL := proxy.Source
			tagPrefix := proxy.TagPrefix
			tagPostfix := proxy.TagPostfix
			tagMask := proxy.TagMask

			localTags := make([]string, 0, len(proxy.Outbounds))
			for _, ob := range proxy.Outbounds {
				if ob.Tag != "" {
					localTags = append(localTags, ob.Tag)
				}
			}

			tooltipLines := []string{
				fmt.Sprintf("URL: %s", fullURL),
				fmt.Sprintf("tag_prefix: %s", tagPrefix),
				fmt.Sprintf("tag_postfix: %s", tagPostfix),
				fmt.Sprintf("tag_mask: %s", tagMask),
				fmt.Sprintf("local outbounds: %d", len(localTags)),
			}
			if len(localTags) > 0 {
				tooltipLines = append(tooltipLines, "tags: "+strings.Join(localTags, ", "))
			}
			tooltipText := strings.Join(tooltipLines, "\n")

			// Use a regular button as a label-like widget that supports SetToolTip via fyne-tooltip.
			sourceButton := widget.NewButton(label, nil)
			sourceButton.Importance = widget.LowImportance

			if tb, ok := interface{}(sourceButton).(interface{ SetToolTip(string) }); ok {
				tb.SetToolTip(tooltipText)
			}

			row := container.NewHBox(
				sourceButton,
				layout.NewSpacer(),
			)
			sourcesBox.Add(row)
		}

		sourcesBox.Refresh()
	}

	// Ensure sources list is initialized from current model state
	refreshSourcesList()
	guiState.RefreshSourcesList = refreshSourcesList

	sourcesScroll := container.NewVScroll(sourcesBox)
	sourcesScroll.SetMinSize(fyne.NewSize(0, 140))

	// Section 3: Preview Generated Outbounds
	previewLabel := widget.NewLabel("Preview")
	previewLabel.Importance = widget.MediumImportance

	// Use Entry without Disable for black text, but make it read-only via OnChanged
	guiState.OutboundsPreview = widget.NewMultiLineEntry()
	guiState.OutboundsPreview.SetPlaceHolder("Generated outbounds will appear here after clicking Parse...")
	guiState.OutboundsPreview.Wrapping = fyne.TextWrapOff
	previewText := "Generated outbounds will appear here after clicking Parse..."
	guiState.OutboundsPreview.SetText(previewText)
	guiState.OutboundsPreviewLastText = previewText
	// Make field effectively read-only: ignore programmatic updates, restore last preview on user edits
	guiState.OutboundsPreview.OnChanged = func(text string) {
		if guiState.OutboundsPreviewUpdating {
			// Ignore programmatic updates
			return
		}
		// Restore last known preview text
		if guiState.OutboundsPreviewLastText != "" {
			guiState.OutboundsPreview.SetText(guiState.OutboundsPreviewLastText)
		} else {
			guiState.OutboundsPreview.SetText(previewText)
		}
	}

	// Limit width and height of Preview field
	previewScroll := container.NewScroll(guiState.OutboundsPreview)
	previewScroll.Direction = container.ScrollBoth
	// Create dummy Rectangle to set height via container.NewMax
	previewHeightRect := canvas.NewRectangle(color.Transparent)
	previewHeightRect.SetMinSize(fyne.NewSize(0, 90)) // ~8-9 lines (reduced by ~30px)
	// Wrap in Max container with Rectangle to fix height
	previewWithHeight := container.NewMax(
		previewHeightRect,
		previewScroll,
	)

	previewContainer := container.NewVBox(
		previewLabel,
		previewWithHeight,
	)

	// Combine all sections
	content := container.NewVBox(
		widget.NewSeparator(),
		urlContainer,
		widget.NewSeparator(),
		sourcesLabel,
		sourcesScroll,
		widget.NewSeparator(),
		previewContainer,
		widget.NewSeparator(),
	)

	// Add scroll for long content
	scrollContainer := container.NewScroll(content)
	scrollContainer.SetMinSize(fyne.NewSize(0, 620))

	return scrollContainer
}

// CreateOutboundsAndParserConfigTab creates the Outbounds and ParserConfig tab UI.
// For now it reuses the existing ParserConfig editor and Config Outbounds button;
// later it will be extended to embed the outbounds configurator list directly.
func CreateOutboundsAndParserConfigTab(presenter *wizardpresentation.WizardPresenter) fyne.CanvasObject {
	guiState := presenter.GUIState()
	model := presenter.Model()

	// ParserConfig multi-line editor
	guiState.ParserConfigEntry = widget.NewMultiLineEntry()
	guiState.ParserConfigEntry.SetPlaceHolder("Enter ParserConfig JSON here...")
	guiState.ParserConfigEntry.Wrapping = fyne.TextWrapOff
	guiState.ParserConfigEntry.OnChanged = func(string) {
		if guiState.ParserConfigUpdating {
			return
		}
		model := presenter.Model()
		model.PreviewNeedsParse = true
		// Sync GUI to model to update ParserConfigJSON before refreshing outbound options
		presenter.SyncGUIToModel()
		presenter.RefreshOutboundOptions()
		// Preview status will be updated when switching to Preview tab
	}

	// Limit width and height of ParserConfig field
	parserConfigScroll := container.NewScroll(guiState.ParserConfigEntry)
	parserConfigScroll.Direction = container.ScrollBoth
	parserHeightRect := canvas.NewRectangle(color.Transparent)
	parserHeightRect.SetMinSize(fyne.NewSize(0, 200)) // ~10 lines
	parserConfigWithHeight := container.NewMax(
		parserHeightRect,
		parserConfigScroll,
	)

	// Documentation button
	docButton := widget.NewButton("📖 Documentation", func() {
		docURL := "https://github.com/Leadaxe/singbox-launcher/blob/main/docs/ParserConfig.md"
		if err := platform.OpenURL(docURL); err != nil {
			dialog.ShowError(fmt.Errorf("failed to open documentation: %w", err), guiState.Window)
		}
	})

	parserLabel := widget.NewLabel("ParserConfig:")
	parserLabel.Importance = widget.MediumImportance

	// Embedded outbounds configurator: use model.ParserConfig so edits apply in place.
	pc := model.ParserConfig
	if pc == nil {
		pc = &config.ParserConfig{}
		raw := strings.TrimSpace(model.ParserConfigJSON)
		if raw != "" {
			if err := json.Unmarshal([]byte(raw), pc); err != nil {
				debuglog.DebugLog("source_tab: initial parse of ParserConfigJSON failed: %v", err)
			}
		}
		model.ParserConfig = pc
	}

	onConfiguratorApply := func() {
		serialized, err := wizardbusiness.SerializeParserConfig(pc)
		if err != nil {
			debuglog.ErrorLog("source_tab: SerializeParserConfig after configurator change: %v", err)
			dialog.ShowError(fmt.Errorf("Failed to serialize ParserConfig: %w", err), guiState.Window)
			return
		}
		model.ParserConfigJSON = serialized
		model.ParserConfig = pc
		model.PreviewNeedsParse = true
		presenter.UpdateParserConfig(serialized)
		presenter.RefreshOutboundOptions()
		if guiState.RefreshSourcesList != nil {
			guiState.RefreshSourcesList()
		}
	}

	configuratorContent := outbounds_configurator.NewConfiguratorContent(guiState.Window, pc, onConfiguratorApply)

	// No Parse button on this tab per SPEC: update is automatic via configurator callback and tab switch (Rules/Preview).
	headerRow := container.NewHBox(
		parserLabel,
		layout.NewSpacer(),
		docButton,
	)

	parserContainer := container.NewVBox(
		headerRow,
		parserConfigWithHeight,
		widget.NewSeparator(),
		configuratorContent,
	)

	content := container.NewVBox(
		widget.NewSeparator(),
		parserContainer,
		widget.NewSeparator(),
	)

	scrollContainer := container.NewScroll(content)
	scrollContainer.SetMinSize(fyne.NewSize(0, 620))

	return scrollContainer
}

// CreateSourceTab is kept for backward compatibility and currently returns the Sources tab content.
func CreateSourceTab(presenter *wizardpresentation.WizardPresenter) fyne.CanvasObject {
	return CreateSourcesTab(presenter)
}
