package ui

import (
	"context"
	"fmt"
    "strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/t0mk/datui/config"
	"github.com/t0mk/datui/logger"
	"github.com/t0mk/datui/resources"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type tcellEvent = tcell.EventKey

// App is the root TUI application.
type App struct {
	tview       *tview.Application
	pages       *tview.Pages
	table       *ResourceTable
	tableHeader *tview.TextView
	detail      *DetailView
	statusBar   *StatusBar
	switcher    *tview.List

	dc        dynamic.Interface
	cfg       *config.Config
	activeGVR schema.GroupVersionResource
	namespace string
	items     []unstructured.Unstructured
}

// New creates and wires the TUI application.
func New(cfg *config.Config, dc dynamic.Interface) *App {
	tviewApp := tview.NewApplication()
	a := &App{
		tview:       tviewApp,
		pages:       tview.NewPages(),
		detail:      NewDetailView(tviewApp),
		statusBar:   NewStatusBar(),
		tableHeader: tview.NewTextView().SetDynamicColors(true),
		switcher:    buildSwitcherList(),
		cfg:         cfg,
		dc:          dc,
		namespace:   cfg.Namespace,
		activeGVR:   resources.HTTPProxyGVR,
	}
	a.table = NewResourceTable(resources.All[0].Columns)
	a.buildLayout()
	a.bindKeys()
	return a
}

func (a *App) buildLayout() {
	mainLayout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.tableHeader, 1, 0, false).
		AddItem(a.table, 0, 1, true).
		AddItem(a.statusBar, 1, 0, false)

	// Switcher modal: header label + list, no border.
	switcherHeader := tview.NewTextView().
		SetDynamicColors(true).
		SetText("[navy::b]select resource type[-:-:-]")

	listH := len(resources.All) + 1 // +1 for the header line
	switcherModal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(switcherHeader, 1, 0, false).
				AddItem(a.switcher, listH, 0, true), listH+1, 0, true).
			AddItem(nil, 0, 1, false), 32, 0, true).
		AddItem(nil, 0, 1, false)

	a.pages.AddPage("main", mainLayout, true, true)
	a.pages.AddPage("switcher", switcherModal, true, false)
	a.pages.AddPage("detail", a.detail, true, false)

	a.tview.SetRoot(a.pages, true).EnableMouse(false)
	a.tview.SetInputCapture(func(event *tcellEvent) *tcellEvent {
		if event.Key() == tcell.KeyCtrlC {
			a.tview.Stop()
			return nil
		}
		return event
	})
}

func (a *App) setTableHeader(resource string) {
	a.tableHeader.SetText(fmt.Sprintf("[green::b]%s[-:-:-]", strings.ToUpper(resource)))
}

func (a *App) bindKeys() {
	a.table.SetInputCapture(func(event *tcellEvent) *tcellEvent {
		switch event.Key() {
		case tcell.KeyUp:
			if idx := a.table.SelectedIndex(); idx > 0 {
				next := idx - 1
				if row := a.table.firstRowOfItem(next); row >= 0 {
					a.table.Select(row, 0)
					a.table.updateHighlight(next)
				}
			}
			return nil
		case tcell.KeyDown:
			if idx := a.table.SelectedIndex(); idx >= 0 && idx < len(a.items)-1 {
				next := idx + 1
				if row := a.table.firstRowOfItem(next); row >= 0 {
					a.table.Select(row, 0)
					a.table.updateHighlight(next)
				}
			}
			return nil
		}
		switch event.Rune() {
		case ':':
			a.openSwitcher()
			return nil
		case 'q':
			a.tview.Stop()
			return nil
		case 'd':
			a.deleteSelected()
			return nil
		case 'y':
			a.showYAML()
			return nil
		}
		return event
	})

	a.table.SetSelectedFunc(func(_, _ int) {
		if idx := a.table.SelectedIndex(); idx >= 0 {
			a.showDetail(idx)
		}
	})

	a.table.SetItemChangedFunc(func(idx int) {
		a.setTableHeader(a.activeGVR.Resource)
	})

	a.switcher.SetSelectedFunc(func(idx int, _ string, _ string, _ rune) {
		a.switchResourceByIndex(idx)
		a.pages.SwitchToPage("main")
		a.tview.SetFocus(a.table)
	})
	a.switcher.SetInputCapture(func(event *tcellEvent) *tcellEvent {
		if event.Key() == tcell.KeyEscape {
			a.pages.SwitchToPage("main")
			a.tview.SetFocus(a.table)
			return nil
		}
		return event
	})

	a.detail.content.SetInputCapture(func(event *tcellEvent) *tcellEvent {
		if event.Key() == tcell.KeyEscape {
			a.pages.SwitchToPage("main")
			a.tview.SetFocus(a.table)
			return nil
		}
		switch event.Rune() {
		case '/':
			a.detail.openSearch()
			return nil
		case 'n':
			a.detail.nextMatch()
			return nil
		case 'N':
			a.detail.prevMatch()
			return nil
		}
		return event
	})
}

func (a *App) openSwitcher() {
	a.pages.SwitchToPage("switcher")
	a.tview.SetFocus(a.switcher)
}

func (a *App) switchResourceByIndex(idx int) {
	if idx < 0 || idx >= len(resources.All) {
		return
	}
	rd := resources.All[idx]
	a.activeGVR = rd.GVR
	a.table.SetColumns(rd.Columns)
	a.setTableHeader(rd.GVR.Resource)
	go a.refresh()
}

func (a *App) updateStatus(resource string) {
	a.statusBar.Update(a.cfg.ContextName, a.cfg.ProjectName, a.namespace, resource)
}

func (a *App) refresh() {
	list, err := a.dc.Resource(a.activeGVR).Namespace(a.namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		logger.Errorf("list %s: %v", a.activeGVR.Resource, err)
		a.items = nil
		a.tview.QueueUpdateDraw(func() {
			a.table.Populate(nil)
			a.setTableHeader(a.activeGVR.Resource)
			a.statusBar.Update(a.cfg.ContextName, a.cfg.ProjectName, a.namespace,
				fmt.Sprintf("ERROR: %s not found", a.activeGVR.Resource))
		})
		return
	}
	a.items = list.Items
	a.tview.QueueUpdateDraw(func() {
		a.table.Populate(a.items)
		a.table.updateHighlight(0)
		a.setTableHeader(a.activeGVR.Resource)
		a.updateStatus(a.activeGVR.Resource)
	})
}

func (a *App) showDetail(idx int) {
	if idx < 0 || idx >= len(a.items) {
		return
	}
	obj, err := a.dc.Resource(a.activeGVR).Namespace(a.namespace).Get(
		context.Background(), a.items[idx].GetName(), metav1.GetOptions{},
	)
	if err != nil {
		logger.Errorf("get %s/%s: %v", a.activeGVR.Resource, a.items[idx].GetName(), err)
		return
	}
	a.detail.Show(obj, a.activeGVR.Resource)
	a.pages.SwitchToPage("detail")
	a.tview.SetFocus(a.detail.content)
}

func (a *App) showYAML() {
	if idx := a.table.SelectedIndex(); idx >= 0 {
		a.showDetail(idx)
	}
}

func (a *App) deleteSelected() {
	name := a.table.SelectedName()
	if name == "" {
		return
	}
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Delete %s/%s ?", a.activeGVR.Resource, name)).
		AddButtons([]string{"Cancel", "Delete"}).
		SetDoneFunc(func(_ int, label string) {
			if label == "Delete" {
				err := a.dc.Resource(a.activeGVR).Namespace(a.namespace).Delete(
					context.Background(), name, metav1.DeleteOptions{},
				)
				if err != nil {
					logger.Errorf("delete %s/%s: %v", a.activeGVR.Resource, name, err)
				}
				go a.refresh()
			}
			a.pages.RemovePage("confirm")
			a.tview.SetFocus(a.table)
		})
	a.pages.AddPage("confirm", modal, false, true)
	a.tview.SetFocus(modal)
}

// Run starts the initial list fetch and the tview event loop.
func (a *App) Run() error {
	a.setTableHeader(a.activeGVR.Resource)
	go a.refresh()
	return a.tview.Run()
}

// buildSwitcherList creates the borderless resource-picker List.
func buildSwitcherList() *tview.List {
	l := tview.NewList().ShowSecondaryText(false)
	fg, bg := selectionColors()
	l.SetSelectedTextColor(fg).SetSelectedBackgroundColor(bg)
	for _, rd := range resources.All {
		l.AddItem(fmt.Sprintf("%-24s [darkgray]%s[-]", rd.DisplayName, rd.ShortName), "", 0, nil)
	}
	return l
}
