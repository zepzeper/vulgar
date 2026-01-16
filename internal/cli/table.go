package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/zepzeper/vulgar/internal/config"
)

type Column struct {
	Title string
	Width int
}

type Table struct {
	columns []Column
	rows    [][]string
	title   string
}

func NewTable() *Table {
	return &Table{}
}

func (t *Table) Title(title string) *Table {
	t.title = title
	return t
}

func (t *Table) Columns(cols ...Column) *Table {
	t.columns = cols
	return t
}

func (t *Table) Rows(rows [][]string) *Table {
	t.rows = rows
	return t
}

func (t *Table) AddRow(row ...string) *Table {
	t.rows = append(t.rows, row)
	return t
}

func (t *Table) Render() {
	cfg := config.Get()

	if cfg.Defaults.OutputFormat == "json" {
		t.renderJSON()
		return
	}

	if t.title != "" {
		fmt.Println()
		fmt.Println(StyleTitle.Render(t.title))
		fmt.Println(StyleMuted.Render(strings.Repeat("-", len(t.title)+2)))
	}

	if len(t.rows) == 0 {
		PrintWarning("No data to display")
		return
	}

	tableCols := make([]table.Column, len(t.columns))
	for i, col := range t.columns {
		tableCols[i] = table.Column{
			Title: col.Title,
			Width: col.Width,
		}
	}

	tableRows := make([]table.Row, len(t.rows))
	for i, row := range t.rows {
		tableRows[i] = table.Row(row)
	}

	tbl := table.New(
		table.WithColumns(tableCols),
		table.WithRows(tableRows),
		table.WithFocused(false),
		table.WithHeight(len(t.rows)+1),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(ColorPrimary).
		BorderBottom(true).
		Bold(true).
		Foreground(ColorPrimary)
	s.Selected = s.Selected.
		Foreground(lipgloss.NoColor{}).
		Background(lipgloss.NoColor{}).
		Bold(false)
	tbl.SetStyles(s)

	fmt.Println(tbl.View())
}

func (t *Table) renderJSON() {
	var result []map[string]string
	for _, row := range t.rows {
		item := make(map[string]string)
		for i, col := range t.columns {
			if i < len(row) {
				item[col.Title] = row[i]
			}
		}
		result = append(result, item)
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(data))
}

func PrintTable(columns []Column, rows [][]string) {
	NewTable().Columns(columns...).Rows(rows).Render()
}

func Truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

func PadRight(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}

func PadLeft(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return strings.Repeat(" ", width-len(s)) + s
}
