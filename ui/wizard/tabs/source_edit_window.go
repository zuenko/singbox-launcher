package tabs

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
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

func serializeParserAfterSourceEdit(
	presenter *wizardpresentation.WizardPresenter,
	guiState *wizardpresentation.GUIState,
	m *wizardmodels.WizardModel,
) error {
	serialized, err := wizardbusiness.SerializeParserConfig(m.ParserConfig)
	if err != nil {
		debuglog.ErrorLog("source_edit: SerializeParserConfig: %v", err)
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

// showSourceEditWindow opens Settings | Preview for one proxy source (SPEC 026).
func showSourceEditWindow(
	presenter *wizardpresentation.WizardPresenter,
	guiState *wizardpresentation.GUIState,
	parent fyne.Window,
	sourceIndex int,
	shortLabel string,
) {
	if presenter != nil {
		if w := presenter.OpenViewWindow(); w != nil {
			w.RequestFocus()
			return
		}
	}
	app := fyne.CurrentApp()
	if app == nil {
		return
	}
	m := presenter.Model()
	if m == nil || m.ParserConfig == nil || sourceIndex < 0 || sourceIndex >= len(m.ParserConfig.ParserConfig.Proxies) {
		return
	}

	title := locale.Tf("wizard.source.edit_title", shortLabel)
	title = wizardutils.TruncateStringEllipsis(title, wizardutils.MaxLabelRunes, "...")
	win := app.NewWindow(title)
	if presenter != nil {
		presenter.SetViewWindow(win)
		win.SetOnClosed(func() {
			presenter.ClearViewWindow()
			presenter.UpdateChildOverlay()
		})
	}

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

	var exposeOnChanged func(bool)
	exposeOnChanged = func(v bool) {
		if !exposeCheck.Enabled() {
			return
		}
		pp := proxyRef()
		if pp == nil {
			return
		}
		pp.ExposeGroupTagsToGlobal = v
		_ = serializeParserAfterSourceEdit(presenter, guiState, presenter.Model())
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
		if tb, ok := interface{}(exposeCheck).(interface{ SetToolTip(string) }); ok {
			if has {
				tb.SetToolTip("")
			} else {
				tb.SetToolTip(locale.T("wizard.source.expose_tags_tooltip"))
			}
		}
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
	}

	prefixEntry.OnChanged = func(s string) {
		p := proxyRef()
		if p == nil {
			return
		}
		p.TagPrefix = strings.TrimSpace(s)
		wizardbusiness.RenameWizardLocalOutboundTags(p, sourceIndex)
		_ = serializeParserAfterSourceEdit(presenter, guiState, presenter.Model())
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
				dialog.ShowError(fmt.Errorf("%s", locale.T("wizard.source.wizard_tag_conflict")), win)
				return
			}
		} else {
			wizardbusiness.RemoveWizardSelectOutbounds(p)
			wizardbusiness.RemoveWizardAutoOutbounds(p)
			wizardbusiness.SyncExposeFlagWhenNoLocalGroups(p)
		}
		_ = serializeParserAfterSourceEdit(presenter, guiState, presenter.Model())
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
				dialog.ShowError(fmt.Errorf("%s", locale.T("wizard.source.wizard_tag_conflict")), win)
				return
			}
		} else {
			wizardbusiness.RemoveWizardSelectOutbounds(p)
			wizardbusiness.SyncExposeFlagWhenNoLocalGroups(p)
		}
		_ = serializeParserAfterSourceEdit(presenter, guiState, presenter.Model())
		syncFormFromModel()
	}

	excludeCheck.OnChanged = func(v bool) {
		p := proxyRef()
		if p == nil {
			return
		}
		p.ExcludeFromGlobal = v
		_ = serializeParserAfterSourceEdit(presenter, guiState, presenter.Model())
		refreshExcludeHint()
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
	settingsScroll.SetMinSize(fyne.NewSize(360, 280))

	previewStatus := widget.NewLabel(locale.T("wizard.source.preview_loading"))
	previewListHost := container.NewMax()
	previewBox := container.NewBorder(previewStatus, nil, nil, nil, previewListHost)

	refreshPreviewTab := func() {
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
				previewListHost.Objects = nil
				if err != nil {
					previewStatus.SetText(locale.Tf("wizard.source.preview_error", err.Error()))
					previewListHost.Add(layout.NewSpacer())
					previewListHost.Refresh()
					return
				}
				if len(nodes) == 0 {
					previewStatus.SetText(locale.T("wizard.source.view_no_servers"))
					previewListHost.Add(layout.NewSpacer())
					previewListHost.Refresh()
					return
				}
				previewStatus.SetText(locale.Tf("wizard.source.view_server_count", len(nodes)))
				list := widget.NewList(
					func() int { return len(nodes) },
					func() fyne.CanvasObject { return widget.NewLabel("") },
					func(id int, o fyne.CanvasObject) {
						o.(*widget.Label).SetText(nodeDisplayLine(nodes[id]))
					},
				)
				sc := container.NewScroll(list)
				sc.SetMinSize(fyne.NewSize(0, 240))
				previewListHost.Add(sc)
				previewListHost.Refresh()
			})
		}()
	}

	settingsTab := container.NewTabItem(locale.T("wizard.source.tab_settings"), settingsScroll)
	previewTab := container.NewTabItem(locale.T("wizard.source.tab_preview"), previewBox)
	tabs := container.NewAppTabs(settingsTab, previewTab)
	tabs.OnSelected = func(ti *container.TabItem) {
		if ti == previewTab {
			refreshPreviewTab()
		}
	}

	closeBtn := widget.NewButton(locale.T("wizard.source.edit_close"), func() { win.Close() })
	root := container.NewBorder(nil, container.NewHBox(layout.NewSpacer(), closeBtn), nil, nil, tabs)

	win.SetContent(root)
	win.Resize(fyne.NewSize(440, 420))
	win.CenterOnScreen()
	syncFormFromModel()
	win.Show()
	if presenter != nil {
		presenter.UpdateChildOverlay()
	}
}
