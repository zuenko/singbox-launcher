// Package presentation содержит слой представления визарда конфигурации.
//
// Файл presenter_sync.go содержит методы синхронизации данных между моделью и GUI:
//   - SyncModelToGUI - обновляет виджеты GUI из модели данных (SourceURLs, ParserConfigJSON, SelectedFinalOutbound)
//   - SyncGUIToModel - обновляет модель данных из виджетов GUI (обратная синхронизация)
//
// Эти методы обеспечивают двустороннюю синхронизацию между WizardModel и GUIState,
// что является ключевой частью архитектуры MVP.
//
// Синхронизация данных - это отдельная ответственность от других методов презентера.
// Методы синхронизации используются в разных местах (перед сохранением, при инициализации).
//
// Используется в:
//   - wizard.go - SyncModelToGUI вызывается при инициализации визарда для установки начальных значений
//   - presenter_save.go - SyncGUIToModel вызывается перед сохранением для получения актуальных данных
//   - presenter_async.go - SyncGUIToModel вызывается перед парсингом для получения актуальных данных
package presentation

import (
	"encoding/json"
	"fmt"
	"strings"

	"fyne.io/fyne/v2/dialog"

	"singbox-launcher/core/config"
	"singbox-launcher/internal/locale"
	wizardbusiness "singbox-launcher/ui/wizard/business"
	wizardmodels "singbox-launcher/ui/wizard/models"
)

// SyncModelToGUI синхронизирует данные из модели в GUI.
// SourceURLEntry shows SourceURLs (input for Add only); source list is from ParserConfig.Proxies in refreshSourcesList.
//
// Всегда через fyne.Do: прямой вызов SetText/Refresh с потока, где создали окно, даёт зависания/краши при открытии визарда.
// Если сразу после SyncModelToGUI вызывается SyncGUIToModel (Save), DNS-виджеты могут быть ещё пустыми — см. защиту в SyncGUIToModel.
func (p *WizardPresenter) SyncModelToGUI() {
	p.UpdateUI(p.applyWizardWidgetsFromModel)

	// Пересоздаем вкладку Rules, если она уже создана (часть синхронизации модели с GUI)
	// Это обновит чекбоксы и селекторы правил в соответствии с текущим состоянием модели
	if p.createRulesTabFunc != nil && p.guiState.Tabs != nil {
		// Проверяем, существует ли вкладка Rules
		for _, tabItem := range p.guiState.Tabs.Items {
			if tabItem.Text == "Rules" {
				p.RefreshRulesTabAfterLoadState()
				break
			}
		}
	}
}

func (p *WizardPresenter) applyWizardWidgetsFromModel() {
	if p.guiState == nil {
		return
	}
	if p.guiState.SourceURLEntry != nil {
		p.guiState.SourceURLEntry.SetText(p.model.SourceURLs)
	}
	if p.guiState.ParserConfigEntry != nil {
		p.guiState.ParserConfigEntry.SetText(p.model.ParserConfigJSON)
		p.guiState.LastValidParserConfigJSON = p.model.ParserConfigJSON
	}
	if p.guiState.RefreshSourcesList != nil {
		p.guiState.RefreshSourcesList()
	}
	if p.guiState.RefreshDNSList != nil {
		p.guiState.RefreshDNSList()
	}
	if p.guiState.DNSRulesEntry != nil {
		p.guiState.DNSRulesEntry.SetText(p.model.DNSRulesText)
	}
	p.refreshDNSSelectsFromModel()
	if p.guiState.FinalOutboundSelect != nil {
		options := wizardbusiness.EnsureDefaultAvailableOutbounds(wizardbusiness.GetAvailableOutbounds(p.model))
		p.guiState.FinalOutboundSelect.Options = options
		p.guiState.FinalOutboundSelect.SetSelected(p.model.SelectedFinalOutbound)
		p.guiState.FinalOutboundSelect.Refresh()
	}
}

func (p *WizardPresenter) refreshDNSSelectsFromModel() {
	if p.guiState != nil {
		p.guiState.DNSSelectsProgrammatic = true
		defer func() { p.guiState.DNSSelectsProgrammatic = false }()
	}
	tags := wizardbusiness.DNSEnabledTagOptions(p.model)
	if p.guiState.DNSFinalSelect != nil {
		p.guiState.DNSFinalSelect.Options = tags
		sel := p.model.DNSFinal
		if sel != "" && !stringSliceContains(tags, sel) && len(tags) > 0 {
			sel = tags[0]
			p.model.DNSFinal = sel
		}
		if sel != "" {
			p.guiState.DNSFinalSelect.SetSelected(sel)
		} else if len(tags) > 0 {
			p.guiState.DNSFinalSelect.SetSelected(tags[0])
			p.model.DNSFinal = tags[0]
		}
		p.guiState.DNSFinalSelect.Refresh()
	}
	if p.guiState.DNSDefaultResolverSelect != nil {
		notSet := locale.T("wizard.dns.resolver_not_set")
		opts := append([]string{notSet}, tags...)
		p.guiState.DNSDefaultResolverSelect.Options = opts
		if p.model.DefaultDomainResolverUnset || p.model.DefaultDomainResolver == "" {
			p.guiState.DNSDefaultResolverSelect.SetSelected(notSet)
		} else {
			sel := strings.TrimSpace(p.model.DefaultDomainResolver)
			if !stringSliceContains(tags, sel) {
				p.model.DefaultDomainResolver = ""
				p.model.DefaultDomainResolverUnset = true
				p.guiState.DNSDefaultResolverSelect.SetSelected(notSet)
			} else {
				p.guiState.DNSDefaultResolverSelect.SetSelected(sel)
			}
		}
		p.guiState.DNSDefaultResolverSelect.Refresh()
	}
	if p.guiState.DNSStrategySelect != nil {
		def := locale.T("wizard.dns.strategy_default")
		strOpts := []string{def, "ipv4_only", "ipv6_only", "prefer_ipv4", "prefer_ipv6"}
		p.guiState.DNSStrategySelect.Options = strOpts
		if p.model.DNSStrategy == "" {
			p.guiState.DNSStrategySelect.SetSelected(def)
		} else {
			p.guiState.DNSStrategySelect.SetSelected(p.model.DNSStrategy)
		}
		p.guiState.DNSStrategySelect.Refresh()
	}
	if p.guiState.DNSIndependentCacheCheck != nil {
		v := false
		if p.model.DNSIndependentCache != nil {
			v = *p.model.DNSIndependentCache
		}
		p.guiState.DNSIndependentCacheCheck.SetChecked(v)
	}
}

// RefreshDNSDependentSelectsOnly обновляет только селекты Final / resolver / strategy и галочку кэша,
// без пересборки списка строк серверов. Нужен после смены enabled у сервера: полный SyncModelToGUI
// на каждый клик по галочке пересоздаёт все строки и даёт сильные лаги и «зависание» вкладки.
func (p *WizardPresenter) RefreshDNSDependentSelectsOnly() {
	p.UpdateUI(func() {
		if p.guiState == nil {
			return
		}
		p.refreshDNSSelectsFromModel()
	})
}

// RefreshDNSListAndSelects пересобирает список серверов и DNS-селекты (после Add / Edit / Delete).
func (p *WizardPresenter) RefreshDNSListAndSelects() {
	p.UpdateUI(func() {
		if p.guiState == nil {
			return
		}
		if p.guiState.RefreshDNSList != nil {
			p.guiState.RefreshDNSList()
		}
		p.refreshDNSSelectsFromModel()
	})
}

// dnsSelectReadLooksStale: SyncGUIToModel до отработки fyne.Do — Select ещё без выбора, модель уже с тегом.
func dnsSelectReadLooksStale(widgetSelected, modelTag string, model *wizardmodels.WizardModel) bool {
	if strings.TrimSpace(widgetSelected) != "" {
		return false
	}
	mt := strings.TrimSpace(modelTag)
	if mt == "" {
		return false
	}
	return stringSliceContains(wizardbusiness.DNSEnabledTagOptions(model), mt)
}

func stringSliceContains(slice []string, s string) bool {
	for _, x := range slice {
		if x == s {
			return true
		}
	}
	return false
}

// dnsSelectOptionsMissingModelTag: выпадающий список ещё не обновлён под модель (async SyncModelToGUI).
func dnsSelectOptionsMissingModelTag(opts []string, modelTag string) bool {
	mt := strings.TrimSpace(modelTag)
	if mt == "" || len(opts) == 0 {
		return false
	}
	return !stringSliceContains(opts, mt)
}

// SyncGUIToModel синхронизирует данные из GUI в модель.
// Устанавливает флаг изменений, если данные реально изменились.
func (p *WizardPresenter) SyncGUIToModel() {
	changed := false

	if p.guiState.SourceURLEntry != nil {
		newValue := p.guiState.SourceURLEntry.Text
		if p.model.SourceURLs != newValue {
			p.model.SourceURLs = newValue
			changed = true
		}
	}
	if p.guiState.ParserConfigEntry != nil {
		newValue := p.guiState.ParserConfigEntry.Text
		if p.model.ParserConfigJSON != newValue {
			p.model.ParserConfigJSON = newValue
			changed = true
		}
	}
	if p.guiState.FinalOutboundSelect != nil {
		newValue := p.guiState.FinalOutboundSelect.Selected
		if p.model.SelectedFinalOutbound != newValue {
			p.model.SelectedFinalOutbound = newValue
			changed = true
		}
	}
	if p.guiState.DNSRulesEntry != nil {
		newValue := p.guiState.DNSRulesEntry.Text
		// Model→GUI идёт через fyne.Do; до отрисовки кадра Text может быть пустым — не затирать модель.
		if strings.TrimSpace(newValue) == "" && strings.TrimSpace(p.model.DNSRulesText) != "" {
			// keep model
		} else if p.model.DNSRulesText != newValue {
			p.model.DNSRulesText = newValue
			changed = true
		}
	}
	if p.guiState.DNSFinalSelect != nil {
		newValue := p.guiState.DNSFinalSelect.Selected
		opts := p.guiState.DNSFinalSelect.Options
		mf := strings.TrimSpace(p.model.DNSFinal)
		if dnsSelectOptionsMissingModelTag(opts, mf) {
			// keep model — список тегов ещё не из той же модели
		} else if dnsSelectReadLooksStale(newValue, mf, p.model) {
			// keep model
		} else if p.model.DNSFinal != newValue {
			p.model.DNSFinal = newValue
			changed = true
		}
	}
	if p.guiState.DNSDefaultResolverSelect != nil {
		notSet := locale.T("wizard.dns.resolver_not_set")
		sel := p.guiState.DNSDefaultResolverSelect.Selected
		opts := p.guiState.DNSDefaultResolverSelect.Options
		mt := strings.TrimSpace(p.model.DefaultDomainResolver)
		if !p.model.DefaultDomainResolverUnset && dnsSelectOptionsMissingModelTag(opts, mt) {
			// keep model — резолвер из state/шаблона ещё не попал в Options
		} else if sel == notSet {
			if p.model.DefaultDomainResolver != "" || !p.model.DefaultDomainResolverUnset {
				p.model.DefaultDomainResolver = ""
				p.model.DefaultDomainResolverUnset = true
				changed = true
			}
		} else {
			if dnsSelectReadLooksStale(sel, mt, p.model) {
				// keep model
			} else if p.model.DefaultDomainResolver != sel || p.model.DefaultDomainResolverUnset {
				p.model.DefaultDomainResolver = sel
				p.model.DefaultDomainResolverUnset = false
				changed = true
			}
		}
	}
	if p.guiState.DNSStrategySelect != nil {
		def := locale.T("wizard.dns.strategy_default")
		sel := p.guiState.DNSStrategySelect.Selected
		newStr := ""
		if sel != def {
			newStr = sel
		}
		if p.model.DNSStrategy != newStr {
			p.model.DNSStrategy = newStr
			changed = true
		}
	}
	if p.guiState.DNSIndependentCacheCheck != nil {
		v := p.guiState.DNSIndependentCacheCheck.Checked
		cur := false
		if p.model.DNSIndependentCache != nil {
			cur = *p.model.DNSIndependentCache
		}
		if cur != v {
			nv := v
			p.model.DNSIndependentCache = &nv
			changed = true
		}
	}

	// Устанавливаем флаг изменений, если данные изменились
	if changed {
		p.MarkAsChanged()
	}
}

// ValidateAndApplyParserConfigFromEntry parses ParserConfig from the entry, validates it,
// and on success updates model and LastValidParserConfigJSON; on error shows dialog and reverts entry.
// Call when leaving the Outbounds and ParserConfig tab so manual JSON edits are applied or reverted.
func (p *WizardPresenter) ValidateAndApplyParserConfigFromEntry() {
	if p.guiState.ParserConfigEntry == nil {
		return
	}
	text := strings.TrimSpace(p.guiState.ParserConfigEntry.Text)
	if text == "" {
		p.model.ParserConfigJSON = ""
		p.model.ParserConfig = nil
		wizardbusiness.InvalidatePreviewCache(p.model)
		p.guiState.LastValidParserConfigJSON = ""
		return
	}
	pc := &config.ParserConfig{}
	if err := json.Unmarshal([]byte(text), pc); err != nil {
		dialog.ShowError(fmt.Errorf("%s: %w", locale.T("wizard.outbounds.error_invalid_json"), err), p.guiState.Window)
		revert := p.guiState.LastValidParserConfigJSON
		p.guiState.ParserConfigUpdating = true
		p.guiState.ParserConfigEntry.SetText(revert)
		p.guiState.ParserConfigUpdating = false
		return
	}
	if err := wizardbusiness.ValidateParserConfig(pc); err != nil {
		dialog.ShowError(fmt.Errorf("%s: %w", locale.T("wizard.outbounds.error_invalid_config"), err), p.guiState.Window)
		revert := p.guiState.LastValidParserConfigJSON
		p.guiState.ParserConfigUpdating = true
		p.guiState.ParserConfigEntry.SetText(revert)
		p.guiState.ParserConfigUpdating = false
		return
	}
	serialized, err := wizardbusiness.SerializeParserConfig(pc)
	if err != nil {
		dialog.ShowError(fmt.Errorf("%s: %w", locale.T("wizard.outbounds.error_serialize"), err), p.guiState.Window)
		return
	}
	p.model.ParserConfig = pc
	p.model.ParserConfigJSON = serialized
	p.guiState.LastValidParserConfigJSON = serialized
	p.UpdateParserConfig(serialized)
	p.RefreshOutboundOptions()
	if p.guiState.RefreshSourcesList != nil {
		p.guiState.RefreshSourcesList()
	}
	p.model.PreviewNeedsParse = true
	wizardbusiness.InvalidatePreviewCache(p.model)
}
