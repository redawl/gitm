package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type PlaceHolder struct {
	widget.BaseWidget
	label     *widget.Label
	icon      *widget.Icon
	container *fyne.Container
}

func NewPlaceHolder(label string, icon fyne.Resource) *PlaceHolder {
	placeHolder := &PlaceHolder{
		label: &widget.Label{
			Text:       label,
			Importance: widget.LowImportance,
			SizeName:   theme.SizeNameSubHeadingText,
			Alignment:  fyne.TextAlignCenter,
		},
		icon: &widget.Icon{
			Resource: theme.NewDisabledResource(icon),
		},
	}

	placeHolder.container = container.NewCenter(
		container.NewVBox(
			placeHolder.icon,
			placeHolder.label,
		),
	)

	placeHolder.ExtendBaseWidget(placeHolder)

	return placeHolder
}

func (p *PlaceHolder) MinSize() fyne.Size {
	return p.container.MinSize()
}

func (p *PlaceHolder) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(p.container)
}
