// Package ui provides terminal user interface utilities
package ui

import (
	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
)

// Colors defines the color scheme for the application
type Colors struct {
	Header    *color.Color
	Success   *color.Color
	Error     *color.Color
	Warning   *color.Color
	Info      *color.Color
	Highlight *color.Color
	Text      *color.Color
}

// NewColors creates a new color scheme
func NewColors() *Colors {
	return &Colors{
		Header:    color.New(color.FgCyan, color.Bold),
		Success:   color.New(color.FgGreen),
		Error:     color.New(color.FgRed),
		Warning:   color.New(color.FgYellow),
		Info:      color.New(color.FgBlue),
		Highlight: color.New(color.FgMagenta, color.Bold),
		Text:      color.New(color.FgWhite),
	}
}

// TableConfig provides configuration for table styling
type TableConfig struct {
	Style table.Style
}

// NewTableConfig creates a new table configuration
func NewTableConfig() *TableConfig {
	return &TableConfig{
		Style: table.StyleColoredBright,
	}
}

// CreateTable creates a new table with the default styling
func CreateTable() table.Writer {
	t := table.NewWriter()
	t.SetStyle(table.StyleColoredBright)
	return t
}
