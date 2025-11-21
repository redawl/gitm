package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// AnalysisToolbar contains the top-level toolbar for gitm
type AnalysisToolbar struct {
	widget.BaseWidget
	// IsRecording specified whether to record packets
	IsRecording                 bool
	record, stop, decodeHistory *ToolbarButton
}

// NewAnalysisToolbar creates a new RecordButton
// TODO: Reduce the number of parameters. Maybe take OnTapped as an argument?
func NewAnalysisToolbar(packetFilter *PacketFilter, w fyne.Window, decodeHistoryList *widget.List, parentContainer **container.Split) *AnalysisToolbar {
	tb := &AnalysisToolbar{
		record: &ToolbarButton{
			Button: widget.Button{
				Text: lang.L("Record"),
				Icon: theme.Icon(theme.IconNameMediaPlay),
			},
		},
		stop: &ToolbarButton{
			Button: widget.Button{
				Text:       lang.L("Stop"),
				Importance: widget.DangerImportance,
				Icon:       theme.Icon(theme.IconNameMediaStop),
			},
		},
		decodeHistory: &ToolbarButton{
			Button: widget.Button{
				Text: lang.L("Decode History"),
				Icon: theme.Icon(theme.IconNameNavigateBack),
			},
		},
	}

	tb.decodeHistory.OnTapped = func() {
		if decodeHistoryList.Hidden {
			decodeHistoryList.Show()
			tb.decodeHistory.SetIcon(theme.Icon(theme.IconNameNavigateNext))
		} else {
			decodeHistoryList.Hide()
			tb.decodeHistory.SetIcon(theme.Icon(theme.IconNameNavigateBack))
		}
		decodeHistoryList.Refresh()
		(*parentContainer).Refresh()
	}

	tb.record.Enable()
	tb.stop.Disable()

	tb.record.OnTapped = func() {
		if len(packetFilter.Packets) > 0 {
			dialog.ShowConfirm(
				lang.L("Overwrite Packets"),
				lang.L("Starting a new capture will overwrite existing packets. Are you sure?"),
				func(b bool) {
					if b {
						packetFilter.ClearPackets()
						tb.startRecording()
					}
				},
				w)
		} else {
			tb.startRecording()
		}
	}

	tb.stop.OnTapped = tb.stopRecording

	tb.ExtendBaseWidget(tb)

	return tb
}

func (tb *AnalysisToolbar) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(
		widget.NewToolbar(
			tb.record, tb.stop, widget.NewToolbarSpacer(), tb.decodeHistory,
		),
	)
}

func (tb *AnalysisToolbar) startRecording() {
	tb.IsRecording = true
	tb.record.Disable()
	tb.stop.Enable()
}

func (tb *AnalysisToolbar) stopRecording() {
	tb.IsRecording = false
	tb.record.Enable()
	tb.stop.Disable()
}

var _ widget.ToolbarItem = (*ToolbarButton)(nil)

type ToolbarButton struct {
	widget.Button
}

// ToolbarObject implements widget.ToolbarItem.
func (t *ToolbarButton) ToolbarObject() fyne.CanvasObject {
	return &t.Button
}
