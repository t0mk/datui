package ui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

// DetailView is a borderless flex showing a header line + YAML content + optional search bar.
type DetailView struct {
	*tview.Flex
	app        *tview.Application
	header     *tview.TextView
	content    *tview.TextView
	searchBar  *tview.InputField
	searchFlex *tview.Flex // inner flex holding content + searchBar

	plainLines  []string // raw (un-highlighted) YAML lines
	matchLines  []int    // line indices of current search matches
	matchIdx    int      // index into matchLines for current match
	searchOpen  bool
}

// NewDetailView creates a scrollable YAML detail view with no border.
func NewDetailView(app *tview.Application) *DetailView {
	header := tview.NewTextView().SetDynamicColors(true)

	rule := tview.NewTextView().
		SetDynamicColors(true).
		SetText("[darkgray]" + strings.Repeat("─", 200) + "[-]")

	content := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(false)

	searchBar := tview.NewInputField().
		SetLabel("/").
		SetLabelColor(tcell.ColorNavy).
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetFieldTextColor(tcell.ColorBlack)

	searchFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(content, 0, 1, true).
		AddItem(searchBar, 0, 0, false) // height 0 = hidden

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 1, 0, false).
		AddItem(rule, 1, 0, false).
		AddItem(searchFlex, 0, 1, true)

	d := &DetailView{
		Flex:       flex,
		app:        app,
		header:     header,
		content:    content,
		searchBar:  searchBar,
		searchFlex: searchFlex,
	}

	searchBar.SetChangedFunc(func(text string) {
		d.search(text)
	})
	searchBar.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			d.nextMatch()
		case tcell.KeyEscape:
			d.closeSearch()
			app.SetFocus(content)
		}
	})

	return d
}

// Show renders the resource as YAML with syntax highlighting.
func (d *DetailView) Show(obj *unstructured.Unstructured, resource string) {
	d.header.SetText(fmt.Sprintf("[navy::b]%s  %s[-:-:-]   [darkgray](Esc: back  /: search)[-]",
		resource, obj.GetName()))

	clean := obj.DeepCopy()
	clean.SetManagedFields(nil)

	b, err := yaml.Marshal(clean.Object)
	if err != nil {
		d.content.SetText("error marshalling YAML: " + err.Error())
		return
	}

	d.plainLines = strings.Split(string(b), "\n")
	d.matchLines = nil
	d.matchIdx = 0
	d.closeSearch()
	d.renderContent("")
	d.content.ScrollToBeginning()
}

// renderContent sets the content text. If term is non-empty, matching lines get extra highlight.
func (d *DetailView) renderContent(term string) {
	termLower := strings.ToLower(term)
	out := make([]string, len(d.plainLines))
	for i, line := range d.plainLines {
		if term != "" && strings.Contains(strings.ToLower(line), termLower) {
			out[i] = "[::r]" + tview.Escape(line) + "[-:-:-]"
		} else {
			out[i] = highlightYAMLLine(line)
		}
	}
	d.content.SetText(strings.Join(out, "\n"))
}

func (d *DetailView) openSearch() {
	if d.searchOpen {
		return
	}
	d.searchOpen = true
	d.searchFlex.ResizeItem(d.searchBar, 1, 0)
	d.searchBar.SetText("")
	d.matchLines = nil
	d.matchIdx = 0
	d.app.SetFocus(d.searchBar)
}

func (d *DetailView) closeSearch() {
	d.searchOpen = false
	d.searchFlex.ResizeItem(d.searchBar, 0, 0)
	d.matchLines = nil
	d.matchIdx = 0
	d.renderContent("")
}

func (d *DetailView) search(term string) {
	d.matchLines = nil
	d.matchIdx = 0
	if term == "" {
		d.renderContent("")
		return
	}
	termLower := strings.ToLower(term)
	for i, line := range d.plainLines {
		if strings.Contains(strings.ToLower(line), termLower) {
			d.matchLines = append(d.matchLines, i)
		}
	}
	d.renderContent(term)
	if len(d.matchLines) > 0 {
		d.scrollToLine(d.matchLines[0])
	}
}

func (d *DetailView) nextMatch() {
	if len(d.matchLines) == 0 {
		return
	}
	d.matchIdx = (d.matchIdx + 1) % len(d.matchLines)
	d.scrollToLine(d.matchLines[d.matchIdx])
}

func (d *DetailView) prevMatch() {
	if len(d.matchLines) == 0 {
		return
	}
	d.matchIdx = (d.matchIdx - 1 + len(d.matchLines)) % len(d.matchLines)
	d.scrollToLine(d.matchLines[d.matchIdx])
}

func (d *DetailView) scrollToLine(line int) {
	d.content.ScrollTo(line, 0)
}
