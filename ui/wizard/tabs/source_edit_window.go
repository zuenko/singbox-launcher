package tabs

import (
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"singbox-launcher/core/config"
	"singbox-launcher/internal/debuglog"
	"singbox-launcher/internal/locale"
	wizardbusiness "singbox-launcher/ui/wizard/business"
	wizardmodels "singbox-launcher/ui/wizard/models"
	wizardpresentation "singbox-launcher/ui/wizard/presentation"
	wizardutils "singbox-launcher/ui/wizard/utils"
)

// Min heights for Source Edit dialog tab bodies (child window; do not use main window canvas before Show).
const (
	sourceEditSettingsScrollMinH float32 = 260
	sourceEditJSONScrollMinH     float32 = 380
)

func showWizardTagConflictError(win fyne.Window) {
	dialog.ShowError(errors.New(locale.T("wizard.source.wizard_tag_conflict")), win)
}

func setFyneWidgetToolTip(w fyne.CanvasObject, tip string) {
	if tb, ok := interface{}(w).(interface{ SetToolTip(string) }); ok {
		tb.SetToolTip(tip)
	}
}

// formatLocalOutboundPreviewLine is a one-line summary for proxies[i].outbounds[] in the Edit → Preview tab.
func formatLocalOutboundPreviewLine(ob *config.OutboundConfig) string {
	if ob == nil {
		return ""
	}
	typ := ob.Type
	if typ == "" {
		typ = "?"
	}
	comment := strings.TrimSpace(ob.Comment)
	rs := []rune(comment)
	const maxR = 96
	if len(rs) > maxR {
		comment = string(rs[:maxR-1]) + "…"
	}
	if ob.Tag == "" {
		return fmt.Sprintf("[%s]  %s", typ, comment)
	}
	return fmt.Sprintf("%s  [%s]  %s", ob.Tag, typ, comment)
}

func serializeParserAfterSourceEdit(
	presenter *wizardpresentation.WizardPresenter,
	guiState *wizardpresentation.GUIState,
	m *wizardmodels.WizardModel,
	errParent fyne.Window,
) error {
	serialized, err := wizardbusiness.SerializeParserConfig(m.ParserConfig)
	if err != nil {
		debuglog.ErrorLog("source_edit: SerializeParserConfig: %v", err)
		if errParent != nil {
			dialog.ShowError(err, errParent)
		}
		return err
	}
	m.ParserConfigJSON = serialized
	m.PreviewNeedsParse = true
	wizardbusiness.InvalidatePreviewCache(m)
	presenter.UpdateParserConfig(serialized)
	presenter.ScheduleRefreshOutboundOptionsDebounced()
	presenter.MarkAsChanged()
	if guiState.RefreshSourcesList != nil {
		guiState.RefreshSourcesList()
	}
	return nil
}

// showSourceEditWindow opens Settings | Preview | JSON for one proxy source (SPEC 026).
func showSourceEditWindow(
	presenter *wizardpresentation.WizardPresenter,
	guiState *wizardpresentation.GUIState,
	parent fyne.Window,
	sourceIndex int,
	shortLabel string,
) {
	if presenter == nil {
		return
	}
	// One modal child workflow: finish Outbound Edit or another Source Edit (View slot) first.
	if w := presenter.OpenOutboundEditWindow(); w != nil {
		w.RequestFocus()
		return
	}
	if w := presenter.OpenViewWindow(); w != nil {
		w.RequestFocus()
		return
	}
	presenter.MergeGUIToModel()

	app := fyne.CurrentApp()
	if app == nil {
		return
	}
	m := presenter.Model()
	if m == nil {
		return
	}
	if err := wizardbusiness.EnsureWizardModelParserConfig(m); err != nil {
		debuglog.ErrorLog("source_edit: EnsureWizardModelParserConfig: %v", err)
		if parent != nil {
			dialog.ShowError(err, parent)
		}
		return
	}
	if sourceIndex < 0 || sourceIndex >= len(m.ParserConfig.ParserConfig.Proxies) {
		return
	}

	title := locale.Tf("wizard.source.edit_title", shortLabel)
	title = wizardutils.TruncateStringEllipsis(title, wizardutils.MaxLabelRunes, "...")
	win := app.NewWindow(title)
	presenter.SetViewWindow(win)
	win.SetOnClosed(func() {
		presenter.ClearViewWindow()
		presenter.UpdateChildOverlay()
	})

	proxyRef := func() *config.ProxySource {
		mm := presenter.Model()
		if mm == nil || mm.ParserConfig == nil || sourceIndex >= len(mm.ParserConfig.ParserConfig.Proxies) {
			return nil
		}
		return &mm.ParserConfig.ParserConfig.Proxies[sourceIndex]
	}

	prefixEntry := widget.NewEntry()
	prefixEntry.SetPlaceHolder(locale.T("wizard.source.placeholder_prefix"))

	autoCheck := widget.NewCheck(locale.T("wizard.source.local_auto"), nil)
	selectCheck := widget.NewCheck(locale.T("wizard.source.local_select"), nil)
	excludeCheck := widget.NewCheck(locale.T("wizard.source.exclude_global"), nil)
	exposeCheck := widget.NewCheck(locale.T("wizard.source.expose_tags"), nil)
	hintLabel := widget.NewLabel("")
	hintLabel.Wrapping = fyne.TextWrapWord

	var afterSync func()

	var exposeOnChanged func(bool)
	exposeOnChanged = func(v bool) {
		if exposeCheck.Disabled() {
			return
		}
		pp := proxyRef()
		if pp == nil {
			return
		}
		pp.ExposeGroupTagsToGlobal = v
		_ = serializeParserAfterSourceEdit(presenter, guiState, presenter.Model(), win)
		if afterSync != nil {
			afterSync()
		}
	}
	exposeCheck.OnChanged = exposeOnChanged

	refreshExposeAvailability := func() {
		p := proxyRef()
		if p == nil {
			return
		}
		has := wizardbusiness.ProxyHasLocalAuto(p) || wizardbusiness.ProxyHasLocalSelect(p)
		exposeCheck.OnChanged = nil
		if has {
			exposeCheck.Enable()
			exposeCheck.SetChecked(p.ExposeGroupTagsToGlobal)
		} else {
			exposeCheck.Disable()
			exposeCheck.SetChecked(false)
		}
		exposeCheck.OnChanged = exposeOnChanged
		tip := locale.T("wizard.source.expose_tags_tooltip")
		if has {
			tip = ""
		}
		setFyneWidgetToolTip(exposeCheck, tip)
	}

	refreshExcludeHint := func() {
		p := proxyRef()
		if p == nil {
			return
		}
		if p.ExcludeFromGlobal && (!wizardbusiness.ProxyHasLocalAuto(p) || !wizardbusiness.ProxyHasLocalSelect(p)) {
			hintLabel.SetText(locale.T("wizard.source.exclude_hint"))
			hintLabel.Show()
		} else {
			hintLabel.SetText("")
			hintLabel.Hide()
		}
	}

	syncFormFromModel := func() {
		p := proxyRef()
		if p == nil {
			return
		}
		prefixEntry.SetText(p.TagPrefix)
		autoCheck.SetChecked(wizardbusiness.ProxyHasLocalAuto(p))
		selectCheck.SetChecked(wizardbusiness.ProxyHasLocalSelect(p))
		excludeCheck.SetChecked(p.ExcludeFromGlobal)
		refreshExposeAvailability()
		refreshExcludeHint()
		if afterSync != nil {
			afterSync()
		}
	}

	prefixEntry.OnChanged = func(s string) {
		p := proxyRef()
		if p == nil {
			return
		}
		p.TagPrefix = strings.TrimSpace(s)
		wizardbusiness.RenameWizardLocalOutboundTags(p, sourceIndex)
		_ = serializeParserAfterSourceEdit(presenter, guiState, presenter.Model(), win)
		syncFormFromModel()
	}

	autoCheck.OnChanged = func(on bool) {
		p := proxyRef()
		if p == nil {
			return
		}
		if on {
			if err := wizardbusiness.EnsureLocalAuto(p, sourceIndex); err != nil {
				autoCheck.SetChecked(false)
				showWizardTagConflictError(win)
				return
			}
		} else {
			wizardbusiness.RemoveWizardSelectOutbounds(p)
			wizardbusiness.RemoveWizardAutoOutbounds(p)
			wizardbusiness.SyncExposeFlagWhenNoLocalGroups(p)
		}
		_ = serializeParserAfterSourceEdit(presenter, guiState, presenter.Model(), win)
		syncFormFromModel()
	}

	selectCheck.OnChanged = func(on bool) {
		p := proxyRef()
		if p == nil {
			return
		}
		if on {
			if err := wizardbusiness.EnsureLocalSelect(p, sourceIndex); err != nil {
				selectCheck.SetChecked(false)
				showWizardTagConflictError(win)
				return
			}
		} else {
			wizardbusiness.RemoveWizardSelectOutbounds(p)
			wizardbusiness.SyncExposeFlagWhenNoLocalGroups(p)
		}
		_ = serializeParserAfterSourceEdit(presenter, guiState, presenter.Model(), win)
		syncFormFromModel()
	}

	excludeCheck.OnChanged = func(v bool) {
		p := proxyRef()
		if p == nil {
			return
		}
		p.ExcludeFromGlobal = v
		_ = serializeParserAfterSourceEdit(presenter, guiState, presenter.Model(), win)
		refreshExcludeHint()
		if afterSync != nil {
			afterSync()
		}
	}

	settingsContent := container.NewVBox(
		widget.NewLabel(locale.T("wizard.source.label_prefix")),
		prefixEntry,
		widget.NewSeparator(),
		autoCheck,
		selectCheck,
		excludeCheck,
		exposeCheck,
		hintLabel,
	)
	settingsScroll := container.NewVScroll(settingsContent)
	settingsScroll.SetMinSize(fyne.NewSize(0, sourceEditSettingsScrollMinH))

	previewStatus := widget.NewLabel(locale.T("wizard.source.preview_loading"))
	previewListHost := container.NewMax()
	previewBox := container.NewBorder(previewStatus, nil, nil, nil, previewListHost)

	previewRefreshSeq := 0
	refreshPreviewTab := func() {
		previewRefreshSeq++
		seq := previewRefreshSeq
		previewStatus.SetText(locale.T("wizard.source.preview_loading"))
		previewListHost.Objects = nil
		previewListHost.Add(layout.NewSpacer())
		previewListHost.Refresh()
		go func() {
			model := presenter.Model()
			var nodes []*config.ParsedNode
			var err error
			if model != nil && model.PreviewNodesBySource != nil {
				if nn, ok := model.PreviewNodesBySource[sourceIndex]; ok {
					nodes = nn
				}
			}
			if len(nodes) == 0 && model != nil {
				_, cacheErr := wizardbusiness.RebuildPreviewCache(model)
				if cacheErr != nil {
					err = cacheErr
				} else if model.PreviewNodesBySource != nil {
					nodes = model.PreviewNodesBySource[sourceIndex]
				}
			}
			if len(nodes) == 0 && err == nil {
				pp := proxyRef()
				if pp != nil {
					nodes, err = fetchAndParseSource(pp.Source, pp.Skip)
				}
			}
			fyne.Do(func() {
				if seq != previewRefreshSeq {
					return
				}
				previewListHost.Objects = nil
				p := proxyRef()
				localLines := make([]string, 0, 8)
				if p != nil {
					for i := range p.Outbounds {
						localLines = append(localLines, formatLocalOutboundPreviewLine(&p.Outbounds[i]))
					}
				}
				if err != nil {
					previewStatus.SetText(locale.Tf("wizard.source.preview_status_err", len(localLines), err.Error()))
				} else {
					previewStatus.SetText(locale.Tf("wizard.source.preview_status_ok", len(localLines), len(nodes)))
				}
				body := container.NewVBox()
				lblLocal := widget.NewLabel(locale.T("wizard.source.preview_section_local"))
				lblLocal.TextStyle = fyne.TextStyle{Bold: true}
				body.Add(lblLocal)
				if len(localLines) == 0 {
					body.Add(widget.NewLabel(locale.T("wizard.source.preview_no_local_outbounds")))
				} else {
					ll := localLines
					localList := widget.NewList(
						func() int { return len(ll) },
						func() fyne.CanvasObject { return widget.NewLabel("") },
						func(id int, o fyne.CanvasObject) {
							o.(*widget.Label).SetText(ll[id])
						},
					)
					h := float32(24 * len(localLines))
					if h > 140 {
						h = 140
					}
					if h < 48 {
						h = 48
					}
					scLoc := container.NewScroll(localList)
					scLoc.SetMinSize(fyne.NewSize(0, h))
					body.Add(scLoc)
				}
				body.Add(widget.NewSeparator())
				lblSrv := widget.NewLabel(locale.T("wizard.source.preview_section_servers"))
				lblSrv.TextStyle = fyne.TextStyle{Bold: true}
				body.Add(lblSrv)
				// On parse error, previewStatus already shows preview_status_err; no duplicate line here.
				if err == nil {
					if len(nodes) == 0 {
						body.Add(widget.NewLabel(locale.T("wizard.source.view_no_servers")))
					} else {
						nn := nodes
						srvList := widget.NewList(
							func() int { return len(nn) },
							func() fyne.CanvasObject { return widget.NewLabel("") },
							func(id int, o fyne.CanvasObject) {
								o.(*widget.Label).SetText(nodeDisplayLine(nn[id]))
							},
						)
						sc := container.NewScroll(srvList)
						sc.SetMinSize(fyne.NewSize(0, 220))
						body.Add(sc)
					}
				}
				previewListHost.Add(container.NewVScroll(body))
				previewListHost.Refresh()
			})
		}()
	}

	// JSON: same pattern as wizard Preview tab — MultiLineEntry inside Max + VScroll (no duplicate tab title).
	jsonEntry := widget.NewMultiLineEntry()
	jsonEntry.Wrapping = fyne.TextWrapOff
	jsonEntry.OnChanged = func(string) { /* display-only; changes are not saved */ }
	jsonScroll := container.NewVScroll(container.NewMax(
		canvas.NewRectangle(color.Transparent),
		jsonEntry,
	))
	jsonScroll.SetMinSize(fyne.NewSize(0, sourceEditJSONScrollMinH))

	refreshJSONTab := func() {
		p := proxyRef()
		if p == nil {
			jsonEntry.SetText("")
			return
		}
		b, jerr := json.MarshalIndent(p, "", "  ")
		if jerr != nil {
			jsonEntry.SetText(jerr.Error())
			return
		}
		jsonEntry.SetText(string(b))
	}

	jsonHint := widget.NewLabel(locale.T("wizard.source.json_hint"))
	jsonHint.Wrapping = fyne.TextWrapWord
	jsonCol := container.NewVBox(jsonHint, jsonScroll)

	settingsTab := container.NewTabItem(locale.T("wizard.source.tab_settings"), settingsScroll)
	previewTab := container.NewTabItem(locale.T("wizard.source.tab_preview"), previewBox)
	jsonTab := container.NewTabItem(locale.T("wizard.source.tab_json"), jsonCol)
	tabs := container.NewAppTabs(settingsTab, previewTab, jsonTab)
	afterSync = func() {
		if tabs.Selected() == previewTab {
			refreshPreviewTab()
		}
		if tabs.Selected() == jsonTab {
			refreshJSONTab()
		}
	}
	tabs.OnSelected = func(ti *container.TabItem) {
		switch ti {
		case previewTab:
			refreshPreviewTab()
		case jsonTab:
			refreshJSONTab()
		}
	}

	cancelBtn := widget.NewButton(locale.T("wizard.outbound.button_cancel"), func() {
		win.Close()
	})
	saveBtn := widget.NewButton(locale.T("wizard.outbound.button_save"), func() {
		if err := serializeParserAfterSourceEdit(presenter, guiState, presenter.Model(), win); err != nil {
			return
		}
		win.Close()
	})
	buttonsRow := container.NewHBox(layout.NewSpacer(), cancelBtn, saveBtn)
	root := container.NewBorder(nil, buttonsRow, nil, nil, tabs)

	win.SetContent(root)
	win.Resize(fyne.NewSize(440, 420))
	win.CenterOnScreen()
	syncFormFromModel()
	win.Show()
	presenter.UpdateChildOverlay()
}
