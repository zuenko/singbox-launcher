package fynewidget

import (
	"unsafe"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	ttwidget "github.com/dweymouth/fyne-tooltip/widget"
)

// RowHoverGetter returns the [HoverRow] for the current line, or nil (e.g. before the row is assigned).
// Capture the row in a closure: var row *HoverRow; getter := func() *HoverRow { return row }; …; row = NewHoverRow(…).
type RowHoverGetter func() *HoverRow

var (
	_ fyne.Widget       = (*HoverForwardButton)(nil)
	_ desktop.Hoverable = (*HoverForwardButton)(nil)
	_ fyne.Widget       = (*HoverForwardSelect)(nil)
	_ desktop.Hoverable = (*HoverForwardSelect)(nil)
	_ fyne.Widget       = (*HoverForwardTTButton)(nil)
	_ desktop.Hoverable = (*HoverForwardTTButton)(nil)
)

// HoverForwardButton embeds [widget.Button] by value and forwards desktop hover to [HoverRow].
// For ttwidget tooltips on plain actions, prefer [HoverForwardTTButton] or set tooltips on objects that support them.
type HoverForwardButton struct {
	widget.Button
	rowGetter RowHoverGetter
}

// NewHoverForwardButton builds a row-hover-forwarding button (see package doc).
func NewHoverForwardButton(label string, tapped func(), rowGetter RowHoverGetter) *HoverForwardButton {
	b := &HoverForwardButton{rowGetter: rowGetter}
	b.Text = label
	b.OnTapped = tapped
	b.ExtendBaseWidget(b)
	return b
}

// NewHoverForwardButtonWithIcon is like [NewHoverForwardButton] with an icon.
func NewHoverForwardButtonWithIcon(label string, icon fyne.Resource, tapped func(), rowGetter RowHoverGetter) *HoverForwardButton {
	b := &HoverForwardButton{rowGetter: rowGetter}
	b.Text = label
	b.Icon = icon
	b.OnTapped = tapped
	b.ExtendBaseWidget(b)
	return b
}

func (b *HoverForwardButton) MouseIn(e *desktop.MouseEvent) {
	b.Button.MouseIn(e)
	forwardRowHover(b.rowGetter, func(r *HoverRow) { r.MouseIn(e) })
}

func (b *HoverForwardButton) MouseMoved(e *desktop.MouseEvent) {
	b.Button.MouseMoved(e)
	forwardRowHover(b.rowGetter, func(r *HoverRow) { r.MouseMoved(e) })
}

func (b *HoverForwardButton) MouseOut() {
	b.Button.MouseOut()
	forwardRowHover(b.rowGetter, func(r *HoverRow) { r.MouseOut() })
}

// HoverForwardSelect embeds [widget.Select] by value and forwards desktop hover to [HoverRow].
type HoverForwardSelect struct {
	widget.Select
	rowGetter RowHoverGetter
}

// NewHoverForwardSelect builds a row-hover-forwarding select (see package doc).
func NewHoverForwardSelect(options []string, changed func(string), rowGetter RowHoverGetter) *HoverForwardSelect {
	s := &HoverForwardSelect{rowGetter: rowGetter}
	s.Options = options
	s.OnChanged = changed
	s.ExtendBaseWidget(s)
	return s
}

func (s *HoverForwardSelect) MouseIn(e *desktop.MouseEvent) {
	s.Select.MouseIn(e)
	forwardRowHover(s.rowGetter, func(r *HoverRow) { r.MouseIn(e) })
}

func (s *HoverForwardSelect) MouseMoved(e *desktop.MouseEvent) {
	s.Select.MouseMoved(e)
	forwardRowHover(s.rowGetter, func(r *HoverRow) { r.MouseMoved(e) })
}

func (s *HoverForwardSelect) MouseOut() {
	s.Select.MouseOut()
	forwardRowHover(s.rowGetter, func(r *HoverRow) { r.MouseOut() })
}

// HoverForwardTTButton embeds [ttwidget.Button] by value and forwards desktop hover to [HoverRow].
// Do not add a second BaseWidget layer or delegate CreateRenderer from a separate inner button: that
// desynchronizes the renderer from the tree leaf and can shift icon/text on hover refresh.
type HoverForwardTTButton struct {
	ttwidget.Button
	rowGetter RowHoverGetter
}

// NewHoverForwardTTButton builds a tooltip button with row hover forwarding (see package doc).
func NewHoverForwardTTButton(text string, tapped func(), rowGetter RowHoverGetter) *HoverForwardTTButton {
	b := &HoverForwardTTButton{rowGetter: rowGetter}
	b.Text = text
	b.OnTapped = tapped
	b.ExtendBaseWidget(b)
	return b
}

func (b *HoverForwardTTButton) MouseIn(e *desktop.MouseEvent) {
	b.ToolTipWidgetExtend.MouseIn(e)
	b.Button.MouseIn(e)
	forwardRowHover(b.rowGetter, func(r *HoverRow) { r.MouseIn(e) })
}

func (b *HoverForwardTTButton) MouseMoved(e *desktop.MouseEvent) {
	b.ToolTipWidgetExtend.MouseMoved(e)
	b.Button.MouseMoved(e)
	forwardRowHover(b.rowGetter, func(r *HoverRow) { r.MouseMoved(e) })
}

func (b *HoverForwardTTButton) MouseOut() {
	b.ToolTipWidgetExtend.MouseOut()
	b.Button.MouseOut()
	forwardRowHover(b.rowGetter, func(r *HoverRow) { r.MouseOut() })
}

// TTWidget returns the same logical control as *HoverForwardTTButton for APIs that require *ttwidget.Button.
// The outer struct’s leading field is the embedded [ttwidget.Button] value; *HoverForwardTTButton and
// *ttwidget.Button share the same address (unsafe.Pointer conversion for the first field; see unsafe package).
func (b *HoverForwardTTButton) TTWidget() *ttwidget.Button {
	if b == nil {
		return nil
	}
	return (*ttwidget.Button)(unsafe.Pointer(b))
}

func forwardRowHover(getter RowHoverGetter, fn func(*HoverRow)) {
	if getter == nil {
		return
	}
	if r := getter(); r != nil {
		fn(r)
	}
}
