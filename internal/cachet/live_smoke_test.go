package cachet

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestLiveSmoke hits the real ArvanCloud status page. It is opt-in so CI (which
// runs the full suite without -short) stays hermetic; run it with
// `ARVAN_LIVE=1 go test -run TestLiveSmoke -v ./internal/cachet`.
func TestLiveSmoke(t *testing.T) {
	if os.Getenv("ARVAN_LIVE") == "" {
		t.Skip("set ARVAN_LIVE=1 to run the live network smoke test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	c := NewClient()
	s, err := c.FetchSummary(ctx)
	if err != nil {
		t.Fatalf("FetchSummary: %v", err)
	}
	byID := s.ServicesByID()
	found := false
	for _, svc := range byID {
		if svc.Name == "Bamdad - Object Storage" {
			found = true
		}
	}
	t.Logf("groups=%d totalNodes=%d bamdadObjectStorage=%v", len(s.Services), len(byID), found)
	for _, g := range s.Services {
		t.Logf("  group %q status=%s (%d children)", g.Name, g.Status.Human, len(g.Children))
	}
	if !found {
		t.Errorf("expected 'Bamdad - Object Storage' in component tree")
	}

	p, err := c.FetchIncidents(ctx)
	if err != nil {
		t.Fatalf("FetchIncidents: %v", err)
	}
	t.Logf("incidents=%d", len(p.Incidents))
	if len(p.Incidents) > 0 {
		i := p.Incidents[0]
		t.Logf("  latest: %q status=%s occurred=%s updates=%d latestUpdate=%.60q",
			i.Title, i.Status.Human, i.Occurred.Format("2006-01-02 15:04 MST"), len(i.Updates), i.LatestUpdate())
	}
}
