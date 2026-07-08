package main

import (
	"flag"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/parham-alvani/arvanstatus-cli/internal/cachet"
	"github.com/parham-alvani/arvanstatus-cli/internal/tui"
)

// Populated at build time via -ldflags by goreleaser.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	showVersion := flag.Bool("version", false, "Show version information and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("arvanstatus-cli %s (commit %s, built %s)\n", version, commit, date)
		return
	}

	client := cachet.NewClient()
	p := tea.NewProgram(tui.New(client))
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "arvanstatus-cli:", err)
		os.Exit(1)
	}
}
