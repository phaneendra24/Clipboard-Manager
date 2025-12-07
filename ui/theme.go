// Package ui provides the graphical user interface for the clipboard manager.
package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// Catppuccin Mocha color palette
var (
	// Base colors
	catBase     = color.NRGBA{R: 30, G: 30, B: 46, A: 255}   // #1e1e2e
	catMantle   = color.NRGBA{R: 24, G: 24, B: 37, A: 255}   // #181825
	catCrust    = color.NRGBA{R: 17, G: 17, B: 27, A: 255}   // #11111b
	catSurface0 = color.NRGBA{R: 49, G: 50, B: 68, A: 255}   // #313244
	catSurface1 = color.NRGBA{R: 69, G: 71, B: 90, A: 255}   // #45475a
	catSurface2 = color.NRGBA{R: 88, G: 91, B: 112, A: 255}  // #585b70

	// Text colors
	catText     = color.NRGBA{R: 205, G: 214, B: 244, A: 255} // #cdd6f4
	catSubtext1 = color.NRGBA{R: 186, G: 194, B: 222, A: 255} // #bac2de
	catSubtext0 = color.NRGBA{R: 166, G: 173, B: 200, A: 255} // #a6adc8
	catOverlay2 = color.NRGBA{R: 147, G: 153, B: 178, A: 255} // #9399b2

	// Accent colors
	catLavender = color.NRGBA{R: 180, G: 190, B: 254, A: 255} // #b4befe
	catBlue     = color.NRGBA{R: 137, G: 180, B: 250, A: 255} // #89b4fa
	catSapphire = color.NRGBA{R: 116, G: 199, B: 236, A: 255} // #74c7ec
	catSky      = color.NRGBA{R: 137, G: 220, B: 235, A: 255} // #89dceb
	catTeal     = color.NRGBA{R: 148, G: 226, B: 213, A: 255} // #94e2d5
	catGreen    = color.NRGBA{R: 166, G: 227, B: 161, A: 255} // #a6e3a1
	catYellow   = color.NRGBA{R: 249, G: 226, B: 175, A: 255} // #f9e2af
	catPeach    = color.NRGBA{R: 250, G: 179, B: 135, A: 255} // #fab387
	catMaroon   = color.NRGBA{R: 235, G: 160, B: 172, A: 255} // #eba0ac
	catRed      = color.NRGBA{R: 243, G: 139, B: 168, A: 255} // #f38ba8
	catMauve    = color.NRGBA{R: 203, G: 166, B: 247, A: 255} // #cba6f7
	catPink     = color.NRGBA{R: 245, G: 194, B: 231, A: 255} // #f5c2e7
	catFlamingo = color.NRGBA{R: 242, G: 205, B: 205, A: 255} // #f2cdcd
	catRosewater= color.NRGBA{R: 245, G: 224, B: 220, A: 255} // #f5e0dc
)

// CatppuccinTheme implements fyne.Theme with Catppuccin Mocha colors
type CatppuccinTheme struct{}

var _ fyne.Theme = (*CatppuccinTheme)(nil)

func (c *CatppuccinTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return catBase
	case theme.ColorNameButton:
		return catSurface0
	case theme.ColorNameDisabled:
		return catSurface1
	case theme.ColorNameDisabledButton:
		return catSurface0
	case theme.ColorNameError:
		return catRed
	case theme.ColorNameFocus:
		return catLavender
	case theme.ColorNameForeground:
		return catText
	case theme.ColorNameForegroundOnError:
		return catBase
	case theme.ColorNameForegroundOnPrimary:
		return catBase
	case theme.ColorNameForegroundOnSuccess:
		return catBase
	case theme.ColorNameForegroundOnWarning:
		return catBase
	case theme.ColorNameHeaderBackground:
		return catMantle
	case theme.ColorNameHover:
		return catSurface1
	case theme.ColorNameHyperlink:
		return catBlue
	case theme.ColorNameInputBackground:
		return catSurface0
	case theme.ColorNameInputBorder:
		return catSurface2
	case theme.ColorNameMenuBackground:
		return catMantle
	case theme.ColorNameOverlayBackground:
		return catMantle
	case theme.ColorNamePlaceHolder:
		return catOverlay2
	case theme.ColorNamePressed:
		return catSurface2
	case theme.ColorNamePrimary:
		return catMauve
	case theme.ColorNameScrollBar:
		return catSurface2
	case theme.ColorNameSelection:
		return catSurface1
	case theme.ColorNameSeparator:
		return catSurface0
	case theme.ColorNameShadow:
		return catCrust
	case theme.ColorNameSuccess:
		return catGreen
	case theme.ColorNameWarning:
		return catYellow
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (c *CatppuccinTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (c *CatppuccinTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (c *CatppuccinTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 14
	case theme.SizeNamePadding:
		return 6
	case theme.SizeNameInnerPadding:
		return 4
	case theme.SizeNameScrollBar:
		return 12
	case theme.SizeNameScrollBarSmall:
		return 4
	default:
		return theme.DefaultTheme().Size(name)
	}
}
