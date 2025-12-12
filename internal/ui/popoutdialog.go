package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/redawl/gitm/internal/util"
)

type PopoutDialog struct {
	dialog.CustomDialog
}

func NewPopoutDialog(title string, dismiss string, creator func() fyne.CanvasObject, parent fyne.Window) dialog.Dialog {
	d := &PopoutDialog{}
	d.CustomDialog = *dialog.NewCustom(title, dismiss, creator(), parent)
	d.SetButtons([]fyne.CanvasObject{
		&widget.Button{
			Text:     dismiss,
			Icon:     theme.ContentClearIcon(),
			OnTapped: d.Dismiss,
		},
		&widget.Button{
			Text: lang.L("Popout"),
			Icon: theme.UploadIcon(),
			OnTapped: func() {
				w := util.NewWindowIfNotExists(title)
				w.SetContent(
					container.NewBorder(
						widget.NewToolbar(
							widget.NewToolbarSpacer(),
							widget.NewToolbarAction(theme.DownloadIcon(), func() {
								w.Hide()
								d.Show()
							}),
						),
						nil, nil, nil,
						creator(),
					),
				)
				d.Dismiss()
				w.Show()
			},
		},
	})

	return d
}
