package cachet

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// fixtures mirror the real Cachet v3 JSON:API shapes, including the group ID /
// component ID overlap that the real site exhibits (group id "3" vs component
// id "3").
const (
	groupsJSON = `{"data":[
		{"id":"3","type":"componentGroups","attributes":{"id":3,"name":"Bamdad (ir-central1-a)","order":2}},
		{"id":"1","type":"componentGroups","attributes":{"id":1,"name":"Global","order":1}}
	]}`

	componentsJSON = `{"data":[
		{"id":"3","type":"components","attributes":{"id":3,"name":"Bamdad - Object Storage","order":5,"status":{"human":"Operational","value":1}},
		 "relationships":{"group":{"data":{"type":"componentGroups","id":"3"}}}},
		{"id":"9","type":"components","attributes":{"id":9,"name":"Bamdad - Cloud Servers","order":1,"status":{"human":"Major Outage","value":4}},
		 "relationships":{"group":{"data":{"type":"componentGroups","id":"3"}}}},
		{"id":"20","type":"components","attributes":{"id":20,"name":"Website","order":1,"status":{"human":"Operational","value":1}},
		 "relationships":{"group":{"data":{"type":"componentGroups","id":"1"}}}},
		{"id":"99","type":"components","attributes":{"id":99,"name":"Orphan Service","order":1,"status":{"human":"Operational","value":1}},
		 "relationships":{"group":{"data":null}}}
	]}`

	incidentsJSON = `{"data":[
		{"id":"56","type":"incidents",
		 "attributes":{"id":56,"name":"Object Storage Connectivity Issue","message":"root message",
			"status":{"human":"Fixed","value":4},
			"occurred":{"string":"2026-07-08 18:15:00"},"updated":{"string":"2026-07-08 19:14:22"}},
		 "relationships":{"components":{"data":[{"type":"components","id":"3"}]},
			"updates":{"data":[{"type":"updates","id":"150"},{"type":"updates","id":"151"}]}}},
		{"id":"40","type":"incidents",
		 "attributes":{"id":40,"name":"No updates incident","message":"fallback body",
			"status":{"human":"Investigating","value":1},
			"occurred":{"string":"2026-06-01 10:00:00"},"updated":{"string":"2026-06-01 10:00:00"}},
		 "relationships":{"components":{"data":[]},"updates":{"data":[]}}}
	],
	"included":[
		{"id":"150","type":"updates","attributes":{"id":150,"message":"older update","status":{"human":"Identified","value":2},"created":{"string":"2026-07-08 18:30:00"}}},
		{"id":"151","type":"updates","attributes":{"id":151,"message":"newest update","status":{"human":"Fixed","value":4},"created":{"string":"2026-07-08 19:00:47"}}}
	]}`
)

func testServer(t *testing.T) *Client {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/component-groups", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(groupsJSON))
	})
	mux.HandleFunc("/api/components", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(componentsJSON))
	})
	mux.HandleFunc("/api/incidents", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(incidentsJSON))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return NewClient(WithBaseURL(srv.URL))
}

func TestFetchSummaryTree(t *testing.T) {
	s, err := testServer(t).FetchSummary(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// Groups are ordered by their `order` field: Global(1) before Bamdad(2).
	if len(s.Services) != 3 {
		t.Fatalf("want 3 top-level groups (incl. Other), got %d: %+v", len(s.Services), s.Services)
	}
	if s.Services[0].Name != "Global" || s.Services[1].Name != "Bamdad (ir-central1-a)" {
		t.Errorf("group order wrong: %q, %q", s.Services[0].Name, s.Services[1].Name)
	}

	// Bamdad has a Major Outage child, so its aggregate must be Major.
	bamdad := s.Services[1]
	if bamdad.Status.Value != CompMajor {
		t.Errorf("Bamdad group status = %d, want %d (worst-of-children)", bamdad.Status.Value, CompMajor)
	}

	// The orphan (null group) is collected under a synthetic "Other" group.
	if last := s.Services[2]; last.Name != "Other" || len(last.Children) != 1 {
		t.Errorf("orphan handling wrong: %+v", last)
	}
}

func TestServicesByIDNoCollision(t *testing.T) {
	s, err := testServer(t).FetchSummary(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	byID := s.ServicesByID()

	// Group id 3 and component id 3 overlap. The map must resolve id 3 to the
	// component (what incidents reference), not the group.
	got, ok := byID[3]
	if !ok {
		t.Fatal("id 3 missing from ServicesByID")
	}
	if got.Name != "Bamdad - Object Storage" {
		t.Errorf("id 3 resolved to %q, want the component 'Bamdad - Object Storage'", got.Name)
	}
	// Only leaf components should be present (4), not groups.
	if len(byID) != 4 {
		t.Errorf("ServicesByID has %d entries, want 4 leaf components", len(byID))
	}
}

func TestFetchIncidents(t *testing.T) {
	p, err := testServer(t).FetchIncidents(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Incidents) != 2 {
		t.Fatalf("want 2 incidents, got %d", len(p.Incidents))
	}

	first := p.Incidents[0]
	if first.Status.Human != "Fixed" {
		t.Errorf("status = %q", first.Status.Human)
	}
	// Updates must be sorted newest-first, so LatestUpdate is the 19:00 one.
	if got := first.LatestUpdate(); got != "newest update" {
		t.Errorf("LatestUpdate = %q, want newest update", got)
	}
	if len(first.ComponentIDs) != 1 || first.ComponentIDs[0] != 3 {
		t.Errorf("ComponentIDs = %v, want [3]", first.ComponentIDs)
	}
	// Tehran offset +03:30 must be applied to the naive timestamp.
	if _, off := first.Occurred.Zone(); off != 3*3600+30*60 {
		t.Errorf("occurred offset = %ds, want Asia/Tehran +12600s", off)
	}

	// Incident with no updates falls back to its own message.
	if got := p.Incidents[1].LatestUpdate(); got != "fallback body" {
		t.Errorf("fallback LatestUpdate = %q", got)
	}
}
