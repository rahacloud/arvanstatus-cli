package tui

import (
	"time"

	"github.com/parham-alvani/arvanstatus-cli/internal/cachet"
)

type tickMsg time.Time

type refreshStartedMsg struct{}

type summaryFetchedMsg struct {
	s   *cachet.Summary
	err error
}

type incidentsFetchedMsg struct {
	p   *cachet.IncidentsPage
	err error
}
