// edit_dialog.go provides the Add/Edit outbound dialog for the configurator.
// The dialog is shown as a separate window (like the Add Rule dialog).
package outbounds_configurator

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"singbox-launcher/core/config"
	"singbox-launcher/internal/platform"
	wizardbusiness "singbox-launcher/ui/wizard/business"
)

// ShowEditDialog opens a separate window to add or edit an outbound. existing may be nil for add.
// ParserConfig is taken from the model (editPresenter.Model()) so the dialog always uses current sources.
// onSave is called with the new config, scopeKind ("global" or "source") and sourceIndex (when scope is source).
// editPresenter is required (Model() is used to get ParserConfig); when set, only one Edit/Add window is allowed.
func ShowEditDialog(
	parent fyne.Window,
	editPresenter OutboundEditPresenter,
	existing *config.OutboundConfig,
	isGlobal bool,
	sourceIndex int,
	existingTags []string,
	onSave func(updated *config.OutboundConfig, scopeKind string, sourceIndex int),
) {
	if editPresenter != nil {
		if w := editPresenter.OpenOutboundEditWindow(); w != nil {
			w.RequestFocus()
			return
		}
	}
	parserConfig := getParserConfig(editPresenter.Model())
	if parserConfig == nil {
		dialog.ShowError(fmt.Errorf("ParserConfig is not available"), parent)
		return
	}
	isAdd := existing == nil
	dialogTitle := "Edit Outbound"
	if isAdd {
		dialogTitle = "Add Outbound"
	}

	tagEntry := widget.NewEntry()
	if existing != nil {
		tagEntry.SetText(existing.Tag)
	}
	tagEntry.SetPlaceHolder("e.g. proxy-out")

	typeSelect := widget.NewSelect([]string{"manual (selector)", "auto (urltest)"}, nil)
	if existing != nil {
		if existing.Type == "urltest" {
			typeSelect.SetSelected("auto (urltest)")
		} else {
			typeSelect.SetSelected("manual (selector)")
		}
	} else {
		typeSelect.SetSelected("manual (selector)")
	}

	commentEntry := widget.NewEntry()
	if existing != nil {
		commentEntry.SetText(existing.Comment)
	}
	commentEntry.SetPlaceHolder("Optional comment")

	// Scope: For all | For source: ...
	scopeOptions := []string{"For all"}
	for i, p := range parserConfig.ParserConfig.Proxies {
		label := p.Source
		if label == "" {
			label = "Source " + strconv.Itoa(i+1)
		}
		if len(label) > 35 {
			label = label[:32] + "..."
		}
		scopeOptions = append(scopeOptions, "For source: "+label)
	}
	scopeSelect := widget.NewSelect(scopeOptions, nil)
	if isAdd {
		scopeSelect.SetSelected("For all")
	} else if isGlobal {
		scopeSelect.SetSelected("For all")
	} else {
		if sourceIndex >= 0 && sourceIndex < len(parserConfig.ParserConfig.Proxies) {
			scopeSelect.SetSelected(scopeOptions[sourceIndex+1])
		} else {
			scopeSelect.SetSelected(scopeOptions[0])
		}
	}

	// Filters: fixed key "tag", value editable
	filterKeyLabel := widget.NewLabel("tag")
	filterValEntry := widget.NewEntry()
	filterValEntry.SetPlaceHolder("e.g. /🇳🇱/i or !/(🇷🇺)/i")
	if existing != nil && existing.Filters != nil {
		if v, ok := existing.Filters["tag"]; ok {
			if s, ok := v.(string); ok {
				filterValEntry.SetText(s)
			}
		} else {
			for _, v := range existing.Filters {
				if s, ok := v.(string); ok {
					filterValEntry.SetText(s)
					break
				}
			}
		}
	}

	// Preferred default: fixed key "tag", value editable
	defKeyLabel := widget.NewLabel("tag")
	defValEntry := widget.NewEntry()
	defValEntry.SetPlaceHolder("e.g. /🇳🇱/i")
	if existing != nil && existing.PreferredDefault != nil {
		if v, ok := existing.PreferredDefault["tag"]; ok {
			if s, ok := v.(string); ok {
				defValEntry.SetText(s)
			}
		} else {
			for _, v := range existing.PreferredDefault {
				if s, ok := v.(string); ok {
					defValEntry.SetText(s)
					break
				}
			}
		}
	}

	// AddOutbounds: direct-out, reject checkboxes + checkboxes for other tags
	directCheck := widget.NewCheck("direct-out", nil)
	rejectCheck := widget.NewCheck("reject", nil)
	otherTagChecks := make([]*widget.Check, 0, len(existingTags))
	otherTagsMap := make(map[string]*widget.Check)
	for _, tag := range existingTags {
		c := widget.NewCheck(tag, nil)
		otherTagChecks = append(otherTagChecks, c)
		otherTagsMap[tag] = c
	}
	if existing != nil && len(existing.AddOutbounds) > 0 {
		for _, t := range existing.AddOutbounds {
			if t == "direct-out" {
				directCheck.SetChecked(true)
			} else if t == "reject" {
				rejectCheck.SetChecked(true)
			} else if c, ok := otherTagsMap[t]; ok {
				c.SetChecked(true)
			}
		}
	}

	otherTagsBox := container.NewVBox()
	for _, c := range otherTagChecks {
		otherTagsBox.Add(c)
	}
	scrollOther := container.NewScroll(otherTagsBox)
	scrollOther.SetMinSize(fyne.NewSize(0, 80))

	// Raw tab: editable JSON (valid outbound object)
	initialConfig := existing
	if initialConfig == nil {
		initialConfig = &config.OutboundConfig{
			Tag:           "",
			Type:          "selector",
			Comment:       "",
			Options:       map[string]interface{}{"interrupt_exist_connections": true},
			AddOutbounds:  nil,
		}
	}
	rawJSONBytes, _ := json.MarshalIndent(initialConfig, "", "  ")
	rawEntry := widget.NewMultiLineEntry()
	rawEntry.SetText(string(rawJSONBytes))
	rawEntry.Wrapping = fyne.TextWrapOff
	rawEntry.SetMinRowsVisible(16)
	rawScroll := container.NewScroll(rawEntry)
	rawScroll.SetMinSize(fyne.NewSize(400, 360))

	// Raw documentation button (opens ParserConfig.md "Секция outbounds")
	rawDocButton := widget.NewButton("📖 Documentation", func() {
		docURL := "https://github.com/Leadaxe/singbox-launcher/blob/main/docs/ParserConfig.md#%D1%81%D0%B5%D0%BA%D1%86%D0%B8%D1%8F-outbounds"
		if err := platform.OpenURL(docURL); err != nil {
			dialog.ShowError(fmt.Errorf("failed to open documentation: %w", err), parent)
		}
	})
	rawHeader := container.NewHBox(
		widget.NewLabel("Raw outbound JSON"),
		layout.NewSpacer(),
		rawDocButton,
	)
	rawContainer := container.NewBorder(
		rawHeader,
		nil,
		nil,
		nil,
		rawScroll,
	)

	var currentTab string = "settings"

	var dialogWin fyne.Window
	getScopeFromForm := func() (scopeKind string, idx int) {
		scopeKind = "global"
		idx = -1
		if scopeSelect.Selected != "" && strings.HasPrefix(scopeSelect.Selected, "For source:") {
			scopeKind = "source"
			for i, opt := range scopeOptions {
				if i > 0 && opt == scopeSelect.Selected {
					idx = i - 1
					break
				}
			}
		}
		return scopeKind, idx
	}
	// buildConfigForPreview builds a config.OutboundConfig snapshot based on current UI state.
	// It is used by the Preview tab; errors are returned to be shown inline.
	buildConfigForPreview := func() (*config.OutboundConfig, error) {
		if currentTab == "raw" {
			var cfg config.OutboundConfig
			if err := json.Unmarshal([]byte(rawEntry.Text), &cfg); err != nil {
				return nil, fmt.Errorf("invalid outbound JSON: %w", err)
			}
			if strings.TrimSpace(cfg.Tag) == "" {
				return nil, fmt.Errorf("tag is required")
			}
			return &cfg, nil
		}

		tag := strings.TrimSpace(tagEntry.Text)
		if tag == "" {
			return nil, fmt.Errorf("tag is required")
		}
		obType := "selector"
		if typeSelect.Selected == "auto (urltest)" {
			obType = "urltest"
		}

		cfg := &config.OutboundConfig{
			Tag:     tag,
			Type:    obType,
			Comment: strings.TrimSpace(commentEntry.Text),
		}
		if existing != nil && existing.Options != nil {
			cfg.Options = make(map[string]interface{})
			for k, v := range existing.Options {
				cfg.Options[k] = v
			}
		} else if obType == "selector" {
			cfg.Options = map[string]interface{}{"interrupt_exist_connections": true}
		} else {
			cfg.Options = map[string]interface{}{
				"url": "https://cp.cloudflare.com/generate_204",
				"interval": "5m", "tolerance": 100,
				"interrupt_exist_connections": true,
			}
		}

		filterVal := strings.TrimSpace(filterValEntry.Text)
		if filterVal != "" {
			cfg.Filters = map[string]interface{}{"tag": filterVal}
		}
		defVal := strings.TrimSpace(defValEntry.Text)
		if defVal != "" {
			cfg.PreferredDefault = map[string]interface{}{"tag": defVal}
		}

		var addOb []string
		if directCheck.Checked {
			addOb = append(addOb, "direct-out")
		}
		if rejectCheck.Checked {
			addOb = append(addOb, "reject")
		}
		for _, c := range otherTagChecks {
			if c.Checked {
				addOb = append(addOb, c.Text)
			}
		}
		cfg.AddOutbounds = addOb

		return cfg, nil
	}

	save := func() {
		if currentTab == "raw" {
			var cfg config.OutboundConfig
			if err := json.Unmarshal([]byte(rawEntry.Text), &cfg); err != nil {
				dialog.ShowError(fmt.Errorf("invalid JSON: %w", err), dialogWin)
				return
			}
			if strings.TrimSpace(cfg.Tag) == "" {
				dialog.ShowError(fmt.Errorf("tag is required"), dialogWin)
				return
			}
			scopeKind, idx := getScopeFromForm()
			if existing != nil && existing.Wizard != nil {
				cfg.Wizard = wizardbusiness.CloneOutbound(existing).Wizard
			}
			onSave(&cfg, scopeKind, idx)
			if dialogWin != nil {
				dialogWin.Close()
			}
			return
		}

		cfg, err := buildConfigForPreview()
		if err != nil {
			dialog.ShowError(err, dialogWin)
			return
		}
		scopeKind, idx := getScopeFromForm()

		// Preserve wizard if editing
		if existing != nil && existing.Wizard != nil {
			cfg.Wizard = wizardbusiness.CloneOutbound(existing).Wizard
		}
		onSave(cfg, scopeKind, idx)
		if dialogWin != nil {
			dialogWin.Close()
		}
	}

	form := container.NewVBox(
		widget.NewLabel("Scope"),
		scopeSelect,
		widget.NewLabel("Tag"),
		tagEntry,
		widget.NewLabel("Type"),
		typeSelect,
		widget.NewLabel("Comment"),
		commentEntry,
		widget.NewLabel("Filters (key and value; use !/regex/i for negation)"),
		container.NewGridWithColumns(2, filterKeyLabel, filterValEntry),
		widget.NewLabel("Preferred default (filter for default node)"),
		container.NewGridWithColumns(2, defKeyLabel, defValEntry),
		widget.NewLabel("Add outbounds at start (direct-out, reject, others)"),
		container.NewHBox(directCheck, rejectCheck),
		scrollOther,
	)
	// Right margin inside scroll so the scrollbar does not overlap form elements
	const scrollbarGap = 20
	rightGap := canvas.NewRectangle(color.Transparent)
	rightGap.SetMinSize(fyne.NewSize(scrollbarGap, 0))
	formWithGap := container.NewBorder(nil, nil, nil, rightGap, form)
	widthSpacer := canvas.NewRectangle(color.Transparent)
	widthSpacer.SetMinSize(fyne.NewSize(400, 0))
	scrollContent := container.NewMax(widthSpacer, formWithGap)
	dialogScroll := container.NewScroll(scrollContent)
	dialogScroll.SetMinSize(fyne.NewSize(400, 400))

	// Preview tab: uses preview cache from the wizard model (via editPresenter.Model()).
	previewStatusLabel := widget.NewLabel("Switch to Preview to see nodes for this outbound.")
	type previewRow struct {
		text  string
		color color.Color
	}
	var previewRows []previewRow
	previewList := widget.NewList(
		func() int { return len(previewRows) },
		func() fyne.CanvasObject { return canvas.NewText("", color.White) },
		func(id int, o fyne.CanvasObject) {
			if id < 0 || id >= len(previewRows) {
				return
			}
			if txt, ok := o.(*canvas.Text); ok {
				txt.Text = previewRows[id].text
				txt.Color = previewRows[id].color
			}
		},
	)
	previewListScroll := container.NewScroll(previewList)
	previewListScroll.SetMinSize(fyne.NewSize(400, 320))
	previewContent := container.NewBorder(
		previewStatusLabel,
		nil,
		nil,
		nil,
		previewListScroll,
	)

	buildPreview := func() {
		previewRows = nil
		previewList.Refresh()

		if editPresenter == nil {
			previewStatusLabel.SetText("Preview is not available in this context (no presenter).")
			return
		}
		model := editPresenter.Model()
		if model == nil {
			previewStatusLabel.SetText("Preview is not available: wizard model is nil.")
			return
		}

		cfg, err := buildConfigForPreview()
		if err != nil {
			previewStatusLabel.SetText("Failed to build preview: invalid outbound JSON. Please switch to the \"Raw\" tab and fix JSON first.")
			return
		}

		// Ensure preview cache is up to date.
		errorCount, err := wizardbusiness.RebuildPreviewCache(model)
		if err != nil {
			previewStatusLabel.SetText(fmt.Sprintf("Failed to build preview cache: %v", err))
			return
		}
		allNodes := model.PreviewNodes
		if len(allNodes) == 0 {
			previewStatusLabel.SetText("No nodes available for preview. Please configure sources and try again.")
			return
		}

		// Use core/config helper to get filtered nodes and default tag, consistent with generator logic.
		filteredNodes, defaultTag := config.PreviewSelectorNodes(allNodes, *cfg)
		filteredSet := make(map[*config.ParsedNode]bool, len(filteredNodes))
		for _, n := range filteredNodes {
			filteredSet[n] = true
		}

		// Map node pointer to source label using PreviewNodesBySource and ParserConfig.
		sourceLabels := make(map[*config.ParsedNode]string)
		if model.ParserConfig != nil && model.PreviewNodesBySource != nil {
			for si, nodes := range model.PreviewNodesBySource {
				if si < 0 || si >= len(model.ParserConfig.ParserConfig.Proxies) {
					continue
				}
				proxy := model.ParserConfig.ParserConfig.Proxies[si]
				label := proxy.Source
				if label == "" {
					label = fmt.Sprintf("Source %d", si+1)
				}
				if len(label) > 40 {
					label = label[:37] + "..."
				}
				for _, n := range nodes {
					sourceLabels[n] = label
				}
			}
		}

		// Build rows: default node first, then the rest in original allNodes order.
		defaultRows := make([]previewRow, 0)
		otherRows := make([]previewRow, 0, len(allNodes))

		for _, node := range allNodes {
			inSelector := filteredSet[node]
			isDefault := inSelector && node.Tag == defaultTag

			src := sourceLabels[node]
			if src == "" {
				src = "Unknown source"
			}
			text := node.Tag
			if text == "" {
				// Fallback formatting when tag is empty.
				if node.Label != "" {
					text = node.Label
				} else if node.Server != "" {
					text = fmt.Sprintf("%s:%d", node.Server, node.Port)
				} else {
					text = node.Scheme
				}
			}
			text = fmt.Sprintf("%s — %s", text, src)
			if isDefault {
				text = "[default] " + text
			}

			var rowColor color.Color
			switch {
			case isDefault:
				rowColor = color.RGBA{R: 0, G: 128, B: 255, A: 255} // blue
			case inSelector:
				rowColor = color.RGBA{R: 0, G: 160, B: 0, A: 255} // green
			default:
				rowColor = color.RGBA{R: 200, G: 0, B: 0, A: 255} // red
			}

			row := previewRow{text: text, color: rowColor}
			if isDefault {
				defaultRows = append(defaultRows, row)
			} else {
				otherRows = append(otherRows, row)
			}
		}

		previewRows = append(defaultRows, otherRows...)
		previewList.Refresh()

		status := fmt.Sprintf("%d node(s) total, %d matched filters", len(allNodes), len(filteredNodes))
		if defaultTag != "" {
			status += fmt.Sprintf(", default: %s", defaultTag)
		}
		if len(cfg.AddOutbounds) > 0 {
			status += fmt.Sprintf(" | Also includes: %s", strings.Join(cfg.AddOutbounds, ", "))
		}
		if errorCount > 0 {
			status += fmt.Sprintf(" | ⚠️ %d source error(s)", errorCount)
		}
		previewStatusLabel.SetText(status)
	}

	// syncRawToForm parses the Raw tab JSON and updates Settings form fields (tag, type, comment, filters, etc.).
	// Called when user switches from Raw to Settings so the form reflects the raw JSON.
	syncRawToForm := func() {
		var cfg config.OutboundConfig
		if err := json.Unmarshal([]byte(rawEntry.Text), &cfg); err != nil {
			return // invalid JSON: leave form as is
		}
		if strings.TrimSpace(cfg.Tag) == "" {
			return
		}
		tagEntry.SetText(cfg.Tag)
		if cfg.Type == "urltest" {
			typeSelect.SetSelected("auto (urltest)")
		} else {
			typeSelect.SetSelected("manual (selector)")
		}
		commentEntry.SetText(cfg.Comment)
		filterValEntry.SetText("")
		if cfg.Filters != nil {
			if v, ok := cfg.Filters["tag"]; ok {
				if s, ok := v.(string); ok {
					filterValEntry.SetText(s)
				}
			}
		}
		defValEntry.SetText("")
		if cfg.PreferredDefault != nil {
			if v, ok := cfg.PreferredDefault["tag"]; ok {
				if s, ok := v.(string); ok {
					defValEntry.SetText(s)
				}
			}
		}
		directCheck.SetChecked(false)
		rejectCheck.SetChecked(false)
		for _, c := range otherTagChecks {
			c.SetChecked(false)
		}
		if len(cfg.AddOutbounds) > 0 {
			for _, t := range cfg.AddOutbounds {
				if t == "direct-out" {
					directCheck.SetChecked(true)
				} else if t == "reject" {
					rejectCheck.SetChecked(true)
				} else if c, ok := otherTagsMap[t]; ok {
					c.SetChecked(true)
				}
			}
		}
	}

	tabs := container.NewAppTabs(
		container.NewTabItem("Settings", dialogScroll),
		container.NewTabItem("Raw", rawContainer),
		container.NewTabItem("Preview", previewContent),
	)
	tabs.OnSelected = func(t *container.TabItem) {
		switch t.Text {
		case "Raw":
			currentTab = "raw"
		case "Preview":
			currentTab = "preview"
			buildPreview()
		default:
			currentTab = "settings"
			syncRawToForm()
		}
	}

	cancelBtn := widget.NewButton("Cancel", func() {
		if dialogWin != nil {
			dialogWin.Close()
		}
	})
	saveBtn := widget.NewButton("Save", func() { save() })

	buttonsContainer := container.NewHBox(
		layout.NewSpacer(),
		cancelBtn,
		saveBtn,
	)
	mainContent := container.NewBorder(
		nil,
		buttonsContainer,
		nil,
		nil,
		tabs,
	)

	app := fyne.CurrentApp()
	if app == nil {
		return
	}
	dialogWin = app.NewWindow(dialogTitle)
	if editPresenter != nil {
		editPresenter.SetOutboundEditWindow(dialogWin)
		dialogWin.SetOnClosed(func() {
			editPresenter.ClearOutboundEditWindow()
			editPresenter.UpdateChildOverlay()
		})
	}
	dialogWin.Resize(fyne.NewSize(440, 560))
	dialogWin.CenterOnScreen()
	dialogWin.SetContent(mainContent)
	dialogWin.Show()
	if editPresenter != nil {
		editPresenter.UpdateChildOverlay()
	}
}
