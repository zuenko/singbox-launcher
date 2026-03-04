// Package outbounds_configurator provides reusable UI for configuring outbounds in the wizard:
// list of all outbounds (global + per-source), Edit/Delete/Add, and helpers to apply changes back to ParserConfig.
package outbounds_configurator

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"singbox-launcher/core/config"
)

// outboundRow identifies one outbound in the list (global or per-source).
type outboundRow struct {
	IsGlobal     bool
	SourceIndex  int
	IndexInSlice int
	Outbound     *config.OutboundConfig
	SourceLabel  string
}

// collectRows builds the flat list: local outbounds first (per source), then global.
// Order matters: lower items can reference upper items (e.g. in addOutbounds), not the other way around.
func collectRows(pc *config.ParserConfig) []outboundRow {
	var rows []outboundRow
	for si, proxy := range pc.ParserConfig.Proxies {
		label := proxy.Source
		if label == "" {
			label = "Source " + strconv.Itoa(si+1)
		}
		if len(label) > 40 {
			label = label[:37] + "..."
		}
		for i := range proxy.Outbounds {
			rows = append(rows, outboundRow{
				IsGlobal:     false,
				SourceIndex:  si,
				IndexInSlice: i,
				Outbound:     &pc.ParserConfig.Proxies[si].Outbounds[i],
				SourceLabel:  label,
			})
		}
	}
	for i := range pc.ParserConfig.Outbounds {
		rows = append(rows, outboundRow{IsGlobal: true, IndexInSlice: i, Outbound: &pc.ParserConfig.Outbounds[i], SourceLabel: "Global"})
	}
	return rows
}

// collectAllTags returns all outbound tags in display order (local first, then global).
func collectAllTags(pc *config.ParserConfig) []string {
	var tags []string
	for si := range pc.ParserConfig.Proxies {
		for i := range pc.ParserConfig.Proxies[si].Outbounds {
			tags = append(tags, pc.ParserConfig.Proxies[si].Outbounds[i].Tag)
		}
	}
	for i := range pc.ParserConfig.Outbounds {
		tags = append(tags, pc.ParserConfig.Outbounds[i].Tag)
	}
	return tags
}

// tagsAbove returns tags of rows that appear before rowIndex (only those can be used in addOutbounds).
func tagsAbove(rows []outboundRow, rowIndex int) []string {
	if rowIndex <= 0 {
		return nil
	}
	tags := make([]string, 0, rowIndex)
	for i := 0; i < rowIndex; i++ {
		tags = append(tags, rows[i].Outbound.Tag)
	}
	return tags
}

// sameScope returns true if both rows are in the same scope (same source or both global).
func sameScope(a, b outboundRow) bool {
	if a.IsGlobal && b.IsGlobal {
		return true
	}
	return !a.IsGlobal && !b.IsGlobal && a.SourceIndex == b.SourceIndex
}

// moveOutboundUp swaps the outbound with the previous one in the same scope.
func moveOutboundUp(parserConfig *config.ParserConfig, r outboundRow) {
	if r.IsGlobal {
		if r.IndexInSlice <= 0 {
			return
		}
		s := parserConfig.ParserConfig.Outbounds
		s[r.IndexInSlice], s[r.IndexInSlice-1] = s[r.IndexInSlice-1], s[r.IndexInSlice]
	} else {
		prox := &parserConfig.ParserConfig.Proxies[r.SourceIndex]
		if r.IndexInSlice <= 0 {
			return
		}
		prox.Outbounds[r.IndexInSlice], prox.Outbounds[r.IndexInSlice-1] = prox.Outbounds[r.IndexInSlice-1], prox.Outbounds[r.IndexInSlice]
	}
}

// moveOutboundDown swaps the outbound with the next one in the same scope.
func moveOutboundDown(parserConfig *config.ParserConfig, r outboundRow) {
	if r.IsGlobal {
		s := parserConfig.ParserConfig.Outbounds
		if r.IndexInSlice >= len(s)-1 {
			return
		}
		s[r.IndexInSlice], s[r.IndexInSlice+1] = s[r.IndexInSlice+1], s[r.IndexInSlice]
	} else {
		prox := &parserConfig.ParserConfig.Proxies[r.SourceIndex]
		if r.IndexInSlice >= len(prox.Outbounds)-1 {
			return
		}
		prox.Outbounds[r.IndexInSlice], prox.Outbounds[r.IndexInSlice+1] = prox.Outbounds[r.IndexInSlice+1], prox.Outbounds[r.IndexInSlice]
	}
}

// NewConfiguratorContent builds a reusable outbounds configurator content for embedding into tabs.
// parserConfig is modified in place; callers are responsible for serializing and syncing it back.
func NewConfiguratorContent(parent fyne.Window, parserConfig *config.ParserConfig) fyne.CanvasObject {
	listContent := container.NewVBox()

	var refreshList func()
	refreshList = func() {
		rows := collectRows(parserConfig)
		items := make([]fyne.CanvasObject, 0, len(rows))
		for rowIdx, r := range rows {
			r := r
			rowIdx := rowIdx
			label := r.Outbound.Tag + " (" + r.Outbound.Type + ") — " + r.SourceLabel
			canUp := rowIdx > 0 && sameScope(rows[rowIdx], rows[rowIdx-1])
			canDown := rowIdx < len(rows)-1 && sameScope(rows[rowIdx], rows[rowIdx+1])

			upBtn := widget.NewButtonWithIcon("", theme.MoveUpIcon(), func() {
				rowsNow := collectRows(parserConfig)
				idx := -1
				for i := range rowsNow {
					if rowsNow[i].Outbound == r.Outbound {
						idx = i
						break
					}
				}
				if idx <= 0 || !sameScope(rowsNow[idx], rowsNow[idx-1]) {
					return
				}
				moveOutboundUp(parserConfig, rowsNow[idx])
				refreshList()
			})
			if !canUp {
				upBtn.Disable()
			}

			downBtn := widget.NewButtonWithIcon("", theme.MoveDownIcon(), func() {
				rowsNow := collectRows(parserConfig)
				idx := -1
				for i := range rowsNow {
					if rowsNow[i].Outbound == r.Outbound {
						idx = i
						break
					}
				}
				if idx < 0 || idx >= len(rowsNow)-1 || !sameScope(rowsNow[idx], rowsNow[idx+1]) {
					return
				}
				moveOutboundDown(parserConfig, rowsNow[idx])
				refreshList()
			})
			if !canDown {
				downBtn.Disable()
			}

			editBtn := widget.NewButton("Edit", func() {
				rowsNow := collectRows(parserConfig)
				idx := -1
				for i := range rowsNow {
					if rowsNow[i].Outbound == r.Outbound {
						idx = i
						break
					}
				}
				tagsForAdd := tagsAbove(rowsNow, idx)
				ShowEditDialog(parent, parserConfig, r.Outbound, r.IsGlobal, r.SourceIndex, tagsForAdd, func(updated *config.OutboundConfig, scopeKind string, sourceIndex int) {
					*r.Outbound = *updated
					refreshList()
				})
			})

			delBtn := widget.NewButton("Delete", func() {
				rowsNow := collectRows(parserConfig)
				idx := -1
				for i := range rowsNow {
					if rowsNow[i].Outbound == r.Outbound {
						idx = i
						break
					}
				}
				if idx < 0 {
					return
				}
				r2 := rowsNow[idx]
				if r2.IsGlobal {
					pc := parserConfig
					pc.ParserConfig.Outbounds = append(pc.ParserConfig.Outbounds[:r2.IndexInSlice], pc.ParserConfig.Outbounds[r2.IndexInSlice+1:]...)
				} else {
					prox := &parserConfig.ParserConfig.Proxies[r2.SourceIndex]
					prox.Outbounds = append(prox.Outbounds[:r2.IndexInSlice], prox.Outbounds[r2.IndexInSlice+1:]...)
				}
				refreshList()
			})

			row := container.NewHBox(upBtn, downBtn, widget.NewLabel(label), layout.NewSpacer(), editBtn, delBtn)
			items = append(items, row)
		}
		listContent.Objects = items
		listContent.Refresh()
	}

	refreshList()

	addBtn := widget.NewButton("Add", func() {
		existingTags := collectAllTags(parserConfig)
		ShowEditDialog(parent, parserConfig, nil, true, -1, existingTags, func(updated *config.OutboundConfig, scopeKind string, sourceIndex int) {
			if scopeKind == "global" || sourceIndex < 0 {
				parserConfig.ParserConfig.Outbounds = append(parserConfig.ParserConfig.Outbounds, *updated)
			} else {
				for sourceIndex >= len(parserConfig.ParserConfig.Proxies) {
					parserConfig.ParserConfig.Proxies = append(parserConfig.ParserConfig.Proxies, config.ProxySource{})
				}
				parserConfig.ParserConfig.Proxies[sourceIndex].Outbounds = append(parserConfig.ParserConfig.Proxies[sourceIndex].Outbounds, *updated)
			}
			refreshList()
		})
	})

	scroll := container.NewScroll(listContent)
	scroll.SetMinSize(fyne.NewSize(0, 280))

	top := container.NewBorder(nil, nil, nil, addBtn, widget.NewLabel("Outbounds (local per source first, then global; lower can use upper in addOutbounds):"))
	return container.NewBorder(
		top,
		nil,
		nil, nil,
		scroll,
	)
}
