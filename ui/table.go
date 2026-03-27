package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const dataRowOffset = 2 // row 0 = header, row 1 = separator, rows 2+ = data

// ResourceTable is a tview Table that displays Kubernetes resources.
type ResourceTable struct {
	*tview.Table
	columns  []string
	items    []unstructured.Unstructured
	rowToIdx []int // table row → items index; -1 for header/separator rows
}

// NewResourceTable creates a ResourceTable with the given column headers.
func NewResourceTable(columns []string) *ResourceTable {
	t := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetFixed(dataRowOffset, 0).
		SetSelectedStyle(tcell.StyleDefault) // we do multi-row highlighting manually

	rt := &ResourceTable{Table: t, columns: columns}
	rt.setHeaders()
	return rt
}

func (rt *ResourceTable) setHeaders() {
	headerStyle := tcell.StyleDefault.
		Foreground(tview.Styles.SecondaryTextColor).
		Bold(true)

	for i, col := range rt.columns {
		rt.SetCell(0, i, tview.NewTableCell(col).SetStyle(headerStyle).SetSelectable(false))
		rt.SetCell(1, i, tview.NewTableCell("").SetSelectable(false))
	}
}

// SetColumns updates the column definitions and clears all rows.
func (rt *ResourceTable) SetColumns(columns []string) {
	rt.columns = columns
	rt.items = nil
	rt.rowToIdx = nil
	rt.Clear()
	rt.setHeaders()
}

// Populate replaces all data rows. Items with multi-value columns span multiple rows.
func (rt *ResourceTable) Populate(items []unstructured.Unstructured) {
	rt.Clear()
	rt.setHeaders()
	rt.items = items

	// Pre-fill rowToIdx for header + separator.
	rt.rowToIdx = make([]int, dataRowOffset)
	for i := range rt.rowToIdx {
		rt.rowToIdx[i] = -1
	}

	currentRow := dataRowOffset
	for itemIdx, item := range items {
		// Compute lines per column; multi-value columns return >1 line.
		colLines := make([][]string, len(rt.columns))
		maxLines := 1
		for c, col := range rt.columns {
			lines := cellLines(item, col)
			colLines[c] = lines
			if len(lines) > maxLines {
				maxLines = len(lines)
			}
		}

		for r := 0; r < maxLines; r++ {
			rt.rowToIdx = append(rt.rowToIdx, itemIdx)
			for c := range rt.columns {
				val := ""
				if r < len(colLines[c]) {
					val = colLines[c][r]
				}
				rt.SetCell(currentRow, c, tview.NewTableCell(val))
			}
			currentRow++
		}
	}
}

// SelectedIndex returns the 0-based items index for the selected row, or -1.
func (rt *ResourceTable) SelectedIndex() int {
	row, _ := rt.GetSelection()
	if row < 0 || row >= len(rt.rowToIdx) {
		return -1
	}
	return rt.rowToIdx[row]
}

// SelectedName returns the resource name for the selected row, or "".
func (rt *ResourceTable) SelectedName() string {
	idx := rt.SelectedIndex()
	if idx < 0 || idx >= len(rt.items) {
		return ""
	}
	return rt.items[idx].GetName()
}

// firstRowOfItem returns the first table row that belongs to items[idx], or -1.
func (rt *ResourceTable) firstRowOfItem(idx int) int {
	for row, i := range rt.rowToIdx {
		if i == idx {
			return row
		}
	}
	return -1
}

// updateHighlight repaints all data-row cells: selectedIdx's rows get selection
// style, all others revert to default.
func (rt *ResourceTable) updateHighlight(selectedIdx int) {
	sel := selectionStyle()
	nRows := rt.GetRowCount()
	nCols := rt.GetColumnCount()
	for row := dataRowOffset; row < nRows; row++ {
		if row >= len(rt.rowToIdx) {
			break
		}
		var style tcell.Style
		if selectedIdx >= 0 && rt.rowToIdx[row] == selectedIdx {
			style = sel
		} else {
			style = tcell.StyleDefault
		}
		for col := 0; col < nCols; col++ {
			if cell := rt.GetCell(row, col); cell != nil {
				cell.SetStyle(style)
			}
		}
	}
}

// SetItemChangedFunc registers a callback called with the item index whenever the
// selected item changes. It also drives multi-row highlighting.
func (rt *ResourceTable) SetItemChangedFunc(f func(idx int)) {
	rt.Table.SetSelectionChangedFunc(func(row, _ int) {
		idx := -1
		if row >= 0 && row < len(rt.rowToIdx) {
			idx = rt.rowToIdx[row]
		}
		rt.updateHighlight(idx)
		if f != nil {
			f(idx)
		}
	})
}

// cellLines returns one string per display row for a column.
// Multi-value columns (Listeners, Hostnames) return >1 entry.
func cellLines(obj unstructured.Unstructured, col string) []string {
	switch col {
	case "Listeners":
		return listenersLines(obj)
	case "Hostnames":
		hostnames, _, _ := unstructured.NestedStringSlice(obj.Object, "status", "hostnames")
		if len(hostnames) == 0 {
			return []string{""}
		}
        for i := range hostnames {
            hostnames[i] = " - " + hostnames[i]
        }
		return hostnames
	default:
		return []string{cellValue(obj, col)}
	}
}

func cellValue(obj unstructured.Unstructured, col string) string {
	switch col {
	case "Name":
		return obj.GetName()
	case "Namespace":
		return obj.GetNamespace()
	case "Age":
		return age(obj.GetCreationTimestamp().Time)
	case "Class":
		v, _, _ := unstructured.NestedString(obj.Object, "spec", "connectorClassName")
		return v
	case "Accepted":
		return conditionStatus(obj, "Accepted")
	case "Programmed":
		return conditionStatus(obj, "Programmed")
	case "Cert Ready":
		return conditionStatus(obj, "CertificatesReady")
	case "DNS":
		return conditionStatus(obj, "HostnamesVerified")
	case "Ready":
		return conditionStatus(obj, "Ready")
	case "Hostname":
		ann := obj.GetAnnotations()
		if h := ann["dns.datumapis.com/hostname"]; h != "" {
			return h
		}
		return ann["dns.networking.miloapis.com/display-name"]
	case "Type":
		v, _, _ := unstructured.NestedString(obj.Object, "spec", "recordType")
		return v
	default:
		return ""
	}
}

// listenersLines returns one line per unique hostname: "<abbr_host>:<port1>,<port2>".
func listenersLines(obj unstructured.Unstructured) []string {
	listeners, _, _ := unstructured.NestedSlice(obj.Object, "spec", "listeners")

	type group struct{ ports []string }
	var order []string
	groups := map[string]*group{}

	for _, l := range listeners {
		m, ok := l.(map[string]interface{})
		if !ok {
			continue
		}
		hostname, _, _ := unstructured.NestedString(m, "hostname")
		portRaw, exists, _ := unstructured.NestedFieldNoCopy(m, "port")
		if !exists {
			continue
		}
		portStr := fmt.Sprintf("%v", portRaw)
		if f, ok := portRaw.(float64); ok {
			portStr = fmt.Sprintf("%d", int(f))
		}
		if _, seen := groups[hostname]; !seen {
			order = append(order, hostname)
			groups[hostname] = &group{}
		}
		groups[hostname].ports = append(groups[hostname].ports, portStr)
	}

	lines := make([]string, 0, len(order))
	for _, h := range order {
		lines = append(lines, " - "+h+":"+strings.Join(groups[h].ports, ","))
	}
	if len(lines) == 0 {
		return []string{""}
	}
	return lines
}


func conditionStatus(obj unstructured.Unstructured, condType string) string {
	conditions, _, _ := unstructured.NestedSlice(obj.Object, "status", "conditions")
	for _, c := range conditions {
		m, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		if t, _, _ := unstructured.NestedString(m, "type"); t == condType {
			s, _, _ := unstructured.NestedString(m, "status")
			return s
		}
	}
	return "-"
}

func age(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}
