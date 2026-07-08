package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/parham-alvani/arvanstatus-cli/internal/cachet"
)

func (m Model) renderServicesPane(width, height int) string {
	border := m.paneBorder(m.focused == paneServices)
	innerW := width - 2 // border
	innerH := height - 2
	if innerW < 4 || innerH < 2 {
		return border.Width(width - 2).Height(height - 2).Render("")
	}

	title := titleStyle.Render(m.servicesTitle())
	body := m.renderServicesBody(innerW, innerH-1)

	content := lipgloss.JoinVertical(lipgloss.Left, title, body)
	return border.Width(innerW).Height(innerH).Render(content)
}

func (m Model) servicesTitle() string {
	name := cachet.PageName
	if m.summary != nil && m.summary.PageName != "" {
		name = m.summary.PageName
	}
	badge := "all systems operational"
	badgeStyle := lipgloss.NewStyle().Foreground(colOperational)
	if m.summary != nil {
		if deg := countDegraded(m.summary.Services); deg > 0 {
			badge = fmt.Sprintf("%d service(s) affected", deg)
			badgeStyle = lipgloss.NewStyle().Foreground(colMajor)
		}
	}
	return fmt.Sprintf("%s  %s", name, badgeStyle.Render(badge))
}

// countDegraded counts leaf components whose status is worse than Operational.
// Only leaves are counted so a degraded group doesn't double-count with its
// child.
func countDegraded(xs []cachet.Service) int {
	n := 0
	for _, x := range xs {
		if len(x.Children) == 0 {
			if x.Status.Value > cachet.CompOperational {
				n++
			}
		} else {
			n += countDegraded(x.Children)
		}
	}
	return n
}

func (m Model) renderServicesBody(width, height int) string {
	if len(m.flatServices) == 0 {
		if m.pendingSummary {
			return mutedStyle.Render("  " + m.spinner.View() + " loading services…")
		}
		return mutedStyle.Render("  no data")
	}

	cursor := m.servicesCursor
	start := 0
	if cursor >= height {
		start = cursor - height + 1
	}
	end := start + height
	if end > len(m.flatServices) {
		end = len(m.flatServices)
	}

	var lines []string
	for i := start; i < end; i++ {
		fs := m.flatServices[i]
		indent := strings.Repeat("  ", fs.depth)
		bullet := lipgloss.NewStyle().
			Foreground(componentStatusColor(fs.svc.Status.Value)).
			Render("●")

		name := fs.svc.Name
		label := fmt.Sprintf("%s%s %s", indent, bullet, name)

		if i == cursor && m.focused == paneServices {
			label = selectedStyle.Render("▸ ") + label
		} else {
			label = "  " + label
		}

		label = lipgloss.NewStyle().MaxWidth(width).Render(label)
		lines = append(lines, label)
	}
	return padLines(strings.Join(lines, "\n"), height)
}
