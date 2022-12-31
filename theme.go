package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type new_theme struct{}

var _ fyne.Theme = (*new_theme)(nil)

func (*new_theme) Font(style fyne.TextStyle) fyne.Resource {
	return resourceJfOpenhuninn11Ttf
}

func (*new_theme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	return theme.DefaultTheme().Color(n, v)
}

func (*new_theme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(n)
}

func (*new_theme) Size(n fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(n)
}
