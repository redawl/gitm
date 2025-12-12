package ui

import (
	_ "embed"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

//go:embed assets/encrypted.svg
var encryptedIcon []byte

//go:embed assets/not_encrypted.svg
var notEncryptedIcon []byte

var encryptedIconRes = &fyne.StaticResource{
	StaticName:    "encrypted.svg",
	StaticContent: encryptedIcon,
}

var notEncryptedIconRes = &fyne.StaticResource{
	StaticName:    "not_encrypted.svg",
	StaticContent: notEncryptedIcon,
}

var (
	iconNameEncrypted    = theme.NewThemedResource(encryptedIconRes)
	iconNameNotEncrypted = theme.NewThemedResource(notEncryptedIconRes)
)

func EncryptedIcon() fyne.Resource    { return iconNameEncrypted }
func NotEncryptedIcon() fyne.Resource { return iconNameNotEncrypted }
