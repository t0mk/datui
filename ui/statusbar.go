package ui

import (
	"fmt"

	"github.com/rivo/tview"
)

// StatusBar is a single-line bar at the bottom of the screen.
type StatusBar struct {
	*tview.TextView
}

// NewStatusBar creates a new StatusBar.
func NewStatusBar() *StatusBar {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	tv.SetBorder(false)
	return &StatusBar{tv}
}

// Update refreshes the status bar content.
func (s *StatusBar) Update(context, project, namespace, resource string) {
	proj := project
	if proj == "" {
		proj = "-"
	}
	s.SetText(fmt.Sprintf(
		" [blue]CTX:[-] %s   [blue]PRJ:[-] %s   [blue]NS:[-] %s   [blue]RES:[-] %s"+
			"   [darkgray]<:> switch  <Enter> describe  <d> delete  <q> quit[-]",
		context, proj, namespace, resource,
	))
}
