package tui

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/parham-alvani/arvanstatus-cli/internal/cachet"
)

func (m Model) renderDetail() string {
	if m.width <= 4 || m.height <= 4 {
		return ""
	}
	box := detailStyle.
		Width(m.width - 2).
		Height(m.height - 2)
	return box.Render(m.detail.View())
}

// renderDetailBody is called by the Update handler before populating the
// viewport. Pure function of inputs — no lipgloss state mutation.
func renderDetailBody(inc cachet.Incident, byID map[int]cachet.Service, width int) string {
	if width <= 4 {
		width = 60
	}

	statusStyle := lipgloss.NewStyle().
		Foreground(incidentStatusColor(inc.Status.Value)).
		Bold(true)

	header := titleStyle.Render(inc.Title)
	badges := statusStyle.Render(fmt.Sprintf("[%s]", strings.ToUpper(inc.Status.Human)))
	occurred := mutedStyle.Render("Occurred: ") + fmtTime(inc.Occurred)
	updated := mutedStyle.Render("Updated:  ") + fmtTime(inc.UpdatedAt)

	names := resolveServiceNames(inc.ComponentIDs, byID)
	affected := mutedStyle.Render("Affected services:")
	if len(names) == 0 {
		affected += " " + mutedStyle.Render("(none listed)")
	} else {
		affected += "\n  " + strings.Join(names, ", ")
	}

	latest := mutedStyle.Render("Latest update:")
	body := wrap(inc.LatestUpdate(), width-2)

	return strings.Join([]string{
		header,
		badges,
		"",
		occurred,
		updated,
		"",
		affected,
		"",
		latest,
		body,
	}, "\n")
}

// fmtTime renders a LocalTime, or a dash when it is unset.
func fmtTime(t cachet.LocalTime) string {
	if t.IsZero() {
		return "—"
	}
	return t.Format("2006-01-02 15:04 MST")
}

func resolveServiceNames(ids []int, byID map[int]cachet.Service) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if s, ok := byID[id]; ok {
			out = append(out, s.Name)
		}
	}
	sort.Strings(out)
	return out
}

// wrap is a tiny word-wrap helper (no extra dep). Not perfect for CJK but
// adequate for English status-page text.
func wrap(s string, width int) string {
	if width <= 0 {
		return s
	}
	var out strings.Builder
	for _, para := range strings.Split(s, "\n") {
		line := 0
		for _, word := range strings.Fields(para) {
			wl := len(word)
			if line > 0 && line+1+wl > width {
				out.WriteByte('\n')
				line = 0
			}
			if line > 0 {
				out.WriteByte(' ')
				line++
			}
			out.WriteString(word)
			line += wl
		}
		out.WriteByte('\n')
	}
	return strings.TrimRight(out.String(), "\n")
}
