package tui

import (
	"image/color"

	"charm.land/lipgloss/v2"

	"github.com/parham-alvani/arvanstatus-cli/internal/cachet"
)

var (
	colOperational = lipgloss.Color("#22c55e")
	colMinor       = lipgloss.Color("#eab308")
	colMajor       = lipgloss.Color("#ef4444")
	colIdentified  = lipgloss.Color("#f97316")
	colMonitoring  = lipgloss.Color("#eab308")
	colMuted       = lipgloss.Color("#71717a")
	colText        = lipgloss.Color("#e4e4e7")
	colAccent      = lipgloss.Color("#d946ef")
	colBorderIdle  = lipgloss.Color("#3f3f46")
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(colAccent).
			Bold(true).
			Padding(0, 1)

	mutedStyle = lipgloss.NewStyle().Foreground(colMuted)

	textStyle = lipgloss.NewStyle().Foreground(colText)

	selectedStyle = lipgloss.NewStyle().
			Foreground(colAccent).
			Bold(true)

	errStyle = lipgloss.NewStyle().
			Foreground(colMajor).
			Bold(true)

	paneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colBorderIdle)

	paneFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colAccent)

	detailStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colAccent).
			Padding(0, 1)
)

// componentStatusColor maps a Cachet component status value to its bullet
// color: 1 Operational → green … 4 Major Outage → red.
func componentStatusColor(value int) color.Color {
	switch value {
	case cachet.CompOperational:
		return colOperational
	case cachet.CompPerformance:
		return colMinor
	case cachet.CompPartial:
		return colIdentified
	case cachet.CompMajor:
		return colMajor
	}
	return colMuted
}

// incidentStatusColor maps a Cachet incident status value to an accent color:
// 1 Investigating → red, 2 Identified → orange, 3 Watching → yellow,
// 4 Fixed → green.
func incidentStatusColor(value int) color.Color {
	switch value {
	case cachet.StatusInvestigating:
		return colMajor
	case cachet.StatusIdentified:
		return colIdentified
	case cachet.StatusWatching:
		return colMonitoring
	case cachet.StatusFixed:
		return colOperational
	}
	return colMuted
}
