// Package components содержит переиспользуемые UI-компоненты.
//
// tooltip_wrapper.go — обёртка, показывающая tooltip при наведении мыши.
// Использует overlay поверх кнопки (кнопка не реализует Hoverable, поэтому
// обёртка не получала события). Overlay реализует desktop.Hoverable и Tappable.
package components

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// tooltipOverlay — прозрачный overlay поверх кнопки, получает hover и пересылает tap.
type tooltipOverlay struct {
	widget.BaseWidget
	tooltip string
	canvas  fyne.Canvas
	onTap   func()
	popup   *widget.PopUp
}

func newTooltipOverlay(tooltip string, c fyne.Canvas, onTap func()) *tooltipOverlay {
	o := &tooltipOverlay{tooltip: tooltip, canvas: c, onTap: onTap}
	o.ExtendBaseWidget(o)
	return o
}

func (o *tooltipOverlay) MouseIn(e *desktop.MouseEvent) {
	if o.tooltip == "" || o.canvas == nil || o.popup != nil {
		return
	}
	lbl := widget.NewLabel(o.tooltip)
	o.popup = widget.NewPopUp(lbl, o.canvas)
	pos := e.AbsolutePosition.Add(fyne.NewPos(12, 24))
	o.popup.ShowAtPosition(pos)
}

func (o *tooltipOverlay) MouseMoved(*desktop.MouseEvent) {}

func (o *tooltipOverlay) MouseOut() {
	if o.popup != nil {
		o.popup.Hide()
		o.popup = nil
	}
}

func (o *tooltipOverlay) Tapped(*fyne.PointEvent) {
	if o.onTap != nil {
		o.onTap()
	}
}

func (o *tooltipOverlay) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(canvas.NewRectangle(color.Transparent))
}

// NewToolTipWrapper создаёт контейнер: кнопка + прозрачный overlay поверх.
// Overlay получает hover (tooltip) и tap (передаёт в кнопку).
func NewToolTipWrapper(btn *widget.Button, tooltip string, c fyne.Canvas) fyne.CanvasObject {
	if tooltip == "" || c == nil {
		return btn
	}
	overlay := newTooltipOverlay(tooltip, c, btn.OnTapped)
	return container.NewStack(btn, overlay)
}
