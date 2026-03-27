package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ApplyTheme sets the light theme (transparent backgrounds, dark text).
func ApplyTheme() {
	tview.Styles = tview.Theme{
		PrimitiveBackgroundColor:    tcell.ColorDefault,
		ContrastBackgroundColor:     tcell.ColorDefault,
		MoreContrastBackgroundColor: tcell.ColorDefault,
		BorderColor:                 tcell.ColorDarkGray,
		TitleColor:                  tcell.ColorBlack,
		GraphicsColor:               tcell.ColorDarkGray,
		PrimaryTextColor:            tcell.ColorBlack,
		SecondaryTextColor:          tcell.ColorNavy,
		TertiaryTextColor:           tcell.ColorDarkGreen,
		InverseTextColor:            tcell.ColorBlack,
		ContrastSecondaryTextColor:  tcell.ColorDarkCyan,
	}
}

func selectionStyle() tcell.Style {
	return tcell.StyleDefault.
		Foreground(tcell.ColorNavy).
		Background(tcell.ColorLightCyan).
		Bold(true)
}

func selectionColors() (fg, bg tcell.Color) {
	return tcell.ColorNavy, tcell.ColorLightCyan
}
