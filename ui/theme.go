// Package ui provides the graphical user interface for the clipboard manager.
package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// Dracula color palette - https://draculatheme.com/
// Modified with darker background for user preference
var (
	// Base colors - Extra dark
	draculaBackground = color.NRGBA{R: 18, G: 18, B: 28, A: 255}    // #12121c - very dark
	draculaCurrent    = color.NRGBA{R: 30, G: 30, B: 46, A: 255}    // #1e1e2e - dark surface
	draculaSelection  = color.NRGBA{R: 40, G: 40, B: 60, A: 255}    // #28283c - slightly lighter
	draculaComment    = color.NRGBA{R: 98, G: 114, B: 164, A: 255}  // #6272a4

	// Text colors
	draculaForeground = color.NRGBA{R: 248, G: 248, B: 242, A: 255} // #f8f8f2

	// Accent colors
	draculaCyan   = color.NRGBA{R: 139, G: 233, B: 253, A: 255} // #8be9fd
	draculaGreen  = color.NRGBA{R: 80, G: 250, B: 123, A: 255}  // #50fa7b
	draculaOrange = color.NRGBA{R: 255, G: 184, B: 108, A: 255} // #ffb86c
	draculaPink   = color.NRGBA{R: 255, G: 121, B: 198, A: 255} // #ff79c6
	draculaPurple = color.NRGBA{R: 189, G: 147, B: 249, A: 255} // #bd93f9
	draculaRed    = color.NRGBA{R: 255, G: 85, B: 85, A: 255}   // #ff5555
	draculaYellow = color.NRGBA{R: 241, G: 250, B: 140, A: 255} // #f1fa8c
)

// DraculaTheme implements fyne.Theme with Dracula colors
type DraculaTheme struct{}

// Keep CatppuccinTheme as alias for backwards compatibility
type CatppuccinTheme = DraculaTheme

var _ fyne.Theme = (*DraculaTheme)(nil)

func (d *DraculaTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return draculaBackground
	case theme.ColorNameButton:
		return draculaCurrent
	case theme.ColorNameDisabled:
		return draculaComment
	case theme.ColorNameDisabledButton:
		return draculaCurrent
	case theme.ColorNameError:
		return draculaRed
	case theme.ColorNameFocus:
		return draculaPurple
	case theme.ColorNameForeground:
		return draculaForeground
	case theme.ColorNameForegroundOnError:
		return draculaBackground
	case theme.ColorNameForegroundOnPrimary:
		return draculaBackground
	case theme.ColorNameForegroundOnSuccess:
		return draculaBackground
	case theme.ColorNameForegroundOnWarning:
		return draculaBackground
	case theme.ColorNameHeaderBackground:
		return draculaCurrent
	case theme.ColorNameHover:
		return color.NRGBA{R: 50, G: 50, B: 75, A: 255} // Slightly brighter on hover
	case theme.ColorNameHyperlink:
		return draculaCyan
	case theme.ColorNameInputBackground:
		return draculaCurrent
	case theme.ColorNameInputBorder:
		return draculaComment
	case theme.ColorNameMenuBackground:
		return draculaBackground
	case theme.ColorNameOverlayBackground:
		return draculaBackground
	case theme.ColorNamePlaceHolder:
		return draculaComment
	case theme.ColorNamePressed:
		return draculaSelection
	case theme.ColorNamePrimary:
		return draculaPurple
	case theme.ColorNameScrollBar:
		return draculaComment
	case theme.ColorNameSelection:
		return color.NRGBA{R: 189, G: 147, B: 249, A: 120} // Purple with transparency - main highlight
	case theme.ColorNameSeparator:
		return draculaCurrent
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 120}
	case theme.ColorNameSuccess:
		return draculaGreen
	case theme.ColorNameWarning:
		return draculaOrange
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (d *DraculaTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (d *DraculaTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (d *DraculaTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 14
	case theme.SizeNamePadding:
		return 6
	case theme.SizeNameInnerPadding:
		return 4
	case theme.SizeNameScrollBar:
		return 10
	case theme.SizeNameScrollBarSmall:
		return 4
	case theme.SizeNameLineSpacing:
		return 4
	default:
		return theme.DefaultTheme().Size(name)
	}
}
