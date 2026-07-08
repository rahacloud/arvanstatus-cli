package tui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/parham-alvani/arvanstatus-cli/internal/cachet"
)

func (m Model) renderIncidentsPane(width, height int) string {
	border := m.paneBorder(m.focused == paneIncidents)
	innerW := width - 2
	innerH := height - 2
	if innerW < 4 || innerH < 2 {
		return border.Width(width - 2).Height(height - 2).Render("")
	}

	title := titleStyle.Render(m.incidentsTitle())
	body := m.renderIncidentsBody(innerW, innerH-1)

	content := lipgloss.JoinVertical(lipgloss.Left, title, body)
	return border.Width(innerW).Height(innerH).Render(content)
}

func (m Model) incidentsTitle() string {
	total := len(m.incidents)
	t := fmt.Sprintf("Incidents (%d shown)", total)
	return t
}

func (m Model) renderIncidentsBody(width, height int) string {
	if len(m.incidents) == 0 {
		if m.pendingIncidents {
			return mutedStyle.Render("  " + m.spinner.View() + " loading incidents…")
		}
		return mutedStyle.Render("  no incidents")
	}

	cursor := m.incidentsCursor
	start := 0
	if cursor >= height {
		start = cursor - height + 1
	}
	end := start + height
	if end > len(m.incidents) {
		end = len(m.incidents)
	}

	now := time.Now()
	var lines []string
	for i := start; i < end; i++ {
		line := renderIncidentRow(m.incidents[i], width-2, now)
		if i == cursor && m.focused == paneIncidents {
			line = selectedStyle.Render("▸ ") + line
		} else {
			line = "  " + line
		}
		line = lipgloss.NewStyle().MaxWidth(width).Render(line)
		lines = append(lines, line)
	}
	return padLines(strings.Join(lines, "\n"), height)
}

func renderIncidentRow(inc cachet.Incident, width int, now time.Time) string {
	statusStyle := lipgloss.NewStyle().
		Foreground(incidentStatusColor(inc.Status.Value)).
		Bold(true)

	bullet := statusStyle.Render("●")
	badge := statusStyle.Render(fmt.Sprintf("[%s]", strings.ToUpper(inc.Status.Human)))
	when := mutedStyle.Render(relTime(inc.UpdatedAt.Time, now))

	head := fmt.Sprintf("%s %s %s", bullet, badge, textStyle.Render(inc.Title))
	tail := fmt.Sprintf("  %s", when)

	avail := width - lipgloss.Width(tail)
	if avail < 10 {
		return head
	}
	head = lipgloss.NewStyle().MaxWidth(avail).Render(head)
	return head + tail
}

func relTime(t, now time.Time) string {
	if t.IsZero() {
		return ""
	}
	d := now.Sub(t)
	switch {
	case d < 0:
		return "just now"
	case d < time.Minute:
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	default:
		return t.Format("2006-01-02")
	}
}
