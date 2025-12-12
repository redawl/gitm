package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// HistoryItem
// TODO: Docs
type HistoryItem struct {
	widget.BaseWidget

	button *widget.Button
	label  *widget.Label
}

func NewHistoryItem(text string) *HistoryItem {
	h := &HistoryItem{
		button: widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
			fyne.CurrentApp().Clipboard().SetContent(text)
		}),
		label: widget.NewLabel(text),
	}
	h.label.Truncation = fyne.TextTruncateEllipsis
	h.ExtendBaseWidget(h)

	return h
}

func (h *HistoryItem) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewBorder(nil, nil, nil, h.button, h.label))
}

func (h *HistoryItem) SetText(text string) {
	h.button.OnTapped = func() {
		fyne.CurrentApp().Clipboard().SetContent(text)
	}
	h.label.SetText(text)
	h.Refresh()
}
