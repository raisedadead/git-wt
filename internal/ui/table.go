package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

var (
	defaultCellStyle   = lipgloss.NewStyle().Padding(0, 1)
	defaultHeaderStyle = defaultCellStyle.Bold(true)
	borderStyle        = lipgloss.NewStyle().Foreground(Subtle)
)

func NewTable() *table.Table {
	return table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(borderStyle).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return defaultHeaderStyle
			}
			return defaultCellStyle
		})
}

func NewStyledTable(fn func(row, col int) lipgloss.Style) *table.Table {
	return table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(borderStyle).
		StyleFunc(fn)
}
