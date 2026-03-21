package fynewidget

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// SecondaryTapWrap is a thin widget around content that receives secondary (right) taps.
// Primary taps are handled by children (e.g. list row labels and buttons).
type SecondaryTapWrap struct {
	widget.BaseWidget

	Content fyne.CanvasObject

	// OnSecondary is invoked on right-click / long-press secondary tap.
	OnSecondary func(*fyne.PointEvent)
}

// NewSecondaryTapWrap wraps inner content for TappedSecondary handling.
func NewSecondaryTapWrap(inner fyne.CanvasObject) *SecondaryTapWrap {
	w := &SecondaryTapWrap{Content: inner}
	w.ExtendBaseWidget(w)
	return w
}

// CreateRenderer implements fyne.Widget.
func (w *SecondaryTapWrap) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.Content)
}

// TappedSecondary implements fyne.SecondaryTappable.
func (w *SecondaryTapWrap) TappedSecondary(pe *fyne.PointEvent) {
	if w.OnSecondary != nil {
		w.OnSecondary(pe)
	}
}
