package cachet

import (
	"context"
	"encoding/json"
	"net/url"
	"sort"
)

// Incident status values (Cachet). Higher means further toward resolution.
const (
	StatusInvestigating = 1
	StatusIdentified    = 2
	StatusWatching      = 3
	StatusFixed         = 4
)

// Incident is one row of the /api/incidents list.
type Incident struct {
	ID           int
	Title        string
	Message      string
	Status       Status
	Occurred     LocalTime
	UpdatedAt    LocalTime
	ComponentIDs []int
	Updates      []IncidentUpdate
}

// IncidentUpdate is one entry in an incident's timeline, newest-relevant first
// after assembly.
type IncidentUpdate struct {
	Message   string
	Status    Status
	CreatedAt LocalTime
}

// LatestUpdate returns the most recent update's message, falling back to the
// incident's own message when there are no updates.
func (i Incident) LatestUpdate() string {
	if len(i.Updates) > 0 {
		return i.Updates[0].Message
	}
	return i.Message
}

// IncidentsPage is the assembled incident list, newest first.
type IncidentsPage struct {
	Incidents []Incident
}

type incidentAttr struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Message  string `json:"message"`
	Status   Status `json:"status"`
	Occurred stamp  `json:"occurred"`
	Updated  stamp  `json:"updated"`
}

type updateAttr struct {
	ID      int    `json:"id"`
	Message string `json:"message"`
	Status  Status `json:"status"`
	Created stamp  `json:"created"`
}

// FetchIncidents retrieves recent incidents with their updates and affected
// components resolved.
func (c *Client) FetchIncidents(ctx context.Context) (*IncidentsPage, error) {
	var doc document
	q := url.Values{"include": {"updates,components"}, "per_page": {"50"}}
	if err := c.getJSON(ctx, "incidents", q, &doc); err != nil {
		return nil, err
	}

	// Index side-loaded updates by ID so each incident can resolve its own.
	updates := map[string]IncidentUpdate{}
	for _, r := range doc.Included {
		if r.Type != "updates" {
			continue
		}
		var a updateAttr
		if err := json.Unmarshal(r.Attributes, &a); err != nil {
			continue
		}
		updates[r.ID] = IncidentUpdate{
			Message:   a.Message,
			Status:    a.Status,
			CreatedAt: parseStamp(a.Created),
		}
	}

	var rs []resource
	if err := json.Unmarshal(doc.Data, &rs); err != nil {
		return nil, err
	}

	incidents := make([]Incident, 0, len(rs))
	for _, r := range rs {
		var a incidentAttr
		if err := json.Unmarshal(r.Attributes, &a); err != nil {
			return nil, err
		}
		inc := Incident{
			ID:        a.ID,
			Title:     a.Name,
			Message:   a.Message,
			Status:    a.Status,
			Occurred:  parseStamp(a.Occurred),
			UpdatedAt: parseStamp(a.Updated),
		}
		if rel, ok := r.Relationships["components"]; ok {
			inc.ComponentIDs = rel.many()
		}
		if rel, ok := r.Relationships["updates"]; ok {
			var refs []resourceRef
			if err := json.Unmarshal(rel.Data, &refs); err == nil {
				for _, ref := range refs {
					if u, ok := updates[ref.ID]; ok {
						inc.Updates = append(inc.Updates, u)
					}
				}
			}
		}
		// Cachet returns an incident's updates in an unspecified order; sort
		// newest first so LatestUpdate() is meaningful.
		sort.Slice(inc.Updates, func(i, j int) bool {
			return inc.Updates[i].CreatedAt.After(inc.Updates[j].CreatedAt.Time)
		})
		incidents = append(incidents, inc)
	}

	return &IncidentsPage{Incidents: incidents}, nil
}
